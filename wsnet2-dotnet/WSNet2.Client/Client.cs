using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Net.WebSockets;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using MessagePack;
using WSNet2.Core;

namespace WSNet2.Client
{
    public class Client
    {
        static async Task Main(string[] args)
        {
            var client = new HttpClient();
            client.DefaultRequestHeaders.Add("X-App-Id", "testapp");
            client.DefaultRequestHeaders.Add("X-User-Id", "id0001");

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
            param.roomOption = new RoomOption(10, 100, pubProps, privProps).WithClientDeadline(10);
            param.clientInfo = new ClientInfo("id0001", cliProps);

            var opt = MessagePackSerializer.Serialize(param);
            Console.WriteLine(MessagePackSerializer.ConvertToJson(opt));

            var content = new ByteArrayContent(opt);
            var res = await client.PostAsync("http://localhost:8080/rooms", content);

            Console.WriteLine(res.StatusCode);
            Console.WriteLine(res.Content.Headers.ToString());

            var body = await res.Content.ReadAsByteArrayAsync();

            Console.WriteLine(MessagePackSerializer.ConvertToJson(body));

            var join = MessagePackSerializer.Deserialize<JoinedResponse>(body);
            Console.WriteLine("room url:"+join.url);

            var ws = new ClientWebSocket();
            ws.Options.SetRequestHeader("X-Wsnet-App", "testapp");
            ws.Options.SetRequestHeader("X-Wsnet-User", "id0001");
            ws.Options.SetRequestHeader("X-Wsnet-LastEventSeq", "0");

            var cts = new CancellationTokenSource();
            await ws.ConnectAsync(new Uri(join.url), cts.Token);

            try{
                var tasks = new Task[2];

                tasks[0] = Receiver(ws, cts.Token);
                tasks[1] = Sender(ws, cts.Token);

                await Task.WhenAny(tasks);
            }
            catch(Exception e) {
                Console.WriteLine("exception: "+e);
            }
            cts.Cancel();

            Console.WriteLine("close: "+ws.CloseStatusDescription);
        }

        static async Task Sender(ClientWebSocket ws, CancellationToken ct)
        {
            var seqnum = 1;
            var utf8 = new UTF8Encoding();
            while (true) {
                if (ws.State != WebSocketState.Open) {
                    await Console.Out.WriteLineAsync("sender: state != open"+ws.State);
                    break;
                }

                Console.Write("message?: ");
                var msg = Console.ReadLine();

                ct.ThrowIfCancellationRequested();

                var len = utf8.GetByteCount(msg);

                var buf = new byte[len+4];
                buf[0] = 34; // MsgTypeBroadcast
                buf[1] = (byte)((seqnum & 0xff0000) >> 16);
                buf[2] = (byte)((seqnum & 0xff00) >> 8);
                buf[3] = (byte)(seqnum & 0xff);
                utf8.GetBytes(msg, 0, msg.Length, buf, 4);
                seqnum++;

                await ws.SendAsync(buf, WebSocketMessageType.Binary, true, ct);
                await Task.Delay(100);
            }
            await Console.Out.WriteLineAsync("sender finish");
        }

        static async Task Receiver(ClientWebSocket ws, CancellationToken ct)
        {
            var utf8 = new UTF8Encoding();
            while (true){
                if (ws.State != WebSocketState.Open) {
                    await Console.Out.WriteLineAsync("receiver: state != open"+ws.State);
                    break;
                }

                var buf = new byte[1024];
                var ret = await ws.ReceiveAsync(buf, ct);

                await Console.Out.WriteLineAsync(BitConverter.ToString(new ArraySegment<byte>(buf,0,ret.Count).ToArray()));

                if (ret.Count<1){
                    break;
                }
                var evtype = buf[0];
                if (evtype == 1) {
                    // EvTypePeerReady
                    var seqnum = ((int)buf[1]<<16) + ((int)buf[2]<<8)+ (int)buf[3];
                    await Console.Out.WriteLineAsync("type:1 PeerReady: seqnum="+seqnum);
                }
                if (evtype >= 30){
                    var seqnum = ((int)buf[1] << 24) + ((int)buf[2] << 16) + ((int)buf[3]<<8) + (int)buf[4];
                    await Console.Out.WriteLineAsync("type:"+evtype+" seq:"+seqnum);
                    if (evtype==34){
                        // username as Str8
                        var len = (int)buf[6];
                        await Console.Out.WriteLineAsync("msg len="+len);
                        var name = utf8.GetString(buf, 7, len);
                        // payload
                        var msg = utf8.GetString(buf, 7+len, ret.Count-7-len);
                        await Console.Out.WriteLineAsync("receive msg: "+ name + ": " + msg);
                    }
                }
            }
            await Console.Out.WriteLineAsync("receiver finish");
        }

    }
}
