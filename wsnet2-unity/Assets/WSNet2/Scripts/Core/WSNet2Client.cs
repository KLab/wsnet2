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
        byte[] authData;

        List<Room> rooms = new List<Room>();
        CallbackPool callbackPool = new CallbackPool();

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="baseUri">LobbyのURI</param>
        /// <param name="appId">Wsnetに登録してあるApplication ID</param>
        /// <param name="userId">プレイヤーIDとなるID</param>
        /// <param name="authData">認証情報（アプリAPIサーバから入手）</param>
        public WSNet2Client(string baseUri, string appId, string userId, byte[] authData)
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
            // create()の中では全体をtry-catchして例外はonFailedに流す。
            Task.Run(() => create(roomOption, clientProps, receiver, onSuccess, onFailed));
        }

        private async Task create(
            RoomOption roomOption,
            IDictionary<string, object> clientProps,
            EventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            try
            {
                var param = new CreateParam();
                param.roomOption = roomOption;
                param.clientInfo = new ClientInfo(userId, clientProps);

                var opt = MessagePackSerializer.Serialize(param);
                var content = new ByteArrayContent(opt);

                // todo: 認証
                var cli = new HttpClient();
                cli.DefaultRequestHeaders.Add("X-App-Id", appId);
                cli.DefaultRequestHeaders.Add("X-User-Id", userId);

                var res = await cli.PostAsync(baseUri + "/rooms", content);
                var body = await res.Content.ReadAsByteArrayAsync();
                if (!res.IsSuccessStatusCode)
                {
                    var msg = System.Text.Encoding.UTF8.GetString(body);
                    throw new Exception($"Create failed: code={res} {msg}");
                }

                var joined = MessagePackSerializer.Deserialize<JoinedResponse>(body);
                var room = new Room(joined, userId, receiver);

                callbackPool.Add(() =>
                {
                    Console.WriteLine("callback onsuccess");
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
