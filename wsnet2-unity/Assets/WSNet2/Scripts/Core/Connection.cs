using System;
using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Threading.Tasks;
using System.Threading;

namespace WSNet2.Core
{
    class Connection
    {
        // todo: 設定を注入できるようにする

        /// <summary>保持できるEventの数</summary>
        const int EvBufPoolSize = 16;

        /// <summary>各Eventのバッファサイズの初期値</summary>
        const int EvBufInitialSize = 256;

        /// <summary>保持できるMsgの数</summary>
        const int MsgPoolSize = 8;

        /// <summary>各Msgのバッファサイズの初期値</summary>
        const int MsgBufInitialSize = 512;

        /// <summary>最大連続再接続試行回数</summary>
        const int MaxReconnection = 5;

        /// <summary>再接続インターバル (milli seconds)</summary>
        const int RetryIntervalMilliSec = 1000;

        /// <summary>最大Ping間隔 (milli seconds)</summary>
        /// Playerの最終Msg時刻のやりとりのため、ある程度で上限を設ける
        const int MaxPingIntervalMilliSec = 10000;

        /// <summary>最小Ping間隔 (milli seconds)</summary>
        const int MinPingIntervalMilliSec = 1000;

        static AuthDataGenerator authgen = new AuthDataGenerator();

        public MsgPool msgPool { get; private set; }

        CancellationTokenSource canceller;

        Room room;
        string appId;
        string clientId;

        Uri uri;
        string authKey;
        volatile int pingInterval;
        volatile uint lastPingTime;
        CancellationTokenSource pingerDelayCanceller;

        TaskCompletionSource<Task> senderTaskSource;
        TaskCompletionSource<Task> pingerTaskSource;

        BlockingCollection<byte[]> evBufPool;
        uint evSeqNum;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public Connection(Room room, string clientId, JoinedRoom joined)
        {
            this.canceller = new CancellationTokenSource();
            this.room = room;
            this.appId = joined.roomInfo.appId;
            this.clientId = clientId;
            this.uri = new Uri(joined.url);
            this.authKey = joined.authKey;
            this.pingInterval = calcPingInterval(room.ClientDeadline);
            this.pingerDelayCanceller = new CancellationTokenSource();

            this.evSeqNum = 0;
            this.evBufPool = new BlockingCollection<byte[]>(
                new ConcurrentStack<byte[]>(), EvBufPoolSize);
            for (var i = 0; i<EvBufPoolSize; i++)
            {
                evBufPool.Add(new byte[EvBufInitialSize]);
            }

            this.msgPool = new MsgPool(MsgPoolSize, MsgBufInitialSize);
        }

        /// <summary>
        ///   強制切断
        /// </summary>
        public void Cancel()
        {
            canceller.Cancel();
        }

        /// <summary>
        ///   websocket接続をはじめる
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     NormalClosure or EndpointUnavailable まで自動再接続する
        ///     もしくはクライアントからの強制切断
        ///   </para>
        /// </remarks>
        public async Task Start()
        {
            int reconnection = 0;

            while(true)
            {
                Exception lastException;
                var retryInterval = Task.Delay(RetryIntervalMilliSec);

                if (canceller.IsCancellationRequested) {
                    return;
                }

                var cts = CancellationTokenSource.CreateLinkedTokenSource(canceller.Token);

                // Receiverの中でEvPeerReadyを受け取ったらSender/Pingerを起動する
                // SenderのTaskをawaitしたいのでこれで受け取る
                senderTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);
                pingerTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);

                try
                {
                    var ws = await Connect(cts.Token);
                    reconnection = 0;

                    var tasks = new Task[]
                    {
                        Task.Run(async() => await Receiver(ws, cts.Token)),
                        Task.Run(async() => await await senderTaskSource.Task),
                        Task.Run(async() => await await pingerTaskSource.Task),
                    };

                    await tasks[Task.WaitAny(tasks)];

                    // finish task without exception: unreconnectable. don't retry.
                    return;
                }
                catch(WebSocketException e)
                {
                    switch (e.WebSocketErrorCode)
                    {
                        case WebSocketError.NotAWebSocket:
                        case WebSocketError.UnsupportedProtocol:
                        case WebSocketError.UnsupportedVersion:
                            // unnable to connect. don't retry.
                            throw;
                    }

                    // retry on other error.
                    lastException = e;
                }
                catch(Exception e)
                {
                    // retry
                    lastException = e;
                }
                finally
                {
                    senderTaskSource.TrySetCanceled();
                    pingerTaskSource.TrySetCanceled();
                    cts.Cancel();
                }

                if (canceller.IsCancellationRequested) {
                    return;
                }

                if (++reconnection > MaxReconnection)
                {
                    throw new Exception($"MaxReconnection: {lastException.Message}", lastException);
                }

                room.handleError(lastException);

                await retryInterval;
            }
        }

        /// <summary>
        ///   Room.handleEvent()に渡したEventを使い終わったら返却する.
        /// </summary>
        public void ReturnEventBuffer(Event ev)
        {
            if (ev.BufferArray != null)
            {
                evBufPool.Add(ev.BufferArray);
            }
        }

        /// <summary>
        ///   Ping間隔を更新する
        /// </summary>
        public void UpdatePingInterval(uint deadline)
        {
            var canceller = pingerDelayCanceller;

            pingInterval = calcPingInterval(deadline);
            pingerDelayCanceller = new CancellationTokenSource();

            canceller.Cancel();
        }

        /// <summary>
        ///   Websocketで接続する
        /// </summary>
        private async Task<ClientWebSocket> Connect(CancellationToken ct)
        {
            var ws = new ClientWebSocket();
            var authdata = authgen.Generate(authKey, clientId);
            ws.Options.SetRequestHeader("Authorization", "Bearer " + authdata);
            ws.Options.SetRequestHeader("Wsnet2-App", appId);
            ws.Options.SetRequestHeader("Wsnet2-User", clientId);
            ws.Options.SetRequestHeader("Wsnet2-LastEventSeq", evSeqNum.ToString());

            await ws.ConnectAsync(uri, ct);
            return ws;
        }

        /// <summary>
        ///   Event受信ループ
        /// </summary>
        private async Task Receiver(ClientWebSocket ws, CancellationToken ct)
        {
            while(true)
            {
                ct.ThrowIfCancellationRequested();
                var ev = await ReceiveEvent(ws, ct);

                if (ev.IsRegular)
                {
                    if (ev.SequenceNum != evSeqNum+1)
                    {
                        // todo: reconnectable?
                        evBufPool.Add(ev.BufferArray);
                        throw new Exception($"invalid event sequence number: {ev.SequenceNum} wants {evSeqNum+1}");
                    }

                    evSeqNum++;
                }

                switch (ev.Type)
                {
                    case EvType.PeerReady:
                        var evpr = ev as EvPeerReady;
                        var sender = Task.Run(async() => await Sender(ws, evpr.LastMsgSeqNum+1, ct));
                        var pinger = Task.Run(async() => await Pinger(ws, ct));
                        senderTaskSource.TrySetResult(sender);
                        pingerTaskSource.TrySetResult(pinger);
                        break;
                    case EvType.Closed:
                        room.handleEvent(ev);
                        return;
                    case EvType.Pong:
                        onPong(ev as EvPong);
                        room.handleEvent(ev);
                        break;
                    default:
                        room.handleEvent(ev);
                        break;
                }
            }
        }

        /// <summary>
        ///   Eventの受信
        /// </summary>
        private async Task<Event> ReceiveEvent(WebSocket ws, CancellationToken ct)
        {
            var buf = evBufPool.Take(ct);
            try
            {
                var pos = 0;
                while(true){
                    var seg = new ArraySegment<byte>(buf, pos, buf.Length-pos);
                    var ret = await ws.ReceiveAsync(seg, ct);

                    if (ret.CloseStatus.HasValue)
                    {
                        switch (ret.CloseStatus.Value)
                        {
                            case WebSocketCloseStatus.NormalClosure:
                            case WebSocketCloseStatus.EndpointUnavailable:
                                // unreconnectable states.
                                evBufPool.Add(buf);
                                return new EvClosed(ret.CloseStatusDescription);
                            default:
                                throw new Exception("ws status:("+ret.CloseStatus.Value+") "+ret.CloseStatusDescription);
                        }
                    }

                    pos += ret.Count;
                    if (ret.EndOfMessage) {
                        break;
                    }

                    // メッセージがbufに収まらないときはbufをリサイズして続きを受信
                    Array.Resize(ref buf, buf.Length*2);
                }

                return Event.Parse(new ArraySegment<byte>(buf, 0, pos));
            }
            catch(Exception)
            {
                evBufPool.Add(buf);
                throw;
            }
        }

        /// <summary>
        ///   Msg送信ループ
        /// </summary>
        /// <param name="seqNum">開始Msg通し番号</param>
        /// <param name="ct">ループ停止するトークン</param>
        private async Task Sender(ClientWebSocket ws, int seqNum, CancellationToken ct)
        {
            while (true)
            {
                msgPool.Wait(ct);

                ArraySegment<byte>? msg;
                while ((msg = msgPool.Take(seqNum)).HasValue)
                {
                    ct.ThrowIfCancellationRequested();
                    await ws.SendAsync(msg.Value, WebSocketMessageType.Binary, true, ct);
                    seqNum++;
                }
            }
        }

        /// <summary>
        ///   Ping送信ループ
        /// </summary>
        private async Task Pinger(ClientWebSocket ws, CancellationToken ct)
        {
            var msg = new MsgPing();

            while (true)
            {
                ct.ThrowIfCancellationRequested();

                var interval = Task.Delay(pingInterval, pingerDelayCanceller.Token);
                var time = (uint)msg.SetTimestamp();
                await ws.SendAsync(msg.Value, WebSocketMessageType.Binary, true, ct);
                lastPingTime = time;
                try
                {
                    await interval;
                    // 対応するPongが返ってきていたらlastPingTimeは書き換わっている
                    if (lastPingTime == time)
                    {
                        throw new Exception("Pong unreceived");
                    }
                }
                catch(TaskCanceledException)
                {
                    // pingerDelayCancellerによるcancelは無視
                }
            }
        }

        /// <summary>
        ///   Pong受信時の処理
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     lastPingTimeのPingに対応するPongを受け取ったらlastPingTimeを異なる値に変更
        ///   </para>
        /// </remarks>
        private void onPong(EvPong ev)
        {
            var time = (uint)ev.PingTimestamp;
            if (lastPingTime == time)
            {
                lastPingTime ^= uint.MaxValue;
            }
        }

        /// <summary>
        ///   サーバから通知されたdeadline(秒)からPing間隔(ミリ秒)を計算
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     旧wsnetではdeadlineは通常25秒、Ping間隔は5秒固定だったので、
        ///     deadline/5を基準として最大最小間隔を超えない値にする。
        ///   </para>
        /// </remarks>
        private int calcPingInterval(uint deadline)
        {
            var ms = (int)deadline * 1000 / 5;
            return (ms < MinPingIntervalMilliSec) ? MinPingIntervalMilliSec
                : (ms > MaxPingIntervalMilliSec) ? MaxPingIntervalMilliSec : ms;
        }
    }
}
