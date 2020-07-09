using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;

namespace WSNet2.Core
{
    /// <summary>
    ///   Room
    /// </summary>
    public class Room
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
        const int MaxReconnection = 100;

        /// <summary>再接続インターバル (milli seconds)</summary>
        const int RetryIntervalMilliSec = 1000;

        /// <summary>RoomのMasterをRPCの対象に指定</summary
        public const string[] RPCToMaster = null;

        /// <summary>RoomID</summary>
        public string Id { get { return info.id; } }

        /// <summary>検索可能</summary>
        public bool Visible { get { return info.visible; } }

        /// <summary>入室可能</summary>
        public bool Joinable { get { return info.joinable; } }

        /// <summary>観戦可能</summary>
        public bool Watchable { get { return info.watchable; } }

        /// <summary>Eventループの動作状態</summary>
        public bool Running { get; set; }

        /// <summary>終了したかどうか</summary>
        public bool Closed { get; private set; }

        /// <summary>自分自身のPlayer</summary>
        public Player Me { get; private set; }

        Dictionary<string, object> publicProps;
        Dictionary<string, object> privateProps;
        List<Player> players;

        RoomInfo info;
        Uri uri;
        EventReceiver eventReceiver;

        ClientWebSocket ws;
        TaskCompletionSource<Task> senderTaskSource;
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
        BlockingCollection<bool> hasMsg;
        MsgPool msgPool;

        CallbackPool callbackPool = new CallbackPool();

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="joined">lobbyからの入室完了レスポンス</param>
        /// <param name="myId">自身のID</param>
        /// <param name="receiver">イベントレシーバ</param>
        public Room(JoinedResponse joined, string myId, EventReceiver receiver)
        {
            this.info = joined.roomInfo;
            this.uri = new Uri(joined.url);
            this.eventReceiver = receiver;
            this.Running = true;
            this.Closed = false;
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

            var reader = Serialization.NewReader(new ArraySegment<byte>(info.publicProps));
            publicProps = reader.ReadDict();

            reader = Serialization.NewReader(new ArraySegment<byte>(info.privateProps));
            privateProps = reader.ReadDict();

            players = new List<Player>(joined.players.Length);
            foreach (var p in joined.players)
            {
                var player = new Player(p);
                players.Add(player);
                if (p.Id == myId)
                {
                    Me = player;
                }
            }
        }

        /// <summary>
        ///   溜めたCallbackを処理する
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     WSNet2Client.ProcessCallbackから呼ばれる。
        ///     Unityではメインスレッドで呼ぶようにする。
        ///   </para>
        /// </remarks>
        public void ProcessCallback()
        {
            if (Running)
            {
                callbackPool.Process();
            }
        }

        /// <summary>
        ///   websocket接続をはじめる
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     NormalClosure or EndpointUnavailable まで自動再接続する (TODO)
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

                // Receiverの中でEvPeerReadyを受け取ったらSenderを起動する
                // SenderのTaskをawaitしたいのでこれで受け取る
                senderTaskSource = new TaskCompletionSource<Task>(TaskCreationOptions.RunContinuationsAsynchronously);

                try
                {
                    ws = await Connect(cts.Token);

                    var tasks = new Task[]
                    {
                        Task.Run(async() => await Receiver(cts.Token)),
                        Task.Run(async() => await await senderTaskSource.Task),
                    };
                    await tasks[Task.WaitAny(tasks)];

                    // finish task without exception: unreconnectable
                    return;
                }
                catch(WebSocketException e)
                {
                    switch (e.WebSocketErrorCode)
                    {
                        case WebSocketError.NotAWebSocket:
                        case WebSocketError.UnsupportedProtocol:
                        case WebSocketError.UnsupportedVersion:
                            callbackPool.Add(() => {
                                Closed = true;
                                eventReceiver.OnError(e);
                                eventReceiver.OnClosed(e.Message);
                            });
                            return;
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

                if (++reconnection > MaxReconnection)
                {
                    callbackPool.Add(() => {
                        Closed = true;
                        var msg = $"MaxReconnection: {lastException.Message}";
                        var e = new Exception(msg, lastException);
                        eventReceiver.OnError(e);
                        eventReceiver.OnClosed(e.Message);
                    });
                    return;
                }

                await retryInterval;
            }
        }

        public void RPC(Action<string, string> rpc, string param, params string[] targets)
        {
            msgPool.PostRPC(getRpcId(rpc), param, targets);
            hasMsg.TryAdd(true);
        }

        public void RPC<T>(Action<string, T> rpc, T param, params string[] targets) where T : class, IWSNetSerializable
        {
            msgPool.PostRPC(getRpcId(rpc), param, targets);
            hasMsg.TryAdd(true);
        }

        private byte getRpcId(Delegate rpc)
        {
            byte rpcId;
            if (!eventReceiver.RPCMap.TryGetValue(rpc, out rpcId))
            {
                var msg = $"RPC target is not registered";
                throw new Exception(msg);
            }

            return rpcId;
        }

        /// <summary>
        ///   Websocketで接続する
        /// </summary>
        private async Task<ClientWebSocket> Connect(CancellationToken ct)
        {
            var ws = new ClientWebSocket();
            ws.Options.SetRequestHeader("X-Wsnet-App", info.appId);
            ws.Options.SetRequestHeader("X-Wsnet-User", Me.Id);
            ws.Options.SetRequestHeader("X-Wsnet-LastEventSeq", evSeqNum.ToString());

            await ws.ConnectAsync(uri, ct);
            return ws;
        }

        /// <summary>
        ///   Event受信ループ
        /// </summary>
        private async Task Receiver(CancellationToken ct)
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
                        throw new Exception($"invalid event sequence number: {ev.SequenceNum} wants {evSeqNum+1}");
                    }

                    evSeqNum++;
                }

                switch (ev)
                {
                    case EvClosed evClosed:
                        OnEvClosed(evClosed);
                        return;

                    case EvPeerReady evPeerReady:
                        OnEvPeerReady(evPeerReady, ct);
                        break;
                    case EvJoined evJoined:
                        OnEvJoined(evJoined);
                        break;
                    case EvRPC evRpc:
                        OnEvRPC(evRpc);
                        break;

                    default:
                        throw new Exception($"unknown event: {ev}");
                }

                // Event受信に使ったバッファはcallbackで参照されるので
                // callbackが呼ばれて使い終わってから返却
                callbackPool.Add(() => evBufPool.Add(ev.BufferArray));
            }
        }

        /// <summary>
        ///   Eventの受信
        /// </summary>
        private async Task<Event> ReceiveEvent(WebSocket ws, CancellationToken ct)
        {
            var buf = evBufPool.Take(ct);
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

        /// <summary>
        ///   Peer準備完了イベント
        /// </summary>
        private void OnEvPeerReady(EvPeerReady ev, CancellationToken ct)
        {
            var task = Task.Run(async() => await Sender(ev.LastMsgSeqNum+1, ct));
            senderTaskSource.TrySetResult(task);
        }

        /// <summary>
        ///   入室イベント
        /// </summary>
        private void OnEvJoined(EvJoined ev)
        {
            if (ev.ClientID == Me.Id)
            {
                Me.Props = ev.GetProps(Me.Props);
                callbackPool.Add(() => eventReceiver.OnJoined(Me));
                return;
            }

            callbackPool.Add(()=>
            {
                var player = new Player(ev.ClientID, ev.GetProps());
                players.Add(player);
                callbackPool.Add(() => eventReceiver.OnOtherPlayerJoined(player));
            });
        }

        /// <summary>
        ///   RPCイベント
        /// </summary>
        private void OnEvRPC(EvRPC ev)
        {
            if (ev.RpcID >= eventReceiver.RPCActions.Count)
            {
                var e = new Exception($"RpcID({ev.RpcID}) is not registered");
                callbackPool.Add(() => eventReceiver.OnError(e));
                return;
            }

            var action = eventReceiver.RPCActions[ev.RpcID];
            callbackPool.Add(() => action(ev.SenderID, ev.Reader));
        }

        /// <summary>
        ///   退室イベント
        /// </summary>
        private void OnEvClosed(EvClosed ev)
        {
            callbackPool.Add(() =>
            {
                Closed = true;
                eventReceiver.OnClosed(ev.Description);
            });
        }

        /// <summary>
        ///   Msg送信ループ
        /// </summary>
        /// <param name="seqNum">開始Msg通し番号</param>
        /// <param name="ct">ループ停止するトークン</param>
        private async Task Sender(int seqNum, CancellationToken ct)
        {
            do
            {
                ArraySegment<byte>? msg;
                while ((msg = msgPool.Take(seqNum)).HasValue)
                {
                    ct.ThrowIfCancellationRequested();
                    await ws.SendAsync(msg.Value, WebSocketMessageType.Binary, true, ct);
                    seqNum++;
                }
            }
            while (hasMsg.Take(ct));
        }
    }
}
