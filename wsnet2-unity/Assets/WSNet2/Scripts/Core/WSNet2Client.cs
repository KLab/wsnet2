using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using MessagePack;

namespace WSNet2.Core
{
    public class WSNet2Client
    {
        string baseUri;
        string appId;
        string userId;
        byte[] authData;

        List<Room> rooms = new List<Room>();
        CallbackPool callbackPool = new CallbackPool();

        public WSNet2Client(string baseUri, string appId, string userId, byte[] authData)
        {
            this.baseUri = baseUri;
            this.appId = appId;
            this.userId = userId;
            this.authData = authData;
        }

        public void ProcessCallback()
        {
            callbackPool.Process();
            lock(rooms)
            {
                foreach (var room in rooms)
                {
                    room.ProcessCallback();
                    // todo: 終わったroomを削除
                }
            }
        }

        public void Create(
            RoomOption roomOption,
            IDictionary<string, object> clientProps,
            IEventReceiver receiver,
            Func<Room, bool> onSuccess,
            Action<Exception> onFailed)
        {
            var _ = create(roomOption, clientProps, receiver, onSuccess, onFailed);
        }

        private async Task create(
            RoomOption roomOption,
            IDictionary<string, object> clientProps,
            IEventReceiver receiver,
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
                    _ = Task.Run(room.Start);
                });
            }
            catch (Exception e)
            {
                callbackPool.Add(() => onFailed(e));
            }
        }

    }

}
