using System;
using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Security.Cryptography;
using System.Threading.Tasks;
using System.Threading;

namespace WSNet2
{
    class Connection
    {
        /// <summary>受信バッファ拡張単位</summary>
        /// <remarks>Ethernet frame payload size</remarks>
        const int EvBufExpandSize = 1500;

        /// <summary>Closeメッセージを送り返すときのタイムアウト</summary>
        /// <remarks>サーバ側のconn.Close()までの猶予時間に合わせる</remarks>
        const int SendCloseTimeout = 3000;

        static AuthDataGenerator authgen = new AuthDataGenerator();

        public MsgPool msgPool { get; private set; }

        CancellationTokenSource canceller;
        DateTime reconnectLimit;

        Room room;
        string appId;
        string clientId;

        Uri uri;
        string authKey;
        HMAC hmac;
        volatile int pingInterval;
        volatile uint lastPingTime;
        CancellationTokenSource pingerDelayCanceller;
        SemaphoreSlim sendSemaphore;

        TaskCompletionSource<Task> senderTaskSource;
        TaskCompletionSource<Task> pingerTaskSource;

        BlockingCollection<byte[]> evBufPool;
        uint evSeqNum;

        Logger logger;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public Connection(Room room, string clientId, HMAC hmac, JoinedRoom joined, Logger logger)
        {
            this.logger = logger;
            this.canceller = new CancellationTokenSource();
            this.reconnectLimit = DateTime.Now.AddSeconds(room.ClientDeadline);
            this.room = room;
            this.appId = joined.roomInfo.appId;
            this.clientId = clientId;
            this.uri = new Uri(joined.url);
            this.authKey = joined.authKey;
            this.hmac = hmac;
            this.pingInterval = calcPingInterval(room.ClientDeadline);
            this.pingerDelayCanceller = new CancellationTokenSource();
            this.sendSemaphore = new SemaphoreSlim(1, 1);

            this.evSeqNum = 0;
            this.evBufPool = new BlockingCollection<byte[]>(
                new ConcurrentStack<byte[]>(), WSNet2Settings.EvPoolSize);
            for (var i = 0; i < WSNet2Settings.EvPoolSize; i++)
            {
                evBufPool.Add(new byte[WSNet2Settings.EvBufInitialSize]);
            }

            this.msgPool = new MsgPool(WSNet2Settings.MsgPoolSize, WSNet2Settings.MsgBufInitialSize, hmac);
        }

        /// <summary>
        ///   強制切断
        /// </summary>
        public void Cancel()
        {
            canceller.Cancel();
            logger?.Debug("connection canceled");
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
            bool connected = false;
            int reconnection = 0;

            while (true)
            {
                Exception lastException;
                var retryInterval = Task.Delay(WSNet2Settings.RetryIntervalMilliSec);

                if (canceller.IsCancellationRequested)
                {
                    return;
                }

                var cts = CancellationTokenSource.CreateLinkedTokenSource(canceller.Token);

                // Receiverの中でEvPeerReadyを受け取ったらSender/Pingerを起動する
                // SenderのTaskをawaitしたいのでこれで受け取る
                senderTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);
                pingerTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);

                var receiverTask = Task.CompletedTask;

                try
                {
                    using var ws = await Connect(cts.Token);
                    connected = true;
                    room.ConnectionStateChanged(true);

                    receiverTask = Task.Run(async () => await Receiver(ws, cts.Token));
                    await await Task.WhenAny(
                        receiverTask,
                        Task.Run(async () => await await senderTaskSource.Task),
                        Task.Run(async () => await await pingerTaskSource.Task));

                    // finish task without exception: unreconnectable. don't retry.
                    return;
                }
                catch (Exception e)
                {
                    logger?.Warning(e, "connection exception: {0}", e);
                    // retry
                    lastException = e;
                }
                finally
                {
                    senderTaskSource.TrySetCanceled();
                    pingerTaskSource.TrySetCanceled();
                    cts.Cancel();
                    if (connected)
                    {
                        connected = false;
                        room.ConnectionStateChanged(false);
                    }
                }

                if (canceller.IsCancellationRequested)
                {
                    return;
                }

                room.handleError(lastException);

                await retryInterval;

                try
                {
                    await receiverTask; // recconectLimitへの書き込みがなくなるのを待つ
                }
                catch { }

                if (DateTime.Now > reconnectLimit)
                {
                    throw new Exception($"Gave up on Reconnection: {lastException.Message}", lastException);
                }

                reconnection++;
                logger?.Info("reconnect now:{0}, limit:{1}, count:{2}", DateTime.Now, reconnectLimit, reconnection);
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
            var authdata = authgen.GenerateBearer(authKey, clientId);
            ws.Options.SetRequestHeader("Authorization", authdata);
            ws.Options.SetRequestHeader("Wsnet2-App", appId);
            ws.Options.SetRequestHeader("Wsnet2-User", clientId);
            ws.Options.SetRequestHeader("Wsnet2-LastEventSeq", evSeqNum.ToString());
            ws.Options.AddSubProtocol("wsnet2");

            logger?.Info("connecting to {0}", uri);
            var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
            cts.CancelAfter(WSNet2Settings.ConnectTimeoutMilliSec);
            NetworkInformer.OnRoomConnectRequest(room, uri.AbsoluteUri);
            await ws.ConnectAsync(uri, cts.Token);
            logger?.Info("connected");
            return ws;
        }

        /// <summary>
        ///   Event受信ループ
        /// </summary>
        private async Task Receiver(ClientWebSocket ws, CancellationToken ct)
        {
            try
            {
                while (true)
                {
                    ct.ThrowIfCancellationRequested();
                    var ev = await ReceiveEvent(ws, ct);

                    this.reconnectLimit = DateTime.Now.AddSeconds(room.ClientDeadline);

                    if (ev.IsRegular)
                    {
                        if (ev.SequenceNum != evSeqNum + 1)
                        {
                            evBufPool.Add(ev.BufferArray);
                            throw new Exception($"invalid event sequence number: {ev.SequenceNum} wants {evSeqNum + 1}");
                        }

                        evSeqNum++;
                    }

                    switch (ev.Type)
                    {
                        case EvType.Closed:
                            // 正常終了。もう再接続しない。
                            room.handleEvent(ev);
                            return;

                        case EvType.PeerReady:
                            var evpr = ev as EvPeerReady;
                            logger?.Info("receive peer-ready: lastMsgSeqNum={0}", evpr.LastMsgSeqNum);
                            var sender = Task.Run(async () => await Sender(ws, evpr.LastMsgSeqNum + 1, ct));
                            var pinger = Task.Run(async () => await Pinger(ws, ct));
                            senderTaskSource.TrySetResult(sender);
                            pingerTaskSource.TrySetResult(pinger);
                            break;
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
            catch (OperationCanceledException)
            {
                // ctのキャンセルはループを抜けて終了
            }
        }

        /// <summary>
        ///   Eventの受信
        /// </summary>
        private async Task<Event> ReceiveEvent(ClientWebSocket ws, CancellationToken ct)
        {
            var buf = evBufPool.Take(ct);
            try
            {
                var pos = 0;
                while (true)
                {
                    var seg = new ArraySegment<byte>(buf, pos, buf.Length - pos);
                    var ret = await ws.ReceiveAsync(seg, ct);

                    if (ret.MessageType == WebSocketMessageType.Close)
                    {
                        await SendClose(ws, ret.CloseStatusDescription, ct);

                        switch (ret.CloseStatus)
                        {
                            case WebSocketCloseStatus.NormalClosure:
                            case WebSocketCloseStatus.EndpointUnavailable: // server: CloseGoingAway
                                // unreconnectable states.
                                evBufPool.Add(buf);
                                return new EvClosed(ret.CloseStatusDescription);
                            default:
                                throw new Exception("ws status:(" + ret.CloseStatus.Value + ") " + ret.CloseStatusDescription);
                        }
                    }

                    pos += ret.Count;
                    if (ret.EndOfMessage)
                    {
                        break;
                    }

                    // バッファの空きが少ないときは拡張して続きを受信
                    if (buf.Length - pos < EvBufExpandSize)
                    {
                        var expandSize = (buf.Length < EvBufExpandSize) ? buf.Length : EvBufExpandSize;
                        Array.Resize(ref buf, buf.Length + expandSize);
                    }
                }

                var ev = Event.Parse(new ArraySegment<byte>(buf, 0, pos));
                NetworkInformer.OnRoomReceive(room, pos, ev);
                return ev;
            }
            catch (Exception)
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
                    if (ct.IsCancellationRequested)
                    {
                        return; // ctのキャンセルで終了
                    }

                    await Send(ws, msg.Value, ct);
                    seqNum++;
                }
            }
        }

        /// <summary>
        ///   Ping送信ループ
        /// </summary>
        private async Task Pinger(ClientWebSocket ws, CancellationToken ct)
        {
            var msg = new MsgPing(hmac);

            while (true)
            {
                if (ct.IsCancellationRequested)
                {
                    return; // ctのキャンセルで終了
                }

                var interval = Task.Delay(pingInterval, pingerDelayCanceller.Token);
                var time = (uint)msg.SetTimestamp();
                lastPingTime = time;
                await Send(ws, msg.Value, ct);
                try
                {
                    await interval;
                    // 対応するPongが返ってきていたらlastPingTimeは書き換わっている
                    if (lastPingTime == time)
                    {
                        throw new Exception("Pong unreceived");
                    }
                }
                catch (TaskCanceledException)
                {
                    // pingerDelayCancellerによるcancelは無視
                }
            }
        }

        /// <summary>
        ///   websocketメッセージを送信
        /// </summary>
        private async Task Send(ClientWebSocket ws, ArraySegment<byte> msg, CancellationToken ct)
        {
            await sendSemaphore.WaitAsync(ct);
            try
            {
                NetworkInformer.OnRoomSend(room, msg);
                await ws.SendAsync(msg, WebSocketMessageType.Binary, true, ct);
            }
            finally
            {
                sendSemaphore.Release();
            }
        }

        private async Task SendClose(ClientWebSocket ws, string msg, CancellationToken ct)
        {
            await sendSemaphore.WaitAsync(ct);
            try
            {
                var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
                cts.CancelAfter(SendCloseTimeout);
                await ws.CloseAsync(WebSocketCloseStatus.NormalClosure, msg, cts.Token);
            }
            finally
            {
                sendSemaphore.Release();
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
            return (ms < WSNet2Settings.MinPingIntervalMilliSec) ? WSNet2Settings.MinPingIntervalMilliSec
                : (ms > WSNet2Settings.MaxPingIntervalMilliSec) ? WSNet2Settings.MaxPingIntervalMilliSec : ms;
        }
    }
}
