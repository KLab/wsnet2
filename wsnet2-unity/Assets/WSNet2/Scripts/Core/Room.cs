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

        /// <summary>入室イベント通知</summary>
        public Action<Player> OnJoined;
        /// <summary>退室イベント通知</summary>
        public Action<string> OnClosed;
        /// <summary>他のプレイヤーの入室通知</summary>
        public Action<Player> OnOtherPlayerJoined;
        /// <summary>他のプレイヤーの退室通知</summary>
        public Action<Player> OnOtherPlayerLeft;
        /// <summary>マスタークライアントの変更通知</summary>
        public Action<Player, Player> OnMasterPlayerSwitched;
        /// <summary>部屋のプロパティの変更通知</summary>
        public Action<Dictionary<string, object>, Dictionary<string, object>> OnRoomPropertyChanged;
        /// <summary>プレイヤーのプロパティの変更通知</summary>
        public Action<Player, Dictionary<string, object>> OnPlayerPropertyChanged;
        /// <summary>エラー通知</summary>
        public Action<Exception> OnError;
        /// <summary>エラーによる切断通知</summary>
        public Action<Exception> OnErrorClosed;

        string myId;
        Dictionary<string, object> publicProps;
        Dictionary<string, object> privateProps;
        Dictionary<string, Player> players;
        string masterId;
        RoomInfo info;

        CallbackPool callbackPool = new CallbackPool();
        Dictionary<Delegate, byte> rpcMap;
        List<Action<string, SerialReader>> rpcActions;

        Connection con;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="joined">lobbyからの入室完了レスポンス</param>
        /// <param name="myId">自身のID</param>
        public Room(JoinedResponse joined, string myId)
        {
            this.myId = myId;
            this.info = joined.roomInfo;

            this.con = new Connection(this, myId, joined);

            this.rpcMap = new Dictionary<Delegate, byte>();
            this.rpcActions = new List<Action<string, SerialReader>>();

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

        internal async Task Start()
        {
            try{
                await con.Start();
            }
            catch(Exception e)
            {
                callbackPool.Add(() =>
                {
                    Closed = true;
                    if (OnErrorClosed != null)
                    {
                        OnErrorClosed(e);
                    }
                });
            }
        }

        /// <summary>
        ///   RPCを登録
        /// </summary>
        public int RegisterRPC(Action<string, string> rpc)
        {
            return registerRPC(
                rpc,
                (senderId, reader) => rpc(senderId, reader.ReadString()));
        }

        public int RegisterRPC<T>(Action<string, T> rpc, bool cacheObject = false) where T : class, IWSNetSerializable, new()
        {
            if (!cacheObject)
            {
                return registerRPC(
                    rpc,
                    (senderId, reader) => rpc(senderId, reader.ReadObject<T>()));
            }

            T obj = new T();
            return registerRPC(
                rpc,
                (senderId, reader) => {
                    obj = reader.ReadObject(obj);
                    rpc(senderId, obj);
                });
        }

        private int registerRPC(Delegate rpc, Action<string, SerialReader> action)
        {
            var id = rpcActions.Count;

            if (id > byte.MaxValue)
            {
                throw new Exception("RPC map full");
            }

            if (rpcMap.ContainsKey(rpc))
            {
                throw new Exception("RPC target already registered");
            }

            rpcMap[rpc] = (byte)id;
            rpcActions.Add(action);

            return id;
        }

        /// <summary>
        ///   退室メッセージを送信
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     送信だけでは退室は完了しない。
        ///     OnClosedイベントを受け取って退室が完了する。
        ///   </para>
        /// </remarks>
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
            if (!rpcMap.TryGetValue(rpc, out rpcId))
            {
                var msg = $"RPC target is not registered";
                throw new Exception(msg);
            }

            return rpcId;
        }

        internal void handleError(Exception e)
        {
            callbackPool.Add(() =>
            {
                if (OnError != null)
                {
                    OnError(e);
                }
            });
        }

        internal void handleEvent(Event ev)
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
            // callbackが呼ばれて使い終わってから返却.
            // 呼び出し中に例外が飛んでもいいように別callbackで。
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
                    if (OnJoined != null)
                    {
                        OnJoined(Me);
                    }
                });
                return;
            }

            callbackPool.Add(()=>
            {
                var player = new Player(ev.ClientID, ev.GetProps());
                players[player.Id] = player;
                if (OnOtherPlayerJoined != null)
                {
                    OnOtherPlayerJoined(player);
                }
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
                    if (OnMasterPlayerSwitched != null)
                    {
                        OnMasterPlayerSwitched(player, Master);
                    }
                }

                players.Remove(player.Id);
                if (OnOtherPlayerLeft != null)
                {
                    OnOtherPlayerLeft(player);
                }
            });
        }

        /// <summary>
        ///   RPCイベント
        /// </summary>
        private void OnEvRPC(EvRPC ev)
        {
            callbackPool.Add(() =>
            {
                if (ev.RpcID >= rpcActions.Count)
                {
                    var e = new Exception($"RpcID({ev.RpcID}) is not registered");
                    if (OnError != null)
                    {
                        OnError(e);
                    }

                    return;
                }

                var action = rpcActions[ev.RpcID];
                action(ev.SenderID, ev.Reader);
            });
        }

        /// <summary>
        ///   退室イベント
        /// </summary>
        private void OnEvClosed(EvClosed ev)
        {
            callbackPool.Add(() =>
            {
                Closed = true;
                if (OnClosed != null)
                {
                    OnClosed(ev.Description);
                }
            });
        }
    }
}
