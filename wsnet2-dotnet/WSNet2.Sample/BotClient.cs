using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
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
        GameSimulator simulator;
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
                simulator = new GameSimulator(false);
                timer = new GameTimer();
                state = new GameState();
                events = new List<PlayerEvent>();
                simulator.Init(state);

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
            query.Equal(WSNet2Helper.PubKey.Game, WSNet2Helper.GameName);
            query.Equal(WSNet2Helper.PubKey.State, GameStateCode.WaitingPlayer.ToString());

            Room joinedRoom = null;
            Exception joinException = null;

            client.RandomJoin(
                WSNet2Helper.SearchGroup,
                query,
                props,
                (room) => {
                    room.Pause();
                    joinedRoom = room;
                },
                (e) => {
                    joinException = e;
                },
                logger);

            while (joinedRoom == null && joinException == null)
            {
                client.ProcessCallback();
                await Task.Delay(100);
            }

            if (joinException != null) {
                throw joinException;
            }

            logger.Info("({0}) Room joined {1}", userId, joinedRoom.Id);
            return joinedRoom;
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

                if (state.Code == GameStateCode.ReadyToStart)
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
