using System;
using System.Collections.Concurrent;
using System.Net.WebSockets;
using System.Security.Cryptography;
using System.Threading.Tasks;
using System.Threading;

namespace WSNet2.Core
{
    class Connection
    {
        static AuthDataGenerator authgen = new AuthDataGenerator();

        public MsgPool msgPool { get; private set; }

        CancellationTokenSource canceller;

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
                catch (WebSocketException e)
                {
                    logger?.Error(e, "websocket exception: {0}", e);
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
                catch (Exception e)
                {
                    logger?.Error(e, "connection exception: {0}", e);
                    // retry
                    lastException = e;
                }
                finally
                {
                    senderTaskSource.TrySetCanceled();
                    pingerTaskSource.TrySetCanceled();
                    cts.Cancel();
                }

                if (canceller.IsCancellationRequested)
                {
                    return;
                }

                if (++reconnection > WSNet2Settings.MaxReconnection)
                {
                    throw new Exception($"MaxReconnection: {lastException.Message}", lastException);
                }

                room.handleError(lastException);

                logger?.Info("reconnect: {0}", reconnection);

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
            var authdata = authgen.GenerateBearer(authKey, clientId);
            ws.Options.SetRequestHeader("Authorization", authdata);
            ws.Options.SetRequestHeader("Wsnet2-App", appId);
            ws.Options.SetRequestHeader("Wsnet2-User", clientId);
            ws.Options.SetRequestHeader("Wsnet2-LastEventSeq", evSeqNum.ToString());

            logger?.Info("connecting to {0}", uri);
            var cts = CancellationTokenSource.CreateLinkedTokenSource(ct);
            cts.CancelAfter(WSNet2Settings.ConnectTimeoutMilliSec);
            await ws.ConnectAsync(uri, cts.Token);
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

                    if (ev.IsRegular)
                    {
                        if (ev.SequenceNum != evSeqNum + 1)
                        {
                            // todo: reconnectable?
                            evBufPool.Add(ev.BufferArray);
                            throw new Exception($"invalid event sequence number: {ev.SequenceNum} wants {evSeqNum + 1}");
                        }

                        evSeqNum++;
                    }

                    switch (ev.Type)
                    {
                        case EvType.PeerReady:
                            var evpr = ev as EvPeerReady;
                            var sender = Task.Run(async () => await Sender(ws, evpr.LastMsgSeqNum + 1, ct));
                            var pinger = Task.Run(async () => await Pinger(ws, ct));
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
            catch (OperationCanceledException)
            {
                // ctのキャンセルはループを抜けて終了
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
                while (true)
                {
                    var seg = new ArraySegment<byte>(buf, pos, buf.Length - pos);
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
                                throw new Exception("ws status:(" + ret.CloseStatus.Value + ") " + ret.CloseStatusDescription);
                        }
                    }

                    pos += ret.Count;
                    if (ret.EndOfMessage)
                    {
                        break;
                    }

                    // メッセージがbufに収まらないときはbufをリサイズして続きを受信
                    Array.Resize(ref buf, buf.Length * 2);
                }

                return Event.Parse(new ArraySegment<byte>(buf, 0, pos));
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
            try
            {
                while (true)
                {
                    msgPool.Wait(ct);

                    ArraySegment<byte>? msg;
                    while ((msg = msgPool.Take(seqNum)).HasValue)
                    {
                        ct.ThrowIfCancellationRequested();
                        await Send(ws, msg.Value, ct);
                        seqNum++;
                    }
                }
            }
            catch (OperationCanceledException)
            {
                // ctのキャンセルはループを抜けて終了
            }
        }

        /// <summary>
        ///   Ping送信ループ
        /// </summary>
        private async Task Pinger(ClientWebSocket ws, CancellationToken ct)
        {
            var msg = new MsgPing(hmac);

            try
            {
                while (true)
                {
                    ct.ThrowIfCancellationRequested();

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
            catch (OperationCanceledException)
            {
                // ctのキャンセルはループを抜けて終了
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
                await ws.SendAsync(msg, WebSocketMessageType.Binary, true, ct);
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
