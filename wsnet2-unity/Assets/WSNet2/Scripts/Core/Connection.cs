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

        /// <summary>最大再接続試行回数</summary>
        const int MaxReconnection = 30;

        /// <summary>再接続インターバル (milli seconds)</summary>
        const int RetryIntervalMilliSec = 1000;

        public MsgPool msgPool { get; private set; }

        Room room;
        string appId;
        string clientId;

        Uri uri;
        AuthToken token;
        int pingInterval;
        CancellationTokenSource pingerDelayCanceller;

        TaskCompletionSource<Task> senderTaskSource;
        TaskCompletionSource<Task> pingerTaskSource;
        int reconnection;

        BlockingCollection<byte[]> evBufPool;
        uint evSeqNum;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public Connection(Room room, string clientId, JoinedResponse joined)
        {
            this.room = room;
            this.appId = joined.roomInfo.appId;
            this.clientId = clientId;
            this.uri = new Uri(joined.url);
            this.token = joined.token;
            this.pingInterval = calcPingInterval(room.ClientDeadline);
            this.pingerDelayCanceller = new CancellationTokenSource();
            this.reconnection = 0;

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
            while(true)
            {
                Exception lastException;
                var retryInterval = Task.Delay(RetryIntervalMilliSec);

                var cts = new CancellationTokenSource();

                // Receiverの中でEvPeerReadyを受け取ったらSender/Pingerを起動する
                // SenderのTaskをawaitしたいのでこれで受け取る
                senderTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);
                pingerTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);

                try
                {
                    var ws = await Connect(cts.Token);

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

            pingerDelayCanceller = new CancellationTokenSource();
            pingInterval = calcPingInterval(deadline);

            canceller.Cancel();
        }

        /// <summary>
        ///   Websocketで接続する
        /// </summary>
        private async Task<ClientWebSocket> Connect(CancellationToken ct)
        {
            var ws = new ClientWebSocket();
            ws.Options.SetRequestHeader("X-Wsnet-App", appId);
            ws.Options.SetRequestHeader("X-Wsnet-User", clientId);
            ws.Options.SetRequestHeader("X-Wsnet-Nonce", token.nonce);
            ws.Options.SetRequestHeader("X-Wsnet-Hash", token.hash);
            ws.Options.SetRequestHeader("X-Wsnet-LastEventSeq", evSeqNum.ToString());

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
                msg.SetTimestamp();
                await ws.SendAsync(msg.Value, WebSocketMessageType.Binary, true, ct);
                try
                {
                    await interval;
                }
                catch(TaskCanceledException)
                {
                    // pingerDelayCancellerによるcancelは無視
                }
            }
        }

        /// <summary>
        ///   サーバから通知されたdeadline(秒)からPing間隔(ミリ秒)を計算
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     deadlineの半分の時間停止していても切断しないような間隔とする.
        ///   </para>
        /// </remarks>
        private int calcPingInterval(uint deadline)
        {
            return (int)deadline * 1000 / 3;
        }
    }
}
