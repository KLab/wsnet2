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

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="baseUri">LobbyのURI</param>
        /// <param name="appId">Wsnetに登録してあるApplication ID</param>
        /// <param name="userId">プレイヤーIDとなるID</param>
        /// <param name="authData">認証情報（アプリAPIサーバから入手）</param>
        public WSNet2Client(string baseUri, string appId, string userId, string authData)
        {
            this.baseUri = baseUri;
            this.appId = appId;
            this.userId = userId;
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
        ///   部屋を作成して入室
        /// </summary>
        /// <param name="roomOption">部屋オプション</param>
        /// <param name="clientProps">自身のカスタムプロパティ</param>
        /// <param name="onSuccess">成功時callback</param>
        /// <param name="onFailed">失敗時callback</param>
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
        public void Create(
            RoomOption roomOption,
            IDictionary<string, object> clientProps,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new CreateParam();
            param.roomOption = roomOption;
            param.clientInfo = new ClientInfo(userId, clientProps);

            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom("/rooms", content, onSuccess, onFailed));
        }

        /// <summary>
        ///   部屋IDを指定して入室
        /// </summary>
        public void Join(
            string roomId,
            IDictionary<string, object> clientProps,
            Query query,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam(){
                queries = (query!=null) ? query.condsList : null,
                clientInfo = new ClientInfo(userId, clientProps),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/id/{roomId}", content, onSuccess, onFailed));
        }

        /// <summary>
        ///   部屋番号を指定して入室
        /// </summary>
        public void Join(
            int number,
            IDictionary<string, object> clientProps,
            Query query,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam(){
                queries = (query!=null) ? query.condsList : null,
                clientInfo = new ClientInfo(userId, clientProps),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/number/{number}", content, onSuccess, onFailed));
        }

        /// <summary>
        ///   検索クエリに合致する部屋にランダム入室
        /// </summary>
        public void RandomJoin(
            uint group,
            Query query,
            IDictionary<string, object> clientProps,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam(){
                queries = (query!=null) ? query.condsList : null,
                clientInfo = new ClientInfo(userId, clientProps),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/join/random/{group}", content, onSuccess, onFailed));
        }

        /// <summary>
        ///   RoomIDを指定して観戦入室
        /// </summary>
        public void Watch(
            string roomId,
            Query query,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam(){
                queries = (query!=null) ? query.condsList : null,
                clientInfo = new ClientInfo(userId),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/watch/id/{roomId}", content, onSuccess, onFailed));
        }

        /// <summary>
        ///   部屋番号を指定して観戦入室
        /// </summary>
        public void Watch(
            int number,
            Query query,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var param = new JoinParam(){
                queries = (query!=null) ? query.condsList : null,
                clientInfo = new ClientInfo(userId),
            };
            var content = MessagePackSerializer.Serialize(param);

            Task.Run(() => connectToRoom($"/rooms/watch/number/{number}", content, onSuccess, onFailed));
        }

        private async Task connectToRoom(
            string path,
            byte[] content,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            try
            {
                var cli = new HttpClient();
                cli.DefaultRequestHeaders.Add("X-Wsnet-App", appId);
                cli.DefaultRequestHeaders.Add("X-Wsnet-User", userId);
                cli.DefaultRequestHeaders.Add("Authorization", bearer);

                var res = await cli.PostAsync(baseUri + path, new ByteArrayContent(content));
                var body = await res.Content.ReadAsByteArrayAsync();
                if (!res.IsSuccessStatusCode)
                {
                    var msg = System.Text.Encoding.UTF8.GetString(body);
                    throw new Exception($"Connect to room failed: code={res} {msg}");
                }

                var joinedResponse = MessagePackSerializer.Deserialize<JoinedResponse>(body);
                var room = new Room(joinedResponse, userId);

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
