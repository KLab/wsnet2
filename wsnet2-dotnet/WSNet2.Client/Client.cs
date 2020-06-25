using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using WSNet2.Core;

namespace WSNet2.DotnetClient
{
    public class StrMessage : IWSNetSerializable
    {
        string str;

        public StrMessage(){}
        public StrMessage(string str)
        {
            this.str = str;
        }

        public void Serialize(SerialWriter writer)
        {
            writer.Write(str);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            str = reader.ReadString();
        }

        public override string ToString()
        {
            return str;
        }
    }


    public class DotnetClient
    {
        class EventReceiver : IEventReceiver
        {
            CancellationTokenSource cts;

            public EventReceiver(CancellationTokenSource cts)
            {
                this.cts = cts;
            }

            public void OnError(Exception e)
            {
                Console.WriteLine("OnError: "+e);
                cts.Cancel();
            }

            public void OnJoined(Player me)
            {
                Console.WriteLine("OnJoined: "+me.Id);
            }

            public void OnOtherPlayerJoined(Player player)
            {
                Console.WriteLine("OnOtherPlayerJoined: "+player.Id);
            }

            public void OnMessage(EvMessage ev)
            {
                var msg = ev.Body<StrMessage>();
                Console.WriteLine($"OnMessage[{ev.SenderID}]: {msg}");
            }
        }

        static async Task callbackrunner(WSNet2Client cli, CancellationToken ct)
        {
            while(true){
                ct.ThrowIfCancellationRequested();
                cli.ProcessCallback();
                await Task.Delay(1000);
            }
        }

        static async Task Main(string[] args)
        {
            Serialization.Register<StrMessage>(0);

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

            var cts = new CancellationTokenSource();
            var receiver = new EventReceiver(cts);

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

            _ = callbackrunner(client, cts.Token);

            try
            {
                var room = await roomCreated.Task;
                Console.WriteLine("created room = "+room.Id);

                while (true) {
                    cts.Token.ThrowIfCancellationRequested();
                    var str = Console.ReadLine();

                    cts.Token.ThrowIfCancellationRequested();
                    Console.WriteLine($"input ({Thread.CurrentThread.ManagedThreadId}): {str}");

                    var msg = new StrMessage(str);
                    room.Broadcast(msg);
                }
            }
            catch (Exception e)
            {
                Console.WriteLine("exception: "+e);
                cts.Cancel();
            }
        }
    }
}
