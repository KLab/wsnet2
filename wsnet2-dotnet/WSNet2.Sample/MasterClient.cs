using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using WSNet2.Core;
using Sample.Logic;

namespace WSNet2.Sample
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
            Console.WriteLine("OnError: " + e);
        }

        public override void OnJoined(Player me)
        {
            Console.WriteLine("OnJoined: " + me.Id);
        }

        public override void OnOtherPlayerJoined(Player player)
        {
            Console.WriteLine("OnOtherPlayerJoined: " + player.Id);
        }

        public override void OnOtherPlayerLeft(Player player)
        {
            Console.WriteLine("OnOtherPlayerLeft: " + player.Id);
        }

        public override void OnMasterPlayerSwitched(Player pred, Player newly)
        {
            Console.WriteLine($"OnMasterplayerswitched: {pred.Id} -> {newly.Id}");
        }

        public override void OnClosed(string description)
        {
            Console.WriteLine("OnClose: " + description);
            cts.Cancel();
        }

        public void RPCString(string senderId, string str)
        {
            Console.WriteLine($"OnRPCString [{senderId}]: {str}");
        }
    }

    class MasterClient
    {
        string userId;
        WSNet2Client client;
        Dictionary<string, object> props;

        int searchGroup;

        Random rand;
        GameSimulator sim;
        GameState state;
        List<PlayerEvent> events;

        public MasterClient(string server, string appId, string pKey, int serachGroup, int maxPlayer, string userId)
        {
            var authData = WSNet2Helper.GenAuthData("testapppkey", userId);
            client = new WSNet2Client(
                "http://localhost:8080",
                "testapp",
                userId,
                authData);
            props = new Dictionary<string, object>(){
                {"name", userId},
            };
            this.userId = userId;
            this.searchGroup = serachGroup;
            rand = new Random();
            sim = new GameSimulator();
            state = new GameState();
            events = new List<PlayerEvent>();
            sim.Init(state);
        }

        public async Task Serve()
        {
            var queries = new PropQuery[][]{
                new PropQuery[] {
                    new PropQuery{
                        key = "game",
                        op = OpType.Equal,
                        val = WSNet2Helper.Serialize("pong"),
                    },
                },
            };

            var cts = new CancellationTokenSource();
            var receiver = new EventReceiver(cts);

            // この順番は Unity実装と合わせる必要あり.
            receiver.RegisterRPC<EmptyMessage>(RPCKeepAlive);
            receiver.RegisterRPC<GameState>(RPCSyncGameState);
            receiver.RegisterRPC<PlayerEvent>(RPCPlayerEvent);

            var roomJoined = new TaskCompletionSource<Room>(TaskCreationOptions.RunContinuationsAsynchronously);
            Func<Room, bool> onJoined = (Room room) =>
            {
                roomJoined.TrySetResult(room);
                return true;
            };
            Action<Exception> onFailed = (Exception e) =>
            {
                roomJoined.TrySetException(e);
            };

            Console.WriteLine("random join");
            client.RandomJoin(
                (uint)searchGroup,
                queries,
                props,
                receiver,
                onJoined,
                onFailed);

            // NOTE: 起動しとかないとコールバック呼ばれない
            _ = Task.Run(async () =>
            {
                while (!roomJoined.Task.IsCompleted)
                {
                    cts.Token.ThrowIfCancellationRequested();
                    client.ProcessCallback();
                    await Task.Delay(100);
                }
            });

            var room = await roomJoined.Task;
            cts.Token.ThrowIfCancellationRequested();
            Console.WriteLine("joined room = " + room.Id);


            var syncStart = DateTime.UtcNow;
            var lastSync = syncStart;
            while (true)
            {
                cts.Token.ThrowIfCancellationRequested();
                client.ProcessCallback();

                if (state.Code == GameStateCode.WaitingGameMaster)
                {
                    state.Code = GameStateCode.WaitingPlayer;
                    state.MasterId = userId;
                }

                sim.UpdateGame(state, events);
                events.Clear();

                var forceSync = false;
                var now = DateTime.UtcNow;
                if (forceSync || 100.0 <= now.Subtract(lastSync).TotalMilliseconds)
                {
                    room.RPC(RPCSyncGameState, state);
                    lastSync = now;
                }

                await Task.Delay(16);
            }
        }

        void RPCKeepAlive(string sender, EmptyMessage _)
        {
            Console.WriteLine("RPCKeepAlive from " + sender);
        }

        void RPCPlayerEvent(string sender, PlayerEvent msg)
        {
            Console.WriteLine("RPCPlayerEvent from " + sender);
            msg.PlayerId = sender;
            events.Add(msg);
        }

        void RPCSyncGameState(string sender, GameState msg)
        {
            Console.WriteLine("RPCSyncGameState from " + sender);
            Console.WriteLine("MasterId: "+ msg.MasterId);
            Console.WriteLine("State: "+ msg.Code);
        }
    }
}
