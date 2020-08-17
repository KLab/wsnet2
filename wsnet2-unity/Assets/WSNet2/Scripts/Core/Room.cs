using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace WSNet2.Core
{
    /// <summary>
    ///   Room
    /// </summary>
    public class Room
    {
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

        /// <summary>Callbackループの動作状態</summary>
        public bool Running { get; set; }

        /// <summary>終了したかどうか</summary>
        public bool Closed { get; private set; }

        /// <summary>自分自身のPlayer</summary>
        public Player Me { get; private set; }

        /// <summary>部屋内の全Player</summary>
        public IReadOnlyDictionary<string, Player> Players { get { return players; } }

        /// <summary>マスタークライアント</summary>
        public Player Master {
            get
            {
                return players[masterId];
            }
        }

        /// <summary>Ping応答時間 (millisec)</summary>
        public ulong RttMillisec { get; private set; }

        /// <summary>全Playerの最終メッセージ受信時刻 (millisec)</summary>
        public IReadOnlyDictionary<string, ulong> LastMsgTimestamps { get; private set; }

        string myId;
        Dictionary<string, object> publicProps;
        Dictionary<string, object> privateProps;

        Dictionary<string, Player> players;
        string masterId;

        RoomInfo info;
        EventReceiver eventReceiver;

        CallbackPool callbackPool = new CallbackPool();

        Connection con;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="joined">lobbyからの入室完了レスポンス</param>
        /// <param name="myId">自身のID</param>
        /// <param name="receiver">イベントレシーバ</param>
        public Room(JoinedResponse joined, string myId, EventReceiver receiver)
        {
            this.myId = myId;
            this.info = joined.roomInfo;

            this.con = new Connection(this, myId, joined);

            this.eventReceiver = receiver;
            this.Running = true;
            this.Closed = false;

            var reader = Serialization.NewReader(new ArraySegment<byte>(info.publicProps));
            publicProps = reader.ReadDict();

            reader = Serialization.NewReader(new ArraySegment<byte>(info.privateProps));
            privateProps = reader.ReadDict();

            players = new Dictionary<string, Player>(joined.players.Length);
            foreach (var p in joined.players)
            {
                var player = new Player(p);
                players[p.Id] = player;
                if (p.Id == myId)
                {
                    Me = player;
                }
            }

            this.masterId = joined.masterId;
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

        public async Task Start()
        {
            string msg;

            try{
                msg = await con.Start();
            }
            catch(Exception e)
            {
                OnError(e);
                msg = e.Message;
            }
            callbackPool.Add(() =>
            {
                if (!Closed)
                {
                    Closed = true;
                    eventReceiver.OnClosed(msg);
                }
            });
        }

        public void Leave()
        {
            con.msgPool.PostLeave();
        }

        public void RPC(Action<string, string> rpc, string param, params string[] targets)
        {
            con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }

        public void RPC<T>(Action<string, T> rpc, T param, params string[] targets) where T : class, IWSNetSerializable
        {
            con.msgPool.PostRPC(getRpcId(rpc), param, targets);
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


        public void OnError(Exception e)
        {
            callbackPool.Add(() => eventReceiver.OnError(e));
        }

        public void OnEvent(Event ev)
        {
            switch (ev)
            {
                case EvClosed evClosed:
                    OnEvClosed(evClosed);
                    break;

                case EvPong evPong:
                    OnEvPong(evPong);
                    break;
                case EvJoined evJoined:
                    OnEvJoined(evJoined);
                    break;
                case EvLeft evLeft:
                    OnEvLeft(evLeft);
                    break;
                case EvRPC evRpc:
                    OnEvRPC(evRpc);
                    break;

                default:
                    con.ReturnEventBuffer(ev);
                    throw new Exception($"unknown event: {ev}");
            }

            // Event受信に使ったバッファはcallbackで参照されるので
            // callbackが呼ばれて使い終わってから返却
            callbackPool.Add(() => con.ReturnEventBuffer(ev));
        }


        /// <summary>
        ///   Pongイベント
        /// </summary>
        private void OnEvPong(EvPong ev)
        {
            callbackPool.Add(() =>
            {
                info.watchers = ev.WatcherCount;
                RttMillisec = ev.RTT;
                LastMsgTimestamps = ev.lastMsgTimestamps;
            });
        }

        /// <summary>
        ///   入室イベント
        /// </summary>
        private void OnEvJoined(EvJoined ev)
        {
            if (ev.ClientID == myId)
            {
                callbackPool.Add(() =>
                {
                    Me.Props = ev.GetProps(Me.Props);
                    eventReceiver.OnJoined(Me);
                });
                return;
            }

            callbackPool.Add(()=>
            {
                var player = new Player(ev.ClientID, ev.GetProps());
                players[player.Id] = player;
                eventReceiver.OnOtherPlayerJoined(player);
            });
        }

        /// <summary>
        ///   プレイヤー退室イベント
        /// </summary>
        private void OnEvLeft(EvLeft ev)
        {
            callbackPool.Add(() =>
            {
                var player = players[ev.ClientID];

                if (masterId == player.Id)
                {
                    masterId = ev.MasterID;
                    eventReceiver.OnMasterPlayerSwitched(player, Master);
                }

                players.Remove(player.Id);
                eventReceiver.OnOtherPlayerLeft(player);
            });
        }

        /// <summary>
        ///   RPCイベント
        /// </summary>
        private void OnEvRPC(EvRPC ev)
        {
            // fixme: RPCがまだ登録されていないかもしれない。callbackの中で判定すべき。
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
                if (!Closed)
                {
                    Closed = true;
                    eventReceiver.OnClosed(ev.Description);
                }
            });
        }
    }
}
