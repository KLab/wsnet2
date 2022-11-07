using System;
using System.Threading;
using System.Threading.Tasks;
using Sample.Logic;

namespace WSNet2.Sample
{
    class BotClient : IClient
    {
        static int updateIntervalMillSec = 100;
        string userId;
        AuthDataGenerator authgen;
        WSNet2Client client;
        Room room;
        RPCBridge rpc;
        Random rand;
        GameTimer timer;
        GameState state;
        AppLogger logger;

        public BotClient(AppLogger logger)
        {
            logger.Payload.ClientType = "Bot";
            this.logger = logger;
            authgen = new AuthDataGenerator();
            rand = new Random();
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
            this.userId = userId;

            while (true)
            {
                client = new WSNet2Client(server, appId, userId, authgen.Generate(pKey, userId), logger);
                state = new GameState();
                timer = new GameTimer();
                room = null;
                rpc = null;

                var cts = new CancellationTokenSource();
                try
                {
                    JoinRandomRoom(cts);
                    await Updater(cts.Token);
                }
                catch (OperationCanceledException) {}
                catch (RoomNotFoundException) {}
                catch (Exception e)
                {
                    logger.Error(e, "({0}) ServeError {1}", userId, e);
                }

                await Task.Delay(1000);
            }
        }

        void JoinRandomRoom(CancellationTokenSource cts)
        {
            logger.Debug("({0}) Trying to join random room", userId);
            var query = new Query();
            query.Equal(WSNet2Helper.PubKey.Game, WSNet2Helper.GameName);
            query.Equal(WSNet2Helper.PubKey.State, GameStateCode.WaitingPlayer.ToString());

            client.RandomJoin(
                WSNet2Helper.SearchGroup,
                query,
                null,
                (room) => {
                    logger.Info("({0}) Room joined {1}", userId, room.Id);
                    room.OnClosed += msg =>
                    {
                        logger.Info($"Close: {msg}");
                        cts.Cancel();
                    };
                    room.OnError += e =>
                    {
                        logger.Error(e, $"OnError: {e.Message}");
                    };
                    room.OnErrorClosed += e =>
                    {
                        logger.Error(e, $"OnErrorClosed: {e.Message}");
                        cts.Cancel();
                    };
                    room.OnOtherPlayerLeft += (p, msg) =>
                    {
                        logger.Info($"PlayerLeft: {p.Id} Msg: {msg}");
                        cts.Cancel();
                    };
                    rpc = new RPCBridge(room, this);
                    this.room = room;
                },
                (e) => throw e,
                logger);
        }

        async Task Updater(CancellationToken ct)
        {
            while (true)
            {
                ct.ThrowIfCancellationRequested();
                var timeslice = Task.Delay(updateIntervalMillSec);
                client.ProcessCallback();
                GameUpdate();
                await timeslice;
            }
        }

        void GameUpdate()
        {
            if (room == null)
            {
                return;
            }

            if (state.Code == GameStateCode.WaitingPlayer)
            {
                rpc.PlayerEvent(new PlayerEvent
                {
                    Code = PlayerEventCode.Join,
                    PlayerId = userId,
                    Tick = timer.NowTick,
                });
            }

            if (state.Code == GameStateCode.ReadyToStart)
            {
                rpc.PlayerEvent(new PlayerEvent
                {
                    Code = PlayerEventCode.Ready,
                    PlayerId = userId,
                    Tick = timer.NowTick,
                });
            }

            if (state.Code == GameStateCode.InGame)
            {
                rpc.PlayerEvent(new PlayerEvent
                {
                    Code = PlayerEventCode.Move,
                    MoveInput = (MoveInputCode)rand.Next(0, 3),
                    PlayerId = userId,
                    Tick = timer.NowTick,
                });
            }

            if (state.Code == GameStateCode.End)
            {
                room.Leave();
            }
        }

        public void OnSyncServerTick(long tick)
        {
            timer.UpdateServerTick(tick);
        }

        public void OnSyncGameState(GameState state)
        {
            this.state = state;
        }
    }
}
