using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using WSNet2;
using Sample.Logic;

namespace WSNet2.Sample
{
    class BotClient : IClient
    {
        string userId;
        AuthDataGenerator authgen;
        WSNet2Client client;
        Dictionary<string, object> props;

        Random rand;
        GameTimer timer;
        GameSimulator sim;
        GameState state;
        List<PlayerEvent> events;
        AppLogger logger;

        public BotClient(AppLogger logger)
        {
            logger.Payload.ClientType = "Bot";
            this.logger = logger;
            this.authgen = new AuthDataGenerator();
        }

        /// <summary>
        /// 1クライアントとしてルームに参加してランダムな操作を繰り返す
        /// </summary>
        /// <param name="server"></param>
        /// <param name="appId"></param>
        /// <param name="pKey"></param>
        /// <param name="userId"></param>
        /// <returns></returns>
        public async Task Serve(string server, string appId, string pKey, string userId)
        {
            logger.Payload.Server = server;

            while (true)
            {
                var authData = authgen.Generate(pKey, userId);
                client = new WSNet2Client(server, appId, userId, authData, logger);
                this.userId = userId;
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
                catch (RoomNotFoundException) {}
                catch (Exception e)
                {
                    logger.Error(e, "({0}) ServerError: {1}", userId, e);
                }
                await Task.Delay(1000);
            }
        }

        async Task<Room> JoinRandomRoom()
        {
            logger.Debug("({0}) Trying to join random room", userId);
            var query = new Query();
            query.Equal("game", "pong");
            query.Equal("state", GameStateCode.WaitingPlayer.ToString());

            var cts = new CancellationTokenSource();
            var roomJoined = new TaskCompletionSource<Room>(TaskCreationOptions.RunContinuationsAsynchronously);
            Action<Room> onJoined = (Room room) =>
            {
                room.Pause();
                roomJoined.TrySetResult(room);
            };
            Action<Exception> onFailed = (Exception e) =>
            {
                roomJoined.TrySetException(e);
            };

            client.RandomJoin(
                WSNet2Helper.SearchGroup,
                query,
                props,
                onJoined,
                onFailed,
                logger);

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

            logger.Info("({0}) Room joined {1}", userId, room.Id);
            return room;
        }

        public void OnSyncServerTick(long tick)
        {
            timer.UpdateServerTick(tick);
        }

        public void OnSyncGameState(GameState state)
        {
            this.state = state;
        }

        async Task ServeOne()
        {
            var room = await JoinRandomRoom();
            var cts = new CancellationTokenSource();

            Exception closedError = null;
            room.OnErrorClosed += (e) =>
            {
                closedError = e;
            };

            var rpc = new RPCBridge(room, this);
            room.Restart();

            long syncStart = timer.NowTick;
            long lastInputSent = syncStart;

            while (true)
            {
                cts.Token.ThrowIfCancellationRequested();
                client.ProcessCallback();

                logger.Debug("Room: {0} State: {1} Players [{2}]", room.Id, state.Code.ToString(), string.Join(", ", room.Players.Keys));

                if (closedError != null)
                {
                    logger.Error(null, "Closed Error {0}", closedError);
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
                            if (p.Id.StartsWith("master_"))
                            {
                                logger.Info("Switch master to {0}", p.Id);
                                room.ChangeRoomProperty(
                                    null, null, null,
                                    null, null, null,
                                    new Dictionary<string, object> { { "gamemaster", p.Id }, { "masterclient", "joined" } },
                                    new Dictionary<string, object> { });
                                room.SwitchMaster(p);
                                break;
                            }
                        }

                        if (10000 <= new TimeSpan(timer.NowTick - syncStart).TotalMilliseconds)
                        {
                            logger.Debug("Waiting MasterClient timeout.");
                            room.Leave();
                            break;
                        }
                    }
                }
                else if (state.Code == GameStateCode.WaitingPlayer)
                {
                    // 他の参加者を待っている
                    // 参加者が集まるとマスターが ReadyToStart に状態を変更する
                }
                else if (state.Code == GameStateCode.ReadyToStart)
                {
                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Ready,
                        PlayerId = userId,
                        Tick = timer.NowTick,
                    });
                }
                else if (state.Code == GameStateCode.InGame)
                {
                    var now = timer.NowTick;
                    if (1000 <= new TimeSpan(now - lastInputSent).TotalMilliseconds)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Move,
                            MoveInput = (MoveInputCode)rand.Next(0, 3),
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
                    logger.Info("send {0} to {1}", ev.Code, room.Master.Id);
                    rpc.PlayerEvent(ev);
                }
                events.Clear();

                await Task.Delay(100);
            }

            logger.Info("{0} left room {1}", userId, room.Id);
        }
    }
}
