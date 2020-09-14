using System;
using System.Linq;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using WSNet2.Core;
using Sample.Logic;

namespace WSNet2.Sample
{
    class BotClient
    {
        string userId;
        WSNet2Client client;
        Dictionary<string, object> props;

        int searchGroup;

        Random rand;
        GameTimer timer;
        GameSimulator sim;
        GameState state;
        List<PlayerEvent> events;

        /// <summary>
        /// 1クライアントとしてルームに参加してランダムな操作を繰り返す
        /// </summary>
        /// <param name="server"></param>
        /// <param name="appId"></param>
        /// <param name="pKey"></param>
        /// <param name="serachGroup"></param>
        /// <param name="userId"></param>
        /// <returns></returns>
        public async Task Serve(string server, string appId, string pKey, int serachGroup, string userId)
        {
            while (true)
            {
                var authData = WSNet2Helper.GenAuthData(pKey, userId);
                client = new WSNet2Client(server, appId, userId, authData);
                props = new Dictionary<string, object>(){
                    {"name", userId},
                };
                this.userId = userId;
                this.searchGroup = serachGroup;
                rand = new Random();
                sim = new GameSimulator(true);
                timer = new GameTimer();
                state = new GameState();
                events = new List<PlayerEvent>();
                sim.Init(state);

                try
                {
                    await ServeOne();
                }
                catch (Exception e)
                {
                    Console.WriteLine($"({userId}) Serve Error: {e}");
                }
                await Task.Delay((new Random().Next(10) + 10) * 1000);
            }
        }

        async Task<Room> JoinRandomRoom()
        {
            Console.WriteLine($"({userId}) Trying to join random room");
            var queries = new PropQuery[][]{
                new PropQuery[] {
                    new PropQuery{
                        key = "game",
                        op = OpType.Equal,
                        val = WSNet2Helper.Serialize("pong"),
                    },
                    new PropQuery{
                        key = "state",
                        op = OpType.Equal,
                        val = WSNet2Helper.Serialize(GameStateCode.WaitingPlayer.ToString()),
                    },
                },
            };

            var cts = new CancellationTokenSource();
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

            client.RandomJoin((uint)searchGroup, queries, props, onJoined, onFailed);

            // FIXME: 起動しとかないとコールバック呼ばれないが汚い
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

            Console.WriteLine($"({userId}) Room joined {room.Id}");
            return room;
        }

        async Task<Room> CreateRoom()
        {
            Console.WriteLine($"({userId}) Trying to create room");

            var cts = new CancellationTokenSource();
            uint MaxPlayers = 3;
            uint Deadline = 3;
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

            var pubProps = new Dictionary<string, object>(){
                {"game", "pong"},
                {"masterclient", "waiting"},
                {"state", GameStateCode.WaitingGameMaster.ToString()},
            };
            var privProps = new Dictionary<string, object>(){
                {"aaa", "private"},
                {"ccc", false},
            };
            var cliProps = new Dictionary<string, object>(){
                {"userId", userId},
            };
            var roomOpt = new RoomOption(MaxPlayers, (uint)searchGroup, pubProps, privProps);
            roomOpt.WithClientDeadline(Deadline);

            client.Create(roomOpt, cliProps, onJoined, onFailed);

            // FIXME: 起動しとかないとコールバック呼ばれないが汚い
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

            Console.WriteLine($"({userId}) Room created ${room.Id}");
            return room;
        }

        async Task<Room> JoinOrCreateRoom()
        {
            Room room = null;
            try
            {
                room = await JoinRandomRoom();
            }
            catch (Exception e)
            {
                // Console.WriteLine($"({userId}) Failed to join room {e}");
            }

            if (room == null)
            {
                room = await CreateRoom();
            }
            return room;
        }

        async Task ServeOne()
        {
            var room = await JoinOrCreateRoom();
            var cts = new CancellationTokenSource();

            Exception closedError = null;
            room.OnErrorClosed += (e) => {
                closedError = e;
            };

            var RPCSyncServerTick = new Action<string, long>((sender, tick) =>
            {
                if (room.Master.Id == sender)
                {
                    timer.UpdateServerTick(tick);
                }
            });
            var RPCSyncGameState = new Action<string, GameState>((sender, state_) =>
            {
                if (room.Master.Id == sender)
                {
                    // 同一スレッドから呼ばれるのでロック取らなくて良い
                    state = state_;
                }
            });
            var RPCPlayerEvent = new Action<string, PlayerEvent>((sender, ev) => { });

            // この順番は Unity実装と合わせる必要あり.
            room.RegisterRPC<GameState>(RPCSyncGameState);
            room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
            room.RegisterRPC(RPCSyncServerTick);
            room.Restart();

            long syncStart = timer.NowTick;
            long lastInputSent = syncStart;
            bool joinSent = false;
            bool readySent = false;

            while (true)
            {
                cts.Token.ThrowIfCancellationRequested();
                client.ProcessCallback();

                Console.WriteLine("Room: {0} State: {1} Players [{2}]", room.Id, state.Code.ToString(), string.Join(", ", room.Players.Keys));

                if (closedError != null) {
                    Console.WriteLine($"Closed Error {closedError}");
                    break;
                }

                if (state.Code == GameStateCode.WaitingGameMaster)
                {
                    // 本当はルームマスタがルームを作成するシーケンスを想定しているが, サンプルは簡単のため,
                    // マスタークライアントがJoinしてきたら, ルームマスタを委譲する
                    if (room.Me == room.Master)
                    {
                        foreach (var p in room.Players.Values)
                        {
                            if (p.Id.StartsWith("gamemaster"))
                            {
                                Console.WriteLine($"Switch master to {p.Id}");
                                room.ChangeRoomProperty(
                                    null, null, null,
                                    null, null, null,
                                    new Dictionary<string, object> { { "gamemaster", p.Id }, { "masterclient", "joined" } },
                                    new Dictionary<string, object> { });
                                room.SwitchMaster(p);
                                break;
                            }
                        }
                    }
                }
                else if (state.Code == GameStateCode.WaitingPlayer)
                {
                    if (!joinSent)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Join,
                            PlayerId = userId,
                            Tick = timer.NowTick,
                        });
                        joinSent = true;
                    }
                }
                else if (state.Code == GameStateCode.ReadyToStart)
                {
                    if (!readySent)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Ready,
                            PlayerId = userId,
                            Tick = timer.NowTick,
                        });
                        readySent = true;
                    }
                }
                else if (state.Code == GameStateCode.InGame)
                {
                    var now = timer.NowTick;
                    if (1000 <= new TimeSpan(now - lastInputSent).TotalMilliseconds)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Move,
                            MoveInput = (MoveInputCode)rand.Next(0, 2),
                            PlayerId = userId,
                            Tick = now,
                        });
                        lastInputSent = now;
                    }
                }
                else if (state.Code == GameStateCode.End)
                {
                    room.Leave();
                    break;
                }

                foreach (var ev in events)
                {
                    Console.WriteLine("send {0} to {1}", ev.Code, room.Master.Id);
                    room.RPC(RPCPlayerEvent, ev, new string[] { room.Master.Id });
                }
                events.Clear();

                await Task.Delay(100);
            }

            Console.WriteLine(userId + " left from " + room.Id);
        }
    }
}
