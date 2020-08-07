using System;
using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Threading.Tasks;
using System.Threading;

namespace WSNet2.Core
{
    class Connection
    {
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

        Room room;
        string appId;
        string clientId;

        Uri uri;
        AuthToken token;
        int pingInterval;

        ClientWebSocket ws;
        TaskCompletionSource<Task<string>> senderTaskSource;
        TaskCompletionSource<Task<string>> pingerTaskSource;
        int reconnection;

        BlockingCollection<byte[]> evBufPool;
        uint evSeqNum;

        ///<summary>PoolにMsgが追加されたフラグ</summary>
        /// <remarks>
        ///   <para>
        ///     msgPoolにAdd*したあとTryAdd(true)する。
        ///     送信ループがTake()で待機しているので、Addされたら動き始める。
        ///     サイズ=1にしておくことで、送信前に複数回Addされても1度のループで送信される。
        ///   </para>
        /// </remarks>
        public BlockingCollection<bool> hasMsg;
        public MsgPool msgPool;

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
            this.pingInterval = calcPingInterval((int)joined.deadline);
            this.reconnection = 0;

            this.evSeqNum = 0;
            this.evBufPool = new BlockingCollection<byte[]>(
                new ConcurrentStack<byte[]>(), EvBufPoolSize);
            for (var i = 0; i<EvBufPoolSize; i++)
            {
                evBufPool.Add(new byte[EvBufInitialSize]);
            }

            this.msgPool = new MsgPool(MsgPoolSize, MsgBufInitialSize);
            this.hasMsg = new BlockingCollection<bool>(1);
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
        public async Task<string> Start()
        {
            while(true)
            {
                Exception lastException;
                var retryInterval = Task.Delay(RetryIntervalMilliSec);

                var cts = new CancellationTokenSource();

                // Receiverの中でEvPeerReadyを受け取ったらSender/Pingerを起動する
                // SenderのTaskをawaitしたいのでこれで受け取る
                senderTaskSource = new TaskCompletionSource<Task<string>>(TaskCreationOptions.RunContinuationsAsynchronously);
                pingerTaskSource = new TaskCompletionSource<Task<string>>(TaskCreationOptions.RunContinuationsAsynchronously);

                try
                {
                    ws = await Connect(cts.Token);

                    var tasks = new Task<string>[]
                    {
                        Task.Run(async() => await Receiver(cts.Token)),
                        Task.Run(async() => await await senderTaskSource.Task),
                        Task.Run(async() => await await pingerTaskSource.Task),
                    };

                    // finish task without exception: unreconnectable. don't retry.
                    return await tasks[Task.WaitAny(tasks)];
                }
                catch(WebSocketException e)
                {
                    switch (e.WebSocketErrorCode)
                    {
                        case WebSocketError.NotAWebSocket:
                        case WebSocketError.UnsupportedProtocol:
                        case WebSocketError.UnsupportedVersion:
                            room.OnError(e);
                            return e.Message;
                    }

                    // retry on other exception
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
                    cts.Cancel();
                }

                room.OnError(lastException);

                if (++reconnection > MaxReconnection)
                {
                    return $"MaxReconnection: {lastException.Message}";
                }

                await retryInterval;
            }
        }

        /// <summary>
        ///   Room.OnEvent()に渡したEventを使い終わったら返却する.
        /// </summary>
        public void ReturnEventBuffer(Event ev)
        {
            if (ev.BufferArray != null)
            {
                evBufPool.Add(ev.BufferArray);
            }
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
        private async Task<string> Receiver(CancellationToken ct)
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
                        var sender = Task.Run(async() => await Sender(evpr.LastMsgSeqNum+1, ct));
                        var pinger = Task.Run(async() => await Pinger(ct));
                        senderTaskSource.TrySetResult(sender);
                        pingerTaskSource.TrySetResult(pinger);
                        break;
                    default:
                        room.OnEvent(ev);
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
                        evBufPool.Add(buf);
                        switch (ret.CloseStatus.Value)
                        {
                            case WebSocketCloseStatus.NormalClosure:
                            case WebSocketCloseStatus.EndpointUnavailable:
                                // unreconnectable states.
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
        private async Task<string> Sender(int seqNum, CancellationToken ct)
        {
            bool hasNext;
            do
            {
                hasNext = hasMsg.Take();

                ArraySegment<byte>? msg;
                while ((msg = msgPool.Take(seqNum)).HasValue)
                {
                    ct.ThrowIfCancellationRequested();
                    await ws.SendAsync(msg.Value, WebSocketMessageType.Binary, true, ct);
                    seqNum++;
                }
            }
            while (hasNext);

            await ws.CloseAsync(WebSocketCloseStatus.NormalClosure, "client leave", ct);

            return "msg sender closed.";
        }

        /// <summary>
        ///   Ping送信ループ
        /// </summary>
        private async Task<string> Pinger(CancellationToken ct)
        {
            var msg = new MsgPing();

            while (true)
            {
                ct.ThrowIfCancellationRequested();

                // todo: deadline変更時にDelayを中断したい
                var interval = Task.Delay(pingInterval);
                msg.SetTimestamp();
                await ws.SendAsync(msg.Value, WebSocketMessageType.Binary, true, ct);
                await interval;
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
        private int calcPingInterval(int deadline)
        {
            return deadline * 1000 / 3;
        }
    }
}
