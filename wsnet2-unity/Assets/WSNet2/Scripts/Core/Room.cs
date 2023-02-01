using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Security.Cryptography;
using System.Threading.Tasks;
using System.Text;

namespace WSNet2
{
    /// <summary>
    ///   Room
    /// </summary>
    public class Room : PublicRoom
    {
        /// <summary>RoomのMasterをRPCの対象に指定</summary>
        public const string[] RPCToMaster = null;

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

        /// <summary>ルームの非公開プロパティ</summary>
        public IReadOnlyDictionary<string, object> PrivateProps { get => privateProps; }

        /// <summary>マスタークライアント</summary>
        public Player Master { get => players[masterId]; }

        /// <summary>Ping応答時間 (millisec)</summary>
        public ulong RttMillisec { get; private set; }

        /// <summary>全Playerの最終メッセージ受信時刻 (playerId => unixtime millisec)</summary>
        public IReadOnlyDictionary<string, ulong> LastMsgTimestamps { get => lastMsgTimestamps; }

        /// <summary>
        ///   入室イベント通知
        /// </summary>
        /// OnJoined(me)
        public Action<Player> OnJoined;

        /// <summary>
        ///   退室イベント通知
        /// </summary>
        /// OnClosed(message)
        public Action<string> OnClosed;

        /// <summary>
        ///   他のプレイヤーの入室通知
        /// </summary>
        /// OnOtherPlayerJoined(player)
        public Action<Player> OnOtherPlayerJoined;

        /// <summary>
        ///   他のプレイヤーの再入室イベント
        /// </summary>
        public Action<Player> OnOtherPlayerRejoined;

        /// <summary>
        ///   他のプレイヤーの退室通知
        /// </summary>
        /// OnOtherPlayerLeft(player, message)
        public Action<Player, string> OnOtherPlayerLeft;

        /// <summary>
        ///  マスタープレイヤーの変更通知
        /// </summary>
        /// OnMasterPlayerSwitched(previousMaster, newMaster)
        public Action<Player, Player> OnMasterPlayerSwitched;

        /// <summary>
        ///   部屋のプロパティの変更通知
        /// </summary>
        /// OnRoomPropertyChanged(visible, joinable, watchable, searchGroup, maxPlayers, clientDeadline, publicProps, privateProps);
        /// <remarks>
        ///   変更のあったパラメータのみ値が入ります。
        ///   publicProps, privatePropsのキーも変更のあったもののみです。
        /// </remarks>
        public Action<bool?, bool?, bool?, uint?, uint?, uint?, Dictionary<string, object>, Dictionary<string, object>> OnRoomPropertyChanged;

        /// <summary>
        ///   プレイヤーのプロパティの変更通知
        /// </summary>
        /// OnPlayerPropertyChanged(player, props)
        /// <remarks>
        ///   propsには変更のあったキーのみ含まれます。
        /// </remarks>
        public Action<Player, Dictionary<string, object>> OnPlayerPropertyChanged;

        /// <summary>
        ///   Pong受信通知
        /// </summary>
        /// OnPongReceived(rttMillisec, watcherCount, lastMsgTimestamps)
        /// <remarks>
        ///  引数はRoomの RttMillisec, WatcherCount, LastMsgTimestamps と同じ
        /// </remarks>
        public Action<ulong, ulong, IReadOnlyDictionary<string, ulong>> OnPongReceived;

        /// <summary>
        ///   エラー通知
        /// </summary>
        /// OnError(exception)
        public Action<Exception> OnError;

        /// <summary>
        ///   エラーによる切断通知
        /// </summary>
        /// OnErrorClosed(exception)
        public Action<Exception> OnErrorClosed;

        string myId;
        Dictionary<string, object> privateProps;
        Dictionary<string, Player> players;
        string masterId;
        uint clientDeadline;
        Dictionary<string, ulong> lastMsgTimestamps;

        CallbackPool callbackPool;
        Dictionary<Delegate, byte> rpcMap;
        List<Action<string, SerialReader>> rpcActions;
        Dictionary<int, Action<EvResponse>> errorResponseHandler = new Dictionary<int, Action<EvResponse>>();
        Logger logger;

        Connection con;

#if DEBUG
        string[] rpcMethodNames = new string[256];
#endif

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="joined">lobbyからの入室完了レスポンス</param>
        /// <param name="myId">自身のID</param>
        /// <param naem="logger">Logger</param>
        public Room(JoinedRoom joined, string myId, HMAC hmac, Logger logger) : base(joined.roomInfo)
        {
            this.myId = myId;
            this.clientDeadline = joined.deadline;

            logger?.SetRoomInfo(Id, Number);
            this.logger = logger;

            this.con = new Connection(this, myId, hmac, joined, logger);

            this.rpcMap = new Dictionary<Delegate, byte>();
            this.rpcActions = new List<Action<string, SerialReader>>();

            this.Running = true;
            this.Closed = false;

            this.callbackPool = new CallbackPool(() => this.Running);

            var reader = WSNet2Serializer.NewReader(info.privateProps);
            privateProps = reader.ReadDict();

            players = new Dictionary<string, Player>(joined.players.Length);
            lastMsgTimestamps = new Dictionary<string, ulong>(joined.players.Length);
            foreach (var p in joined.players)
            {
                var player = new Player(p);
                players[p.Id] = player;
                lastMsgTimestamps[p.Id] = 0;
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
        ///   WSNet2Client.ProcessCallbackから呼ばれる。
        ///   Unityではメインスレッドで呼ぶようにする。
        /// </remarks>
        internal void ProcessCallback()
        {
            callbackPool.Process();
        }

        /// <summary>
        ///   websocket接続してイベント受信を開始する
        /// </summary>
        internal async Task Start()
        {
            try
            {
                await con.Start();
            }
            catch (Exception e)
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
        ///   強制切断
        /// </summary>
        /// <remarks>
        ///   OnClosedなどは呼ばれない.
        /// </remarks>
        internal void ForceDisconnect()
        {
            con.Cancel();
        }

        /// <summary>
        ///   RPCを登録
        /// </summary>
        public int RegisterRPC(Action<string> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, bool> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadBool()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, sbyte> rpc)
        {
            return registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadSByte()));
        }
        public int RegisterRPC(Action<string, byte> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadByte()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, char> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadChar()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, short> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadShort()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, ushort> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadUShort()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, int> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadInt()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, uint> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadUInt()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, long> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadLong()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, ulong> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadULong()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, float> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadFloat()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, double> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadDouble()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, string> rpc)
        {
            var id = registerRPC(rpc, (senderId, reader) => rpc(senderId, reader.ReadString()));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC<T>(Action<string, T> rpc, T cacheObject = null) where T : class, IWSNet2Serializable, new()
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadObject<T>()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadObject(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, List<object>> rpc, List<object> cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadList()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadList(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, object[]> rpc, object[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadArray()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadArray(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC<T>(Action<string, List<T>> rpc, List<T> cacheObject = null) where T : class, IWSNet2Serializable, new()
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadList<T>()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadList<T>(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC<T>(Action<string, T[]> rpc, T[] cacheObject = null) where T : class, IWSNet2Serializable, new()
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadArray<T>()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadArray<T>(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, Dictionary<string, object>> rpc, Dictionary<string, object> cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadDict()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadDict(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, bool[]> rpc, bool[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadBools()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadBools(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, sbyte[]> rpc, sbyte[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadSBytes()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadSBytes(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, byte[]> rpc, byte[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadBytes()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadBytes(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, char[]> rpc, char[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadChars()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadChars(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, short[]> rpc, short[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadShorts()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadShorts(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, ushort[]> rpc, ushort[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadUShorts()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadUShorts(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, int[]> rpc, int[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadInts()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadInts(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, uint[]> rpc, uint[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadUInts()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadUInts(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, long[]> rpc, long[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadLongs()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadLongs(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, ulong[]> rpc, ulong[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadULongs()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadULongs(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, float[]> rpc, float[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadFloats()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadFloats(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, double[]> rpc, double[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadDoubles()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadDoubles(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
        }
        public int RegisterRPC(Action<string, string[]> rpc, string[] cacheObject = null)
        {
            var id = registerRPC(
                rpc,
                (cacheObject == null)
                ? (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, reader.ReadStrings()))
                : (Action<string, SerialReader>)((senderId, reader) => rpc(senderId, (cacheObject = reader.ReadStrings(cacheObject)))));
            registerRPCMethodName(id, rpc);
            return id;
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
        /// <para>
        ///   CallbackPoolの処理を止めることで部屋の状態の更新と通知を停止する。
        ///   Restart()で再開する。
        /// </para>
        /// <para>
        ///   Unityでシーン遷移するときなど、イベントを受け取れない時間に停止すると良い。
        /// </para>
        public void Pause()
        {
            Running = false;
            logger?.Debug("Room paused");
        }

        /// <summary>
        ///   一時停止したイベント処理を再開する
        /// </summary>
        public void Restart()
        {
            Running = true;
            logger?.Debug("Room restarted");
        }

        /// <summary>
        ///   退室メッセージを送信
        /// </summary>
        /// <param name="message">切断理由など（UTF-8で123byteまで）</param>
        /// <remarks>
        ///   送信だけでは退室は完了しない。
        ///   OnClosedイベントを受け取って退室が完了する。
        ///   messageはOnClosedやOnOtherPlayerLeftの引数となる。
        /// </remarks>
        public int Leave(string message = "")
        {
            if (Encoding.UTF8.GetByteCount(message) > 123)
            {
                throw new Exception("message too long");
            }

            return con.msgPool.PostLeave(message);
        }

        /// <summary>
        ///   Masterを移譲する
        /// </summary>
        /// <param name="newMaster">新Master</param>
        /// <param name="onErrorResponse">サーバ側でエラーになったときのコールバック</param>
        /// <remarks>
        ///   この操作はMasterのみ呼び出せる。
        ///   実際の切り替えは、OnMasterPlayerSwitchedが呼び出されるタイミングで行われる。
        /// </remarks>
        public int SwitchMaster(Player newMaster, Action<EvType, string> onErrorResponse = null)
        {
            if (Me != Master)
            {
                throw new Exception("SwitchMaster is for master only");
            }

            if (!players.ContainsKey(newMaster.Id))
            {
                throw new Exception($"Player \"{newMaster.Id}\" is not in this room");
            }

            var seqNum = con.msgPool.PostSwitchMaster(newMaster.Id);

            if (onErrorResponse != null)
            {
                errorResponseHandler[seqNum] = (ev) =>
                {
                    onErrorResponse(ev.Type, ev.GetSwitchMasterPayload());
                };
            }

            return seqNum;
        }

        /// <summary>
        ///   Roomプロパティを変更する
        /// </summary>
        /// <param name="visible">検索可能</param>
        /// <param name="joinable">入室可能</param>
        /// <param name="watchable">観戦可能</param>
        /// <param name="searchGroup">検索グループ</param>
        /// <param name="maxPlayers">最大人数</param>
        /// <param name="clientDeadline">通信タイムアウト時間(秒)</param>
        /// <param name="publicProps">公開プロパティ（変更するキーのみ）</param>
        /// <param name="privateProps">非公開プロパティ（変更するキーのみ）</param>
        /// <param name="onErrorResponse">サーバ側でエラーになったときのコールバック</param>
        /// <remarks>
        ///   この操作はMasterのみ呼び出せる。
        ///   実際の変更は、OnRoomPropertyChangedが呼び出されるタイミングで行われる。
        /// </remarks>
        /// <example>
        ///   <code>
        ///   // 変更する項目のみ引数に値をセットする
        ///   // Props辞書も同様、変更するキーのみを含める。
        ///   room.ChangeRoomProperty(
        ///       joinable: false, watchable: true, searchGroup: GameState.Playing,
        ///       publicProps: new Dictionary&lt;string, object&gt;(){ {"score", "0 - 0"} });
        ///   </code>
        /// </example>
        public int ChangeRoomProperty(
            bool? visible = null,
            bool? joinable = null,
            bool? watchable = null,
            uint? searchGroup = null,
            uint? maxPlayers = null,
            uint? clientDeadline = null,
            IDictionary<string, object> publicProps = null,
            IDictionary<string, object> privateProps = null,
            Action<EvType, bool?, bool?, bool?, uint?, uint?, uint?, IDictionary<string, object>, IDictionary<string, object>> onErrorResponse = null)
        {
            if (Me != Master)
            {
                throw new Exception("ChangeRoomProperty is for master only");
            }

            var seqNum = con.msgPool.PostRoomProp(
                visible ?? Visible,
                joinable ?? Joinable,
                watchable ?? Watchable,
                searchGroup ?? SearchGroup,
                (ushort)(maxPlayers ?? MaxPlayers),
                (ushort)(clientDeadline ?? 0),
                publicProps ?? null,
                privateProps ?? null);

            if (onErrorResponse != null)
            {
                errorResponseHandler[seqNum] = (ev) =>
                {
                    var payload = ev.GetRoomPropPayload();
                    onErrorResponse(
                        ev.Type,
                        visible,
                        joinable,
                        watchable,
                        searchGroup,
                        maxPlayers,
                        clientDeadline,
                        payload.PublicProps,
                        payload.PrivateProps);
                };
            }

            return seqNum;
        }

        /// <summary>
        ///   自分自身のプロパティを変更する
        /// </summary>
        /// <param name="props">変更するプロパティの辞書（変更するキーのみ）</param>
        /// <param name="onErrorResponse">サーバ側でエラーになったときのコールバック</param>
        /// <remarks>
        ///   この操作はプレイヤーのみ呼び出せる。観戦者はできない。
        ///   実際の変更は、OnPlayerPropertyChangedが呼び出されるタイミングで行われる。
        /// </remarks>
        public int ChangeMyProperty(IDictionary<string, object> props, Action<EvType, IDictionary<string, object>> onErrorResponse = null)
        {
            var seqNum = con.msgPool.PostClientProp(props);

            if (onErrorResponse != null)
            {
                errorResponseHandler[seqNum] = (ev) =>
                {
                    onErrorResponse(ev.Type, ev.GetClientPropPayload());
                };
            }

            return seqNum;
        }

        /// <summary>
        ///   対象のプレイヤーを強制退室させる
        /// </summary>
        /// <param name="target">対象プレイヤー</param>
        /// <param name="onErrorResponse">サーバ側でエラーになったときのコールバック</param>
        /// <remarks>
        ///   この操作はMasterのみ呼び出せる。
        /// </remarks>
        public int Kick(Player target, Action<EvType, string> onErrorResponse = null)
        {
            return Kick(target, "", onErrorResponse);
        }

        /// <summary>
        ///   対象のプレイヤーを強制退室させる
        /// </summary>
        /// <param name="target">対象プレイヤー</param>
        /// <param name="message">メッセージ</param>
        /// <param name="onErrorResponse">サーバ側でエラーになったときのコールバック</param>
        /// <remarks>
        ///   この操作はMasterのみ呼び出せる。
        /// </remarks>
        public int Kick(Player target, string message, Action<EvType, string> onErrorResponse = null)
        {
            if (Me != Master)
            {
                throw new Exception("Kick is for master only");
            }

            if (!players.ContainsKey(target.Id))
            {
                throw new Exception($"Player \"{target.Id}\" is not in this room");
            }

            var seqNum = con.msgPool.PostKick(target.Id, message);

            if (onErrorResponse != null)
            {
                errorResponseHandler[seqNum] = (ev) =>
                {
                    onErrorResponse(ev.Type, ev.GetKickPayload());
                };
            }

            return seqNum;
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
        public int RPC(Action<string, char> rpc, char param, params string[] targets)
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
        public int RPC<T>(Action<string, T> rpc, T param, params string[] targets) where T : class, IWSNet2Serializable
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
        public int RPC<T>(Action<string, List<T>> rpc, List<T> param, params string[] targets) where T : class, IWSNet2Serializable
        {
            return con.msgPool.PostRPC(getRpcId(rpc), param, targets);
        }
        public int RPC<T>(Action<string, T[]> rpc, T[] param, params string[] targets) where T : class, IWSNet2Serializable
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
        public int RPC(Action<string, char[]> rpc, char[] param, params string[] targets)
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
                case EvRejoined evRejoined:
                    OnEvRejoined(evRejoined);
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
                case EvResponse evResponse:
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
                ev.GetLastMsgTimestamps(lastMsgTimestamps);
                if (OnPongReceived != null)
                {
                    OnPongReceived(RttMillisec, info.watchers, lastMsgTimestamps);
                }
            });
        }

        /// <summary>
        ///   入室イベント
        /// </summary>
        private void OnEvJoined(EvJoined ev)
        {
            if (ev.ClientID == myId)
            {
                logger?.Info("joined: me");
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

            logger?.Info("joined: {0}", ev.ClientID);
            callbackPool.Add(() =>
            {
                var player = new Player(ev.ClientID, ev.GetProps());
                players[player.Id] = player;
                lastMsgTimestamps[player.Id] = 0;
                info.players = (uint)players.Count;
                if (OnOtherPlayerJoined != null)
                {
                    OnOtherPlayerJoined(player);
                }
            });
        }

        /// <summary>
        ///   再入室イベント
        /// </summary>
        private void OnEvRejoined(EvRejoined ev)
        {
            if (ev.ClientID == myId)
            {
                logger?.Info("rejoined: me");
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

            logger?.Info("rejoined: {0}", ev.ClientID);
            callbackPool.Add(() =>
            {
                var player = players[ev.ClientID];
                player.Props = ev.GetProps(player.Props);
                if (OnOtherPlayerRejoined != null)
                {
                    OnOtherPlayerRejoined(player);
                }
            });
        }

        /// <summary>
        ///   プレイヤー退室イベント
        /// </summary>
        private void OnEvLeft(EvLeft ev)
        {
            logger?.Info("left: {0}: {1}", ev.ClientID, ev.Message);

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
                lastMsgTimestamps.Remove(player.Id);
                info.players = (uint)players.Count;
                if (OnOtherPlayerLeft != null)
                {
                    OnOtherPlayerLeft(player, ev.Message);
                }
            });
        }

        /// <summary>
        ///   Roomプロパティ変更イベント
        /// </summary>
        private void OnEvRoomProp(EvRoomProp ev)
        {
            logger?.Debug("room property changed");

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
            logger?.Debug("player property changed: {0}", ev.ClientID);

            callbackPool.Add(() =>
            {
                var player = players[ev.ClientID];
                var props = ev.GetProps(player.Props);
                foreach (var kv in props)
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
            logger?.Info("master switched: {0}", ev.NewMasterId);

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
            logger?.Debug("RPC: {0}", ev.RpcID);

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
            logger?.Info("closed: {0}", ev.Description);

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
        private void OnEvResponse(EvResponse ev)
        {
            callbackPool.Add(() =>
            {
                var seqNum = ev.MsgSeqNum;
                Action<EvResponse> handler;
                if (errorResponseHandler.TryGetValue(seqNum, out handler))
                {
                    errorResponseHandler.Remove(seqNum);
                    if (ev.Type != EvType.Succeeded)
                    {
                        handler(ev);
                    }
                }
            });
        }

        [Conditional("DEBUG")]
        void registerRPCMethodName(int rpcId, Delegate rpc)
        {
#if DEBUG
            rpcMethodNames[rpcId] = $"{rpc.Method.DeclaringType}.{rpc.Method.Name}";
#endif
        }

#if DEBUG
        /// <summary>
        ///   RPCのメソッド名取得（NetworkInformer用）
        /// </summary>
        public string MethodName(int rpcId)
        {
            return rpcMethodNames[rpcId];
        }
#endif
    }
}
