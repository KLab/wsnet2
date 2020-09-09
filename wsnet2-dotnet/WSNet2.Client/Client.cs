using System;
using System.Collections.Generic;
using System.Linq;
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
        static AuthDataGenerator authgen = new AuthDataGenerator();

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
        static void RPCMessages(string senderId, StrMessage[] msgs)
        {
            var strs = string.Join<StrMessage>(',', msgs);
            Console.WriteLine($"OnRPCMessages [{senderId}]: {strs}"); 
        }
        static void RPCString(string senderId, string str)
        {
            Console.WriteLine($"OnRPCString [{senderId}]: {str}");
        }

        static async Task Main(string[] args)
        {
            var rand = new Random();
            var userid = $"user{rand.Next(99):000}";
            Console.WriteLine($"user id: {userid}");

            Serialization.Register<StrMessage>(0);

            var authData = authgen.Generate("testapppkey", userid);

            var client = new WSNet2Client(
                "http://localhost:8080",
                "testapp",
                userid,
                authData);

            var cliProps = new Dictionary<string, object>(){
                {"name", userid},
            };

            var cts = new CancellationTokenSource();

            var roomJoined = new TaskCompletionSource<Room>(TaskCreationOptions.RunContinuationsAsynchronously);
            Func<Room,bool> onJoined = (Room room) => {
                roomJoined.TrySetResult(room);

                room.OnError += (e) => Console.WriteLine($"OnError: {e}");
                room.OnErrorClosed += (e) => Console.WriteLine($"OnErrorClosed: {e}");
                room.OnJoined += (me) => Console.WriteLine($"OnJoined: {me.Id}");
                room.OnClosed += (m) => Console.WriteLine($"OnClosed: {m}");
                room.OnOtherPlayerJoined += (p) => Console.WriteLine($"OnOtherPlayerJoined: {p.Id}");
                room.OnOtherPlayerLeft += (p) => Console.WriteLine($"OnOtherplayerleft: {p.Id}");
                room.OnMasterPlayerSwitched += (p, n) => Console.WriteLine($"OnMasterPlayerSwitched: {p.Id} -> {n.Id}");
                room.OnRoomPropertyChanged += (visible, joinable, watchable, searchGroup, maxPlayers, clientDeadline, publicProps, privateProps) =>
                {
                    var flags = !visible.HasValue ? "-" : visible.Value ? "V" : "x";
                    flags += !joinable.HasValue ? "-" : joinable.Value ? "J" : "x";
                    flags += !watchable.HasValue ? "-" : watchable.Value ? "W" : "x";
                    var pubp = "";
                    if (publicProps != null)
                    {
                        foreach (var kv in publicProps)
                        {
                            pubp += $"{kv.Key}:{kv.Value},";
                        }
                    }
                    var prip = "";
                    if (privateProps != null)
                    {
                        foreach (var kv in privateProps)
                        {
                            prip += $"{kv.Key}:{kv.Value},";
                        }
                    }

                    Console.WriteLine($"OnRoomPropertyChanged: flg={flags} sg={searchGroup} mp={maxPlayers} cd={clientDeadline} pub={pubp} priv={prip}");
                };
                room.OnPlayerPropertyChanged += (p, props) =>
                {
                    var propstr = "";
                    foreach (var kv in props)
                    {
                        propstr += $"{kv.Key}:{kv.Value},";
                    }

                    Console.WriteLine($"OnPlayerPropertyChanged: {p.Id} {propstr}");
                };
                room.OnClosed += (_) => cts.Cancel();
                room.OnErrorClosed += (_) => cts.Cancel();

                room.RegisterRPC<StrMessage>(RPCMessage);
                room.RegisterRPC<StrMessage>(RPCMessages);
                room.RegisterRPC(RPCString);

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
                var roomOpt = new RoomOption(10, 100, pubProps, privProps).WithClientDeadline(30).WithNumber(true);

                client.Create(roomOpt, cliProps, onJoined, onFailed);
            }
            else
            {
                var number = int.Parse(args[0]);
                if (args.Length == 1)
                {
                    client.Join(number, cliProps, onJoined, onFailed);
                }
                else
                {
                    client.Watch(number, onJoined, onFailed);
                }
            }

            _ = Task.Run(async () => await callbackrunner(client, cts.Token));

            try
            {
                var room = await roomJoined.Task;
                Console.WriteLine($"joined room = {room.Id} [{room.Number}]");

                foreach (var p in room.Players){
                    var pp = $"  player {p.Key}: ";
                    foreach (var kv in p.Value.Props) {
                        pp += $"{kv.Key}:{kv.Value}, ";
                    }
                    Console.WriteLine(pp);
                }

                int i = 0;

                while (true) {
                    cts.Token.ThrowIfCancellationRequested();
                    var str = Console.ReadLine();

                    cts.Token.ThrowIfCancellationRequested();

                    if (str=="leave")
                    {
                        room.Leave();
                        continue;
                    }

                    if (str=="pause")
                    {
                        room.Pause();
                        continue;
                    }
                    if (str=="restart")
                    {
                        room.Restart();
                        continue;
                    }

                    if (str.StartsWith("switchmaster "))
                    {
                        var newMaster = str.Substring("switchmaster ".Length);
                        Console.WriteLine($"switch master to {newMaster}");
                        try{
                            room.SwitchMaster(
                                room.Players[newMaster],
                                (t, id) => Console.WriteLine($"SwitchMaster({id}) error: {t}"));
                        }
                        catch(Exception e)
                        {
                            Console.WriteLine($"switch master error: {e.Message}");
                        }
                        continue;
                    }

                    if (str.StartsWith("kick "))
                    {
                        var target = str.Substring("kick ".Length);
                        Console.WriteLine($"kick {target}");
                        try{
                            room.Kick(room.Players[target]);
                        }
                        catch(Exception e){
                            Console.WriteLine($"kick error: {e.Message}");
                        }
                        continue;
                    }

                    if (str.StartsWith("roomprop "))
                    {
                        var strs = str.Split(' ');
                        var joinable = !room.Joinable;
                        var deadline = room.ClientDeadline + 3;
                        var pubProps = new Dictionary<string, object>();
                        if (strs.Length > 1)
                        {
                            pubProps["public-modify"] = strs[1];
                        }
                        Dictionary<string, object> privProps = null;
                        if (strs.Length > 2)
                        {
                            privProps = new Dictionary<string, object>(){
                                {"private-modify", strs[2]},
                            };
                        }
                        room.ChangeRoomProperty(
                            joinable: joinable,
                            clientDeadline: deadline,
                            publicProps: pubProps,
                            privateProps: privProps,
                            onErrorResponse: (t,v,j,w,sg,mp,cd,pub,priv) => {
                                var f = !v.HasValue ? "-" : v.Value ? "V" : "x";
                                f += !j.HasValue ? "-" : j.Value ? "J" : "x";
                                f += !w.HasValue ? "-" : w.Value ? "W" : "x";
                                var pubp = "";
                                if (pub != null)
                                {
                                    foreach (var kv in pub)
                                    {
                                        pubp += $"{kv.Key}:{kv.Value},";
                                    }
                                }
                                var prip = "";
                                if (priv != null)
                                {
                                    foreach (var kv in priv)
                                    {
                                        prip += $"{kv.Key}:{kv.Value},";
                                    }
                                }
                                Console.WriteLine($"OnRoomPropertyChanged {t}: flg={f} sg={sg} mp={mp} cd={cd} pub={pubp} priv={prip}");
                            });
                        continue;
                    }

                    if (str.StartsWith("myprop "))
                    {
                        var strs = str.Split(' ');
                        if (strs.Length == 3)
                        {
                            var prop = new Dictionary<string, object>(){{strs[1], strs[2]}};
                            room.ChangeMyProperty(prop);
                        }
                        else
                        {
                            Console.WriteLine("invalid param: myprop <key> <value>");
                        }
                        continue;
                    }

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
                            room.RPC(RPCString, str, target, "nobody");
                            break;
                        case 2:
                            Console.WriteLine($"rpc to all: {str}");
                            var msgs = new StrMessage[]{
                                new StrMessage(str), new StrMessage(str),
                            };
                            room.RPC(RPCMessages, msgs);
                            break;
                    }
                    i++;
                }
            }
            catch (OperationCanceledException)
            {
            }
            catch (Exception e)
            {
                Console.WriteLine("exception: "+e);
                cts.Cancel();
            }
        }
    }
}
