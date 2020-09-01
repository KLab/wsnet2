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
        public string Id { get => info.id; }

        /// <summary>部屋番号</summary>
        public int Number { get => info.number; }

        /// <summary>検索可能</summary>
        public bool Visible { get => info.visible; }

        /// <summary>入室可能</summary>
        public bool Joinable { get => info.joinable; }

        /// <summary>観戦可能</summary>
        public bool Watchable { get => info.watchable; }

        /// <summary>検索グループ</summary>
        public uint SearchGroup { get => info.searchGroup; }

        /// <summary>検索グループ</summary>
        public uint MaxPlayers { get => info.maxPlayers; }

        /// <summary>通信タイムアウト時間(秒)
        public uint ClientDeadline { get => clientDeadline; }

        /// <summary>Callbackループの動作状態</summary>
        public bool Running { get; private set; }

        /// <summary>終了したかどうか</summary>
        public bool Closed { get; private set; }

        /// <summary>自分自身のPlayer</summary>
        public Player Me { get; private set; }

        /// <summary>部屋内の全Player</summary>
        public IReadOnlyDictionary<string, Player> Players { get => players; }

        /// <summary>ルームの公開プロパティ</summary>
        public IReadOnlyDictionary<string, object> PublicProps { get => publicProps; }

        /// <summary>ルームの非公開プロパティ</summary>
        public IReadOnlyDictionary<string, object> PrivateProps { get => privateProps; }

        /// <summary>マスタークライアント</summary>
        public Player Master { get => players[masterId]; }

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

        /// <summary>マスタープレイヤーの変更通知</summary>
        public Action<Player, Player> OnMasterPlayerSwitched;

        /// <summary>部屋のプロパティの変更通知</summary>
        /// OnRoomPropertyChanged(visible, joinable, watchable, searchGroup, maxPlayers, clientDeadline, publicProps, privateProps);
        /// <remarks>
        ///   <para>
        ///     変更のあったパラメータのみ値が入ります。
        ///   </para>
        /// </remarks>
        public Action<bool?, bool?, bool?, uint?, uint?, uint?, Dictionary<string, object>, Dictionary<string, object>> OnRoomPropertyChanged;

        /// <summary>プレイヤーのプロパティの変更通知</summary>
        public Action<Player, Dictionary<string, object>> OnPlayerPropertyChanged;

        /// <summary>Msgの処理エラー</summary>
        /// EvMsgErrorの派生型によってエラー種別を判定します。
        /// - EvPermissionDeny: 権限エラー
        /// - EvTargetNotFound: ターゲット不在
        //public Action<EvMsgError> OnMsgError;

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
        uint clientDeadline;

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
            this.clientDeadline = joined.deadline;

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
        internal void ProcessCallback()
        {
            if (Running)
            {
                callbackPool.Process();
            }
        }

        /// <summary>
        ///   websocket接続してイベント受信を開始する
        /// </summary>
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
        public int RegisterRPC(Action<string> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc (senderId));
        }
        public int RegisterRPC(Action<string, bool> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadBool()));
        }
        public int RegisterRPC(Action<string, sbyte> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadSByte()));
        }
        public int RegisterRPC(Action<string, byte> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadByte()));
        }
        public int RegisterRPC(Action<string, short> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadShort()));
        }
        public int RegisterRPC(Action<string, ushort> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadUShort()));
        }
        public int RegisterRPC(Action<string, int> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadInt()));
        }
        public int RegisterRPC(Action<string, uint> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadUInt()));
        }
        public int RegisterRPC(Action<string, long> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadLong()));
        }
        public int RegisterRPC(Action<string, ulong> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadULong()));
        }
        public int RegisterRPC(Action<string, float> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadFloat()));
        }
        public int RegisterRPC(Action<string, double> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadDouble()));
        }
        public int RegisterRPC(Action<string, string> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadString()));
        }
        public int RegisterRPC<T>(Action<string, T> rpc, T cacheObject = null) where T : class, IWSNetSerializable, new()
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadObject<T>()));
            }

            return registerRPC(rpc, (senderId, reader) => rpc(senderId, (cacheObject = reader.ReadObject(cacheObject))));
        }
        public int RegisterRPC(Action<string, List<object>> rpc, List<object> cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadList()));
            }

            return registerRPC(
                rpc,
                (senderId, reader) => rpc(senderId, (cacheObject = reader.ReadList(cacheObject))));
        }
        public int RegisterRPC(Action<string, object[]> rpc, object[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadArray()));
            }

            return registerRPC(
                rpc,
                (senderId, reader) => rpc(senderId, (cacheObject = reader.ReadArray(cacheObject))));
        }
        public int RegisterRPC<T>(Action<string, List<T>> rpc, List<T> cacheObject = null) where T : class, IWSNetSerializable, new()
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadList<T>()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadList<T>(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC<T>(Action<string, T[]> rpc, T[] cacheObject = null) where T : class, IWSNetSerializable, new()
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadArray<T>()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadArray<T>(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, Dictionary<string, object>> rpc, Dictionary<string, object> cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadDict()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadDict(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, bool[]> rpc, bool[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadBools()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadBools(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, sbyte[]> rpc, sbyte[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadSBytes()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadSBytes(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, byte[]> rpc, byte[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadBytes()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadBytes(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, short[]> rpc, short[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadShorts()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadShorts(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, ushort[]> rpc, ushort[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadUShorts()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadUShorts(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, int[]> rpc, int[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadInts()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadInts(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, uint[]> rpc, uint[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadUInts()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadUInts(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, long[]> rpc, long[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadLongs()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadLongs(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, ulong[]> rpc, ulong[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadULongs()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadULongs(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, float[]> rpc, float[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadFloats()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadFloats(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, double[]> rpc, double[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadDoubles()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadDoubles(cacheObject);
                rpc(senderId, cacheObject);
            });
        }
        public int RegisterRPC(Action<string, string[]> rpc, string[] cacheObject = null)
        {
            if (cacheObject == null)
            {
                return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadStrings()));
            }

            return registerRPC(rpc, (senderId, reader) =>
            {
                cacheObject = reader.ReadStrings(cacheObject);
                rpc(senderId, cacheObject);
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
        ///   イベント処理を一時停止する
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     CallbackPoolの処理を止めることで部屋の状態の更新と通知を停止する。
        ///     Restart()で再開する。
        ///   </para>
        ///   <para>
        ///     Unityでシーン遷移するときなど、イベントを受け取れない時間に停止すると良い。
        ///   </para>
        /// </remarks>
        public void Pause()
        {
            Running = false;
        }

        /// <summary>
        ///   一時停止したイベント処理を再開する
        /// </summary>
        public void Restart()
        {
            Running = true;
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
        public int Leave()
        {
            return con.msgPool.PostLeave();
        }

        /// <summary>
        ///   Masterを移譲する
        /// </summary>
        /// <param name="newMaster">新Master</param>
        public int SwitchMaster(Player newMaster)
        {
            if (Me != Master)
            {
                throw new Exception("SwitchMaster is for master only");
            }

            if (!players.ContainsKey(newMaster.Id))
            {
                throw new Exception($"Player \"{newMaster.Id}\" is not in this room");
            }

            return con.msgPool.PostSwitchMaster(newMaster.Id);
        }

        /// <summary>
        ///   Roomプロパティを変更する
        /// </summary>
        public int ChangeRoomProperty(
            bool? visible = null,
            bool? joinable = null,
            bool? watchable = null,
            uint? searchGroup = null,
            uint? maxPlayers = null,
            uint? clientDeadline = null,
            IDictionary<string, object> publicProps = null,
            IDictionary<string, object> privateProps = null)
        {
            if (Me != Master)
            {
                throw new Exception("ChangeRoomProperty is for master only");
            }

            return con.msgPool.PostRoomProp(
                visible ?? Visible,
                joinable ?? Joinable,
                watchable ?? Watchable,
                searchGroup ?? SearchGroup,
                (ushort)(maxPlayers ?? MaxPlayers),
                (ushort)(clientDeadline ?? 0),
                publicProps ?? null,
                privateProps ?? null);
        }

        /// <summary>
        ///   自分自身のプロパティを変更する
        /// </summary>
        /// <param name="props">変更するプロパティの辞書</param>
        public int ChangeMyProperty(IDictionary<string, object> props)
        {
            return con.msgPool.PostClientProp(props);
        }

        /// <summary>
        ///   RPC呼び出し
        /// </summary>
        public int RPC(Action<string> rpc, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), targets);
        }
        public int RPC(Action<string, bool> rpc, bool param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, sbyte> rpc, sbyte param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, byte> rpc, byte param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, short> rpc, short param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, ushort> rpc, ushort param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, int> rpc, int param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, uint> rpc, uint param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, long> rpc, long param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, ulong> rpc, ulong param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, float> rpc, float param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, double> rpc, double param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, string> rpc, string param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC<T>(Action<string, T> rpc, T param, params string[] targets) where T : class, IWSNetSerializable
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, List<object>> rpc, List<object> param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, object[]> rpc, object[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC<T>(Action<string, List<T>> rpc, List<T> param, params string[] targets) where T : class, IWSNetSerializable
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC<T>(Action<string, T[]> rpc, T[] param, params string[] targets) where T : class, IWSNetSerializable
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, Dictionary<string, object>> rpc, Dictionary<string, object> param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, bool[]> rpc, bool[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, sbyte[]> rpc, sbyte[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, byte[]> rpc, byte[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, short[]> rpc, short[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, ushort[]> rpc, ushort[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, int[]> rpc, int[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, uint[]> rpc, uint[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, long[]> rpc, long[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, ulong[]> rpc, ulong[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, float[]> rpc, float[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, double[]> rpc, double[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC(Action<string, string[]> rpc, string[] param, params string[] targets)
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
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
                case EvPong evPong:
                    OnEvPong(evPong);
                    break;
                case EvJoined evJoined:
                    OnEvJoined(evJoined);
                    break;
                case EvLeft evLeft:
                    OnEvLeft(evLeft);
                    break;
                case EvRoomProp evRoomProp:
                    OnEvRoomProp(evRoomProp);
                    break;
                case EvClientProp evClientProp:
                    OnEvClientProp(evClientProp);
                    break;
                case EvMasterSwitched evMasterSwitched:
                    OnEvMasterSwitched(evMasterSwitched);
                    break;
                case EvRPC evRpc:
                    OnEvRPC(evRpc);
                    break;
                case EvClosed evClosed:
                    OnEvClosed(evClosed);
                    break;
                case IEvResponse evResponse:
                    OnEvResponse(evResponse);
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
        ///   Roomプロパティ変更イベント
        /// </summary>
        private void OnEvRoomProp(EvRoomProp ev)
        {
            if (ev.ClientDeadline > 0)
            {
                // ping間隔はすぐに変更しないとTimeoutする可能性がある
                con.UpdatePingInterval(ev.ClientDeadline);
            }

            callbackPool.Add(() =>
            {
                bool? visible = null;
                bool? joinable = null;
                bool? watchable = null;
                uint? searchGroup = null;
                uint? maxPlayers = null;
                uint? clientDeadline = null;
                Dictionary<string, object> publicProps = null;
                Dictionary<string, object> privateProps = null;

                if (info.visible != ev.Visible)
                {
                    visible = info.visible = ev.Visible;
                }

                if (info.joinable != ev.Joinable)
                {
                    joinable = info.joinable = ev.Joinable;
                }

                if (info.watchable != ev.Watchable)
                {
                    watchable = info.watchable = ev.Watchable;
                }

                if (info.searchGroup != ev.SearchGroup)
                {
                    searchGroup = info.searchGroup = ev.SearchGroup;
                }

                if (info.maxPlayers != ev.MaxPlayers)
                {
                    maxPlayers = info.maxPlayers = ev.MaxPlayers;
                }

                if (this.clientDeadline != ev.ClientDeadline)
                {
                    clientDeadline = this.clientDeadline = ev.ClientDeadline;
                }

                var props = ev.GetPublicProps(this.publicProps);
                if (props != null)
                {
                    publicProps = props;
                    foreach (var kv in props)
                    {
                        this.publicProps[kv.Key] = kv.Value;
                    }
                }

                props = ev.GetPrivateProps(this.privateProps);
                if (props != null)
                {
                    privateProps = props;
                    foreach (var kv in props)
                    {
                        this.privateProps[kv.Key] = kv.Value;
                    }
                }

                if (OnRoomPropertyChanged != null)
                {
                    OnRoomPropertyChanged(
                        visible, joinable, watchable,
                        searchGroup, maxPlayers, clientDeadline,
                        publicProps, privateProps);
                }
            });
        }

        /// <summary>
        ///   プレイヤープロパティ変更イベント
        /// </summary>
        private void OnEvClientProp(EvClientProp ev)
        {
            callbackPool.Add(() =>
            {
                var player = players[ev.ClientID];
                var props = ev.GetProps(player.Props);
                foreach(var kv in props)
                {
                    player.Props[kv.Key] = kv.Value;
                }

                if (OnPlayerPropertyChanged != null)
                {
                    OnPlayerPropertyChanged(player, props);
                }
            });
        }

        /// <summary>
        ///   マスタープレイヤー交代イベント
        /// </summary>
        private void OnEvMasterSwitched(EvMasterSwitched ev)
        {
            callbackPool.Add(() =>
            {
                var prev = Master;
                masterId = ev.NewMasterId;
                if (OnMasterPlayerSwitched != null)
                {
                    OnMasterPlayerSwitched(prev, Master);
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

        /// <summary>
        ///   Msg失敗通知
        /// </summary>
        private void OnEvResponse(IEvResponse ev)
        {
        }
    }
}
