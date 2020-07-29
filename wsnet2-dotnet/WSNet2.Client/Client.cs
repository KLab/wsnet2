using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
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

            public override void OnOtherPlayerLeaved(Player player)
            {
                Console.WriteLine("OnOtherPlayerLeaved: "+player.Id);
            }

            public override void OnMasterPlayerSwitched(Player pred, Player newly)
            {
                Console.WriteLine($"OnMasterplayerswitched: {pred.Id} -> {newly.Id}");
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

        static AuthData genAuthData(string key, string userid)
        {
            var auth = new AuthData();

            auth.Timestamp = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds().ToString();


            var rng = new RNGCryptoServiceProvider();
            var nbuf = new byte[8];
            rng.GetBytes(nbuf);
            auth.Nonce = BitConverter.ToString(nbuf).Replace("-", "").ToLower();

            var str = userid + auth.Timestamp + auth.Nonce;
            var hmac = new HMACSHA256(Encoding.ASCII.GetBytes(key));
            var hash = hmac.ComputeHash(Encoding.ASCII.GetBytes(str));
            auth.Hash = BitConverter.ToString(hash).Replace("-", "").ToLower();

            return auth;
        }

        static void RPCMessage(string senderId, StrMessage msg)
        {
            Console.WriteLine($"OnRPCMessage [{senderId}]: {msg}");
        }

        static async Task Main(string[] args)
        {
            var rand = new Random();
            var userid = $"user{rand.Next(99):000}";
            Console.WriteLine($"user id: {userid}");

            Serialization.Register<StrMessage>(0);

            var authData = genAuthData("testapppkey", userid);

            var client = new WSNet2Client(
                "http://localhost:8080",
                "testapp",
                userid,
                authData);

            var cliProps = new Dictionary<string, object>(){
                {"name", userid},
            };

            var cts = new CancellationTokenSource();
            var receiver = new EventReceiver(cts);
            receiver.RegisterRPC<StrMessage>(RPCMessage);
            receiver.RegisterRPC(receiver.RPCString);

            var roomJoined = new TaskCompletionSource<Room>(TaskCreationOptions.RunContinuationsAsynchronously);
            Func<Room,bool> onJoined = (Room room) => {
                roomJoined.TrySetResult(room);
                return true;
            };
            Action<Exception> onFailed = (Exception e) => roomJoined.TrySetException(e);

            if (args.Length == 0)
            {
                // create room
                var pubProps = new Dictionary<string, object>(){
                    {"aaa", "public"},
                    {"bbb", (int)rand.Next(100)},
                };
                var privProps = new Dictionary<string, object>(){
                    {"aaa", "private"},
                    {"ccc", false},
                };
                var roomOpt = new RoomOption(10, 100, pubProps, privProps).WithClientDeadline(30);

                client.Create(roomOpt, cliProps, receiver, onJoined, onFailed);
            }
            else
            {
                var roomId = args[0];
                if (args.Length == 1)
                {
                    client.Join(roomId, cliProps, receiver, onJoined, onFailed);
                }
                else
                {
                    client.Watch(roomId, cliProps, receiver, onJoined, onFailed);
                }
            }

            _ = callbackrunner(client, cts.Token);

            try
            {
                var room = await roomJoined.Task;
                Console.WriteLine("joined room = "+room.Id);

                int i = 0;

                while (true) {
                    cts.Token.ThrowIfCancellationRequested();
                    var str = Console.ReadLine();

                    cts.Token.ThrowIfCancellationRequested();

                    switch(i%3){
                        case 0:
                            Console.WriteLine($"rpc to master: {str}");
                            var msg = new StrMessage(str);
                            room.RPC(RPCMessage, msg, Room.RPCToMaster);
                            break;
                        case 1:
                            var ids = room.Players.Keys.ToArray();
                            var target = ids[rand.Next(ids.Length)];
                            Console.WriteLine($"rpc to {target}: {str}");
                            room.RPC(receiver.RPCString, str, target);
                            break;
                        case 2:
                            Console.WriteLine($"rpc to all: {str}");
                            room.RPC(receiver.RPCString, str);
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
