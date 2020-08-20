using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using WSNet2.Core;
using Sample.Logic;

namespace WSNet2.Sample
{
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
            // この順番は Unity実装と合わせる必要あり.

            var roomJoined = new TaskCompletionSource<Room>(TaskCreationOptions.RunContinuationsAsynchronously);
            Func<Room, bool> onJoined = (Room room) =>
            {
                room.Pause();
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

            room.RegisterRPC<GameState>(RPCSyncGameState);
            room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
            room.Restart();

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
