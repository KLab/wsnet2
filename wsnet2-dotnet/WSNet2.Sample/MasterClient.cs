using System;
using System.Linq;
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
        GameTimer timer;
        GameSimulator sim;
        List<GameState> states;
        List<PlayerEvent> events;
        List<PlayerEvent> newEvents;

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
                timer = new GameTimer();
                states = new List<GameState>();
                events = new List<PlayerEvent>();
                newEvents = new List<PlayerEvent>();
                var state = new GameState();
                sim.Init(state);
                state.UpdatedAt = timer.NowTick;
                states.Add(state);

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

            var RPCSyncServerTick = new Action<string, long>((sender, tick) => { });

            // この順番は Unity実装と合わせる必要あり.
            room.RegisterRPC<GameState>(RPCSyncGameState);
            room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
            room.RegisterRPC(RPCSyncServerTick);
            room.Restart();

            long syncStart = timer.NowTick;
            long lastSync = syncStart;
            long roomEmptyTick = 0;

            while (true)
            {
                cts.Token.ThrowIfCancellationRequested();
                client.ProcessCallback();

                // ルーム Create したクライアントが Master を譲ってくれるはず
                if (room.Master != room.Me)
                {
                    if (room.PublicProps.ContainsKey("gamemaster"))
                    {
                        // すでに別のMasterクライアントが入室していたら抜ける
                        var gm = room.PublicProps["gamemaster"].ToString();
                        Console.WriteLine($"gamemaster {gm}");
                        if (gm != room.Me.Id)
                        {
                            room.Leave();
                            break;
                        }
                    }

                    foreach (var kv in room.PublicProps)
                    {
                        Console.WriteLine($"(pub) {kv.Key} : {kv.Value}");
                    }

                    await Task.Delay(100);
                    continue;
                }

                // 前回のループから今回までの間にやってきた PlayerEvent が newEvents に格納されている.
                // 再計算可能なもののみを抽出する.
                long oldestTick = timer.NowTick;
                bool newEventExist = 0 < newEvents.Count;
                if (newEventExist)
                {
                    foreach (var ev in newEvents)
                    {
                        if (states[0].UpdatedAt < ev.Tick)
                        {
                            events.Add(ev);
                            oldestTick = Math.Min(oldestTick, ev.Tick);
                        }
                        else
                        {
                            Console.WriteLine("Discard PlayerEvent: too past tick"); // TODO どうハンドルするべきか
                        }
                    }
                    events.Sort((a, b) => a.Tick.CompareTo(b.Tick));
                    newEvents.Clear();
                }

                // 再計算可能な直近の GameState を探しつつ、それよりも新しいものは破棄する.
                while (oldestTick <= states[states.Count - 1].UpdatedAt)
                {
                    states.RemoveAt(states.Count - 1);
                }

                var state = states[states.Count - 1].Copy();

                if (state.Code == GameStateCode.WaitingGameMaster)
                {
                    state.Code = GameStateCode.WaitingPlayer;
                    state.MasterId = userId;
                }

                var now = timer.NowTick;
                var targetEvents = events.Where(ev => oldestTick <= ev.Tick && ev.Tick < now);
                sim.UpdateGame(now, state, targetEvents);
                states.Add(state);

                if (50 < states.Count)
                {
                    // 一番古い GameState を破棄する.
                    // O(n) だが要素数少ないのでよいだろう
                    states.RemoveAt(0);

                    // 残ったもののうち一番古い State よりも古い PlayerEvent はもう復元に使えないので削除する.
                    long t = states[0].UpdatedAt;
                    int idx = events.FindIndex(ev => t < ev.Tick);
                    if (idx != -1) {
                        events.RemoveRange(0, idx);
                    }
                }

                // 0.1秒ごとにゲーム状態の同期メッセージを送信する
                if (newEventExist || 100.0 <= new TimeSpan(now - lastSync).TotalMilliseconds)
                {
                    room.RPC(RPCSyncServerTick, timer.NowTick);
                    room.RPC(RPCSyncGameState, state);
                    lastSync = now;
                }

                // プレイヤーが誰もいなくなった状態が一定時間続いたら部屋から抜ける
                if (room.Players.Count <= 1)
                {
                    if (roomEmptyTick == 0)
                    {
                        roomEmptyTick = now;
                    }

                    if (5000 <= new TimeSpan(now - roomEmptyTick).TotalMilliseconds)
                    {
                        room.Leave();
                        break;
                    }
                }
                else
                {
                    roomEmptyTick = 0;
                }

                await Task.Delay(16);
            }

            Console.WriteLine(userId + " left " + room.Id);
        }

        void RPCPlayerEvent(string sender, PlayerEvent msg)
        {
            Console.WriteLine("RPCPlayerEvent from " + sender);
            msg.PlayerId = sender;
            newEvents.Add(msg);
        }

        void RPCSyncGameState(string sender, GameState msg)
        {
            // 未使用
        }
    }
}
