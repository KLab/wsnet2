using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using MessagePack;

namespace WSNet2.Core
{
    public class AuthData
    {
        public string Timestamp { get; set; }
        public string Nonce { get; set; }
        public string Hash { get; set; }
    }

    /// <summary>
    ///   WSNet2に接続するためのClient
    /// </summary>
    public class WSNet2Client
    {
        string baseUri;
        string appId;
        string userId;
        AuthData authData;

        List<Room> rooms = new List<Room>();
        CallbackPool callbackPool = new CallbackPool();

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="baseUri">LobbyのURI</param>
        /// <param name="appId">Wsnetに登録してあるApplication ID</param>
        /// <param name="userId">プレイヤーIDとなるID</param>
        /// <param name="authData">認証情報（アプリAPIサーバから入手）</param>
        public WSNet2Client(string baseUri, string appId, string userId, AuthData authData)
        {
            this.baseUri = baseUri;
            this.appId = appId;
            this.userId = userId;
            this.authData = authData;
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
        ///   部屋を作成して入室
        /// </summary>
        /// <param name="roomOption">部屋オプション</param>
        /// <param name="clientProps">自身のカスタムプロパティ</param>
        /// <param name="receiver">イベントレシーバ</param>
        /// <param name="onSuccess">成功時callback</param>
        /// <param name="onFailed">失敗時callback</param>
        /// <remarks>
        ///   <para>callbackはProcessCallback経由で呼ばれる</para>
        ///   <para>
        ///     onSuccessが呼ばれた時点ではまだwebsocket接続していない。
        ///     ここでRoom.Running=falseすることで、イベントが処理されるのを止めておける。
        ///     ProcessCallback()は呼び続けて良い。
        ///     シーン遷移後にRoom.Running=trueにするとイベントが処理されレシーバに届くようになる。
        ///   </para>
        /// </remarks>
        public void Create(
            RoomOption roomOption,
            IDictionary<string, object> clientProps,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new CreateParam();
            param.roomOption = roomOption;
            param.clientInfo = new ClientInfo(userId, clientProps);

            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom("/rooms", content, receiver, onSuccess, onFailed));
        }

        /// <summary>
        ///   部屋IDを指定して入室
        /// </summary>
        public void Join(
            string roomId,
            IDictionary<string, object> clientProps,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam();
            param.clientInfo = new ClientInfo(userId, clientProps);

            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/id/{roomId}", content, receiver, onSuccess, onFailed));
        }

        /// <summary>
        ///   部屋番号を指定して入室
        /// </summary>
        /// <remarks>
        ///   TODO: 検索クエリも渡せるようにする（案件側で細かい条件指定をしたい）
        /// </remarks>
        public void Join(
            int number,
            IDictionary<string, object> clientProps,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam();
            param.clientInfo = new ClientInfo(userId, clientProps);
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/number/{number}", content, receiver, onSuccess, onFailed));
        }

        /// <summary>
        ///   検索クエリに合致する部屋にランダム入室
        /// </summary>
        public void RandomJoin(
            uint group,
            PropQuery[][] queries,
            IDictionary<string, object> clientProps,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam();
            param.queries = queries;
            param.clientInfo = new ClientInfo(userId, clientProps);
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/random/{group}", content, receiver, onSuccess, onFailed));
        }

        public void Watch(
            string roomId,
            IDictionary<string, object> clientProps,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam();
            param.clientInfo = new ClientInfo(userId, clientProps);

            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/watch/id/{roomId}", content, receiver, onSuccess, onFailed));
        }

        private async Task connectToRoom(
            string path,
            byte[] content,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            try
            {
                var cli = new HttpClient();
                cli.DefaultRequestHeaders.Add("X-Wsnet-App", appId);
                cli.DefaultRequestHeaders.Add("X-Wsnet-User", userId);
                cli.DefaultRequestHeaders.Add("X-Wsnet-Timestamp", authData.Timestamp);
                cli.DefaultRequestHeaders.Add("X-Wsnet-Nonce", authData.Nonce);
                cli.DefaultRequestHeaders.Add("X-Wsnet-Hash", authData.Hash);

                var res = await cli.PostAsync(baseUri + path, new ByteArrayContent(content));
                var body = await res.Content.ReadAsByteArrayAsync();
                if (!res.IsSuccessStatusCode)
                {
                    var msg = System.Text.Encoding.UTF8.GetString(body);
                    throw new Exception($"Connect to room failed: code={res} {msg}");
                }

                var joinedResponse = MessagePackSerializer.Deserialize<JoinedResponse>(body);
                var room = new Room(joinedResponse, userId, receiver);

                callbackPool.Add(() =>
                {
                    if (!onSuccess(room))
                    {
                        return;
                    }
                    lock(rooms)
                    {
                        rooms.Add(room);
                    }
                    Task.Run(room.Start);
                });
            }
            catch (Exception e)
            {
                callbackPool.Add(() => onFailed(e));
            }
        }

    }

}
