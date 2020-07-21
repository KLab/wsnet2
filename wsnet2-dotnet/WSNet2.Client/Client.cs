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
        class EventReceiver : WSNet2.Core.EventReceiver
        {
            CancellationTokenSource cts;

            public EventReceiver(CancellationTokenSource cts)
            {
                this.cts = cts;
            }

            public override void OnError(Exception e)
            {
                Console.WriteLine("OnError: "+e);
            }

            public override void OnJoined(Player me)
            {
                Console.WriteLine("OnJoined: "+me.Id);
            }

            public override void OnOtherPlayerJoined(Player player)
            {
                Console.WriteLine("OnOtherPlayerJoined: "+player.Id);
            }

            public override void OnLeave(Player player)
            {
                Console.WriteLine("OnLeave: "+player.Id);
            }

            public override void OnClosed(string description)
            {
                Console.WriteLine("OnClose: "+description);
                cts.Cancel();
            }

            public void RPCString(string senderId, string str)
            {
                Console.WriteLine($"OnRPCString [{senderId}]: {str}");
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

        static void RPCMessage(string senderId, StrMessage msg)
        {
            Console.WriteLine($"OnRPCMessage [{senderId}]: {msg}");
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
            receiver.RegisterRPC<StrMessage>(RPCMessage);
            receiver.RegisterRPC(receiver.RPCString);


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

                int i = 0;

                while (true) {
                    cts.Token.ThrowIfCancellationRequested();
                    var str = Console.ReadLine();

                    cts.Token.ThrowIfCancellationRequested();
                    Console.WriteLine($"input ({Thread.CurrentThread.ManagedThreadId}): {str}");

                    switch(i%3){
                        case 0:
                            var msg = new StrMessage(str);
                            room.RPC(RPCMessage, msg); //, Room.RPCToMaster);
                            break;
                        case 1:
                            room.RPC(receiver.RPCString, str); //, "id0001"); // target
                            break;
                        case 2:
                            room.RPC(receiver.RPCString, str); // broadcast
                            break;
                    }
                    i++;
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
