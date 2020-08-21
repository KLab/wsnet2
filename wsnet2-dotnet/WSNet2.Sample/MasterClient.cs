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

        /// <summary>
        /// 1クライアントとしてルームに参加してMasterClientとして振る舞う
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
                var authData = WSNet2Helper.GenAuthData("testapppkey", userId);
                client = new WSNet2Client(server, appId, userId, authData);
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

                try
                {
                    await ServeOne();
                }
                catch (Exception e)
                {
                    // FIXME: 例外の種類を増やすべき
                    if (e.ToString().Contains("Connect to room failed"))
                    {
                        Console.WriteLine($"({userId}) no room found");
                    }
                    else
                    {
                        Console.WriteLine("({userId}) Serve Error: {0}", e);
                    }
                }
                await Task.Delay(1000);
            }
        }

        async Task ServeOne()
        {
            var queries = new PropQuery[][]{
                new PropQuery[] {
                    new PropQuery{
                        key = "game",
                        op = OpType.Equal,
                        val = WSNet2Helper.Serialize("pong"),
                    },
                    new PropQuery{
                        key = "masterclient",
                        op = OpType.Equal,
                        val = WSNet2Helper.Serialize("waiting"),
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

            client.RandomJoin(
                (uint)searchGroup,
                queries,
                props,
                onJoined,
                onFailed);

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

            Console.WriteLine(userId + " joined " + room.Id);

            // この順番は Unity実装と合わせる必要あり.
            room.RegisterRPC<GameState>(RPCSyncGameState);
            room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
            room.Restart();

            var syncStart = DateTime.UtcNow;
            var lastSync = syncStart;
            var roomEmptySince = DateTime.MinValue;

            while (true)
            {
                cts.Token.ThrowIfCancellationRequested();
                client.ProcessCallback();

                // ルーム Create したクライアントが Master を譲ってくれるはず
                if (room.Master != room.Me)
                {
                    if (room.PublicProps.ContainsKey("gamemaster")) {
                        // すでに別のMasterクライアントが入室していたら抜ける
                        var gm = room.PublicProps["gamemaster"].ToString();
                        Console.WriteLine($"gamemaster {gm}");
                        if (gm != room.Me.Id) {
                            room.Leave();
                            break;
                        }
                    }

                    foreach (var kv in room.PublicProps) {
                        Console.WriteLine($"(pub) {kv.Key} : {kv.Value}");
                    }

                    await Task.Delay(100);
                    continue;
                }

                if (state.Code == GameStateCode.WaitingGameMaster)
                {
                    state.Code = GameStateCode.WaitingPlayer;
                    state.MasterId = userId;
                }

                sim.UpdateGame(state, events);
                events.Clear();

                var forceSync = false;
                var now = DateTime.UtcNow;

                // 0.1秒ごとにゲーム状態の同期メッセージを送信する
                if (forceSync || 100.0 <= now.Subtract(lastSync).TotalMilliseconds)
                {
                    room.RPC(RPCSyncGameState, state);
                    lastSync = now;
                }

                // プレイヤーが誰もいなくなった状態が一定時間続いたら部屋から抜ける
                if (room.Players.Count <= 1)
                {
                    if (roomEmptySince == DateTime.MinValue)
                    {
                        roomEmptySince = now;
                    }

                    if (5000 <= now.Subtract(roomEmptySince).TotalMilliseconds)
                    {
                        room.Leave();
                        break;
                    }
                }
                else
                {
                    roomEmptySince = DateTime.MaxValue;
                }

                await Task.Delay(16);
            }

            Console.WriteLine(userId + " left " + room.Id);
        }

        void RPCPlayerEvent(string sender, PlayerEvent msg)
        {
            Console.WriteLine("RPCPlayerEvent from " + sender);
            msg.PlayerId = sender;
            events.Add(msg);
        }

        void RPCSyncGameState(string sender, GameState msg)
        {
            // 未使用
        }
    }
}
