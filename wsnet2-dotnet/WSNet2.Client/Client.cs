using System;
using System.Net.Http;
using System.Collections.Generic;
using System.Threading.Tasks;
using WSNet2.Core;
using MessagePack;

namespace WSNet2.Client
{
    public class Client
    {
        [MessagePackObject]
        public class CreateParam
        {
            [Key("RoomOption")]
            public RoomOption roomOption;
            [Key("ClientInfo")]
            public ClientInfo clientInfo;
        }

        static async Task Main(string[] args)
        {
            var client = new HttpClient();
            client.DefaultRequestHeaders.Add("X-App-Id", "testapp");
            client.DefaultRequestHeaders.Add("X-User-Id", "testuser1");

            var pubProps = new Dictionary<string, object>(){
                {"aaa", "public"},
                {"bbb", (int)13},
            };
            var privProps = new Dictionary<string, object>(){
                {"aaa", "private"},
                {"ccc", false},
            };
            var cliProps = new Dictionary<string, object>(){
                {"name", "FooBar"},
            };

            var param = new CreateParam();
            param.roomOption = new RoomOption(10, 100, pubProps, privProps);
            param.clientInfo = new ClientInfo("id0001", cliProps);

            var opt = MessagePackSerializer.Serialize(param);
            Console.WriteLine(MessagePackSerializer.ConvertToJson(opt));

            var content = new ByteArrayContent(opt);
            var res = await client.PostAsync("http://localhost:8080/rooms", content);

            Console.WriteLine(res.StatusCode);
            Console.WriteLine(res.Content.Headers.ToString());

            var body = await res.Content.ReadAsByteArrayAsync();
            Console.WriteLine(BitConverter.ToString(body));
            Console.WriteLine(MessagePackSerializer.ConvertToJson(body));
        }
    }
}
