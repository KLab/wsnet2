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
        Random rand;
        GameTimer timer;
        GameState state;
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
                timer = new GameTimer();
                state = new GameState();

                try
                {
                    await ServeOne();
                }
                catch (RoomNotFoundException) {}
                catch (Exception e)
                {
                    logger.Error(e, "({0}) ServerError {1}", userId, e);
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
                null,
                (room) => {
                    room.Pause();
                    joinedRoom = room;
                },
                (e) => joinException = e,
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
            var rpc = new RPCBridge(room, this);
            Exception closedError = null;
            room.OnErrorClosed += (e) => closedError = e;
            room.Restart();

            while (true)
            {
                client.ProcessCallback();

                if (closedError != null)
                {
                    logger.Error(closedError, $"({userId}) Room closed with error");
                    break;
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
                    break;
                }

                await Task.Delay(100);
            }

            logger.Info("({0}) Left from room {1}", userId, room.Id);
        }
    }
}
