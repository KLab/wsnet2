using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using MessagePack;

namespace WSNet2.Core
{
    /// <summary>
    ///   WSNet2に接続するためのClient
    /// </summary>
    public class WSNet2Client
    {
        string baseUri;
        string appId;
        string userId;
        string bearer;

        List<Room> rooms = new List<Room>();
        CallbackPool callbackPool = new CallbackPool();

        Logger logger;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="baseUri">LobbyのURI</param>
        /// <param name="appId">Wsnetに登録してあるApplication ID</param>
        /// <param name="userId">プレイヤーIDとなるID</param>
        /// <param name="authData">認証情報（アプリAPIサーバから入手）</param>
        /// <param name="logger">Logger</param>
        public WSNet2Client(string baseUri, string appId, string userId, string authData, IWSNet2Logger<WSNet2LogPayload> logger=null)
        {
            this.appId = appId;
            this.userId = userId;
            this.SetBaseUri(baseUri);
            this.UpdateAuthData(authData);
            this.logger = prepareLogger(logger);
            checkMinThreads();
        }

        /// <summary>
        ///   接続情報を更新
        /// </summary>
        /// <param name="baseUri">LobbyのURIを変更する</param>
        public void SetBaseUri(string baseUri)
        {
            this.baseUri = baseUri;
        }

        /// <summary>
        ///   認証データを更新
        /// </summary>
        /// <param name="authData">認証情報</param>
        public void UpdateAuthData(string authData)
        {
            this.bearer = "Bearer " + authData;
        }

        /// <summary>
        ///   蓄積されたCallbackを処理する。
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     Unityではcallbackをメインスレッドで動かしたいので溜めておいて
        ///     このメソッド経由で実行する。Update()などで呼び出せば良い。
        ///     DotNetの場合は適当なスレッドでループを回す。
        ///   </para>
        /// </remarks>
        public void ProcessCallback()
        {
            callbackPool.Process();
            lock(rooms)
            {
                for (var i = rooms.Count-1; i >= 0; i--)
                {
                    rooms[i].ProcessCallback();
                    if (rooms[i].Closed)
                    {
                        rooms.RemoveAt(i);
                    }
                }
            }
        }

        /// <summary>
        ///   すべての部屋から強制切断する
        /// </summary>
        public void ForceDisconnect()
        {
            foreach (var room in rooms)
            {
                room.ForceDisconnect();
            }
        }

        /// <summary>
        ///   部屋を作成して入室
        /// </summary>
        /// <param name="roomOption">部屋オプション</param>
        /// <param name="clientProps">自身のカスタムプロパティ</param>
        /// <param name="onSuccess">成功時callback. 引数は作成した部屋</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        /// <param name="roomLogger">部屋用のLogger(optional)。Loggerを変更する場合設定する</param>
        /// <remarks>
        ///   <para>callbackはProcessCallback経由で呼ばれる</para>
        ///   <para>
        ///     onSuccessが呼ばれた時点ではまだwebsocket接続していない。
        ///     ここでRoom.Pause()することで、イベントが処理されるのを止めておける（Room.On*やRPCが呼ばれない）。
        ///     その間ProcessCallback()は呼び続けて良い。
        ///     Room.Restart()するとイベント処理を再開する。
        ///   </para>
        ///   <para>
        ///     たとえば、onSuccessでPauseしてシーン遷移し、
        ///     遷移後のシーンでOn*やRPCを登録後にRestartするという使い方を想定している。
        ///   </para>
        /// </remarks>
        /// <remarks>
        ///   <para>
        ///     複数の部屋に同時に入室するときはPayloadの上書きを避けるためLoggerを別のインスタンスにする。
        ///   </para>
        /// </remarks>
        public void Create(
            RoomOption roomOption,
            IDictionary<string, object> clientProps,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            IWSNet2Logger<WSNet2LogPayload> roomLogger = null)
        {
            var logger = prepareLogger(roomLogger);
            logger?.Debug("WSNet2Client.Create()");

            var param = new CreateParam();
            param.roomOption = roomOption;
            param.clientInfo = new ClientInfo(userId, clientProps);

            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom("/rooms", content, onSuccess, onFailed, logger));
        }

        /// <summary>
        ///   部屋IDを指定して入室
        /// </summary>
        /// <param name="roomId">Room ID</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="clientProps">自身のカスタムプロパティ</param>
        /// <param name="onSuccess">成功時callback. 引数は入室した部屋</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void Join(
            string roomId,
            Query query,
            IDictionary<string, object> clientProps,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            IWSNet2Logger<WSNet2LogPayload> roomLogger = null)
        {
            var logger = prepareLogger(roomLogger);
            logger?.Debug("WSNet2Client.Join(roomId={0})", roomId);

            var param = new JoinParam(){
                queries = query?.condsList,
                clientInfo = new ClientInfo(userId, clientProps),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/id/{roomId}", content, onSuccess, onFailed, logger));
        }

        /// <summary>
        ///   部屋番号を指定して入室
        /// </summary>
        /// <param name="number">部屋番号</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="clientProps">自身のカスタムプロパティ</param>
        /// <param name="onSuccess">成功時callback. 引数は入室した部屋</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void Join(
            int number,
            Query query,
            IDictionary<string, object> clientProps,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            IWSNet2Logger<WSNet2LogPayload> roomLogger = null)
        {
            var logger = prepareLogger(roomLogger);
            logger?.Debug("WSNet2Client.Join(number={0})", number);

            var param = new JoinParam(){
                queries = query?.condsList,
                clientInfo = new ClientInfo(userId, clientProps),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/number/{number}", content, onSuccess, onFailed, logger));
        }

        /// <summary>
        ///   検索クエリに合致する部屋にランダム入室
        /// </summary>
        /// <param name="group">検索グループ</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="clientProps">自身のカスタムプロパティ</param>
        /// <param name="onSuccess">成功時callback. 引数は入室した部屋</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void RandomJoin(
            uint group,
            Query query,
            IDictionary<string, object> clientProps,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            IWSNet2Logger<WSNet2LogPayload> roomLogger = null)
        {
            var logger = prepareLogger(roomLogger);
            logger?.Debug("WSNet2Client.RandomJoin(group={0})", group);

            var param = new JoinParam(){
                queries = query?.condsList,
                clientInfo = new ClientInfo(userId, clientProps),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/random/{group}", content, onSuccess, onFailed, logger));
        }

        /// <summary>
        ///   RoomIDを指定して観戦入室
        /// </summary>
        /// <param name="roomId">Room ID</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="onSuccess">成功時callback. 引数は入室した部屋</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void Watch(
            string roomId,
            Query query,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            IWSNet2Logger<WSNet2LogPayload> roomLogger = null)
        {
            var logger = prepareLogger(roomLogger);
            logger?.Debug("WSNet2Client.Watch(roomId={0})", roomId);

            var param = new JoinParam(){
                queries = query?.condsList,
                clientInfo = new ClientInfo(userId),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/watch/id/{roomId}", content, onSuccess, onFailed, logger));
        }

        /// <summary>
        ///   部屋番号を指定して観戦入室
        /// </summary>
        /// <param name="number">部屋番号</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="onSuccess">成功時callback. 引数は入室した部屋</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void Watch(
            int number,
            Query query,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            IWSNet2Logger<WSNet2LogPayload> roomLogger = null)
        {
            var logger = prepareLogger(roomLogger);
            logger?.Debug("WSNet2Client.Watch(number={0})", number);

            var param = new JoinParam(){
                queries = query?.condsList,
                clientInfo = new ClientInfo(userId),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/watch/number/{number}", content, onSuccess, onFailed, logger));
        }

        /// <summary>
        ///   部屋検索
        /// </summary>
        /// <param name="group">検索グループ</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="limit">件数上限</param>
        /// <param name="onSuccess">成功時callback. 引数は検索でヒットした部屋一覧</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void Search(
            uint group,
            Query query,
            int limit,
            Action<PublicRoom[]> onSuccess,
            Action<Exception> onFailed)
        {
            Search(group, query, limit, false, false, onSuccess, onFailed);
        }

        /// <summary>
        ///   部屋検索
        /// </summary>
        /// <param name="group">検索グループ</param>
        /// <param name="query">検索クエリ</param>
        /// <param name="limit">件数上限</param>
        /// <param name="checkJoinable">入室可能な部屋のみ含める</param>
        /// <param name="checkWatchable">観戦可能な部屋のみ含める</param>
        /// <param name="onSuccess">成功時callback. 引数は検索でヒットした部屋一覧</param>
        /// <param name="onFailed">失敗時callback. 引数は例外オブジェクト</param>
        public void Search(
            uint group,
            Query query,
            int limit,
            bool checkJoinable,
            bool checkWatchable,
            Action<PublicRoom[]> onSuccess,
            Action<Exception> onFailed)
        {
            logger?.Debug("WSNet2Client.Search(group={0})", group);

            var param = new SearchParam(){
                group = group,
                queries = query?.condsList,
                limit = limit,
                checkJoinable = checkJoinable,
                checkWatchable = checkWatchable,
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => search(content, onSuccess, onFailed));
        }

        private async Task<LobbyResponse> post(string path, byte[] content)
        {
            var cli = new HttpClient();
            cli.DefaultRequestHeaders.Add("Wsnet2-App", appId);
            cli.DefaultRequestHeaders.Add("Wsnet2-User", userId);
            cli.DefaultRequestHeaders.Add("Authorization", bearer);

            var res = await cli.PostAsync(baseUri + path, new ByteArrayContent(content));
            var body = await res.Content.ReadAsByteArrayAsync();
            if (!res.IsSuccessStatusCode)
            {
                var msg = System.Text.Encoding.UTF8.GetString(body);
                throw new Exception($"wsnet2 {path} failed: code={res.StatusCode} {msg}");
            }

            return MessagePackSerializer.Deserialize<LobbyResponse>(body);
        }

        private async Task connectToRoom(
            string path,
            byte[] content,
            Action<Room> onSuccess,
            Action<Exception> onFailed,
            Logger logger)
        {
            try
            {
                var res = await post(path, content);
                if (res.room == null)
                {
                    throw new RoomNotFoundException(res.msg);
                }

                var room = new Room(res.room, userId, logger);
                logger?.Info("Joined to room: {0}", room.Id);

                callbackPool.Add(() =>
                {
                    onSuccess(room);
                    lock(rooms)
                    {
                        rooms.Add(room);
                    }
                    Task.Run(room.Start);
                });
            }
            catch (Exception e)
            {
                logger?.Error(e, "Failed to connect to room");
                callbackPool.Add(() => onFailed(e));
            }
        }

        private async Task search(
            byte[] content,
            Action<PublicRoom[]> onSuccess,
            Action<Exception> onFailed)
        {
            try
            {
                var res = await post("/rooms/search", content);
                var count = res.rooms?.Length ?? 0;
                var rooms = new PublicRoom[count];
                for (var i=0; i<count; i++)
                {
                    rooms[i] = new PublicRoom(res.rooms[i]);
                }

                callbackPool.Add(() => onSuccess(rooms));
            }
            catch (Exception e)
            {
                logger?.Error(e, "Failed to search");
                callbackPool.Add(() => onFailed(e));
            }
        }

        private void checkMinThreads()
        {
            const int thNumThreads = 10;
            int workerThreads, completionPortThreads;
            System.Threading.ThreadPool.GetMinThreads(out workerThreads, out completionPortThreads);

            if (workerThreads < thNumThreads || completionPortThreads < thNumThreads)
            {
                logger?.Warning(@"The ThreadPool may become congested. Please consider increasing the ThreadPool.MinThreads values.
c.f. https://gist.github.com/JonCole/e65411214030f0d823cb
WorkerThreads: {0}, CompletionPortThreads: {1}", workerThreads, completionPortThreads);
            }
        }

        private Logger prepareLogger(IWSNet2Logger<WSNet2LogPayload> logger)
        {
            if (logger == null)
            {
                return this.logger;
            }

            logger.Payload.AppId = appId;
            logger.Payload.UserId = userId;
            return new Logger(logger);
        }
    }
}
