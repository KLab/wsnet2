using System;
using System.Collections.Generic;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using WSNet2.Core;

namespace WSNet2.DotnetClient
{
    public class DotnetClient
    {
        class EventReceiver : IEventReceiver
        {
            public void OnError(Exception e)
            {
                Console.WriteLine("OnError: "+e);
            }

            public void OnJoined(Player me)
            {
                Console.WriteLine("OnJoined: "+me.Id);
            }

            public void OnOtherPlayerJoined(Player player)
            {
                Console.WriteLine("OnOtherPlayerJoined: "+player.Id);
            }
        }

        static async Task callbackrunner(WSNet2Client cli, CancellationToken ct)
        {
            while(true){
                Console.WriteLine($"callbackrunner: {Thread.CurrentThread.ManagedThreadId}");
                ct.ThrowIfCancellationRequested();
                cli.ProcessCallback();
                await Task.Delay(1000);
            }
        }

        static async Task Main(string[] args)
        {
            var client = new WSNet2Client(
                "http://localhost:8080",
                "testapp",
                "id0001",
                null);

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

            var roomOpt = new RoomOption(10, 100, pubProps, privProps).WithClientDeadline(10);

            var receiver = new EventReceiver();

            var roomCreated = new TaskCompletionSource<Room>(TaskCreationOptions.RunContinuationsAsynchronously);
            client.Create(
                roomOpt,
                cliProps,
                receiver,
                (room) => {
                    roomCreated.TrySetResult(room);
                    return true;
                },
                (e) => {
                    roomCreated.TrySetException(e);
                });

            var cts = new CancellationTokenSource();
            _ = callbackrunner(client, cts.Token);

            try
            {
                var room = await roomCreated.Task;
                Console.WriteLine("created room = "+room.Id);

                var utf8 = new UTF8Encoding();

                while (true) {
                    await Task.Delay(1);
                    Console.Write($"message? ({Thread.CurrentThread.ManagedThreadId}): ");
                    var str = Console.ReadLine();
                    Console.WriteLine("input:"+str);
                }
            }
            catch (Exception e)
            {
                Console.WriteLine("exception: "+e);
                cts.Cancel();
            }
        }
/*
        static async Task Sender(ClientWebSocket ws, CancellationToken ct)
        {
            var seqnum = 1;
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
*/
    }
}
