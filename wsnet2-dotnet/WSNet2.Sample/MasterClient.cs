using System;
using System.Linq;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using WSNet2;
using Sample.Logic;

namespace WSNet2.Sample
{
    class MasterClient : IMasterClient
    {
        /// <summary>Pongゲームの最大プレイヤー数（2Player + MasterClient)</summary>
        const uint MaxPlayers = 3;

        /// <summary>タイムアウト (秒)</summary>
        const uint Deadline = 10;

        /// <summary>1フレームの時間</summary>
        const int frameMilliSec = 1000/60;

        string userId;
        string pKey;
        AuthDataGenerator authgen;
        WSNet2Client client;

        Room room;
        RPCBridge rpc;
        GameTimer timer;
        GameSimulator sim;
        List<GameState> states;
        List<PlayerEvent> events;
        List<PlayerEvent> newEvents;
        AppLogger logger;
        long lastSync;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public MasterClient(string server, string appId, string userId, string pKey, AppLogger logger)
        {
            logger.Payload.ClientType = "Master";
            logger.Payload.ClientId = userId;
            logger.Payload.Server = server;
            this.logger = logger;
            this.authgen = new AuthDataGenerator();
            this.userId = userId;
            this.pKey = pKey;
            this.client = new WSNet2Client(server, appId, userId, authgen.Generate(pKey, userId), logger);

            sim = new GameSimulator(true);
            timer = new GameTimer();
            states = new List<GameState>();
            events = new List<PlayerEvent>();
            newEvents = new List<PlayerEvent>();
        }

        /// <summary>
        /// 部屋を作成しゲーム実行することを繰り返す
        /// </summary>
        public async Task Serve()
        {
            while (true)
            {
                var cts = new CancellationTokenSource();
                try
                {
                    events.Clear();
                    newEvents.Clear();

                    var state = new GameState();
                    state.Tick = timer.NowTick;
                    state.MasterId = userId;
                    sim.Init(state);
                    states.Add(state);
                    lastSync = timer.NowTick;

                    CreateRoom(cts);
                    await Updater(cts.Token);
                }
                catch (TaskCanceledException)
                {
                }
                catch (Exception e)
                {
                    cts.Cancel();
                    logger.Error(e, $"Serve: {0}", e);
                }

                await Task.Delay(1000);
            }
        }

        /// <summary>
        ///   ゲームループを駆動する
        /// </summary>
        /// <remarks>
        ///   frameMillSec毎に、client.ProcessCallbackとGameUpdateを呼び出す。
        ///   client.ProcessCallbackを呼び出さないと、client.Createの引数のCallbackが呼ばれないことに注意。
        /// </remarks>
        async Task Updater(CancellationToken ct)
        {
            while (true)
            {
                ct.ThrowIfCancellationRequested();
                var timeslice = Task.Delay(frameMilliSec);
                client.ProcessCallback();
                GameUpdate();
                await timeslice;
            }
        }

        /// <summary>
        ///   部屋を作成する
        /// </summary>
        void CreateRoom(CancellationTokenSource cts)
        {
            // 部屋の公開プロパティ
            // 入室時のQueryによるフィルタリングにも使われる
            var pubProps = new Dictionary<string, object>()
            {
                {WSNet2Helper.PubKey.Game, WSNet2Helper.GameName},
                {WSNet2Helper.PubKey.State, GameStateCode.WaitingPlayer.ToString()},
                {WSNet2Helper.PubKey.PlayerNum, (byte)0},
                {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
            };

            // Clientのプロパティ
            var props = new Dictionary<string, object>
            {
                {"name", userId},
            };

            // clientを々利用するため、使う直前に認証情報を更新
            client.UpdateAuthData(authgen.Generate(pKey, userId));

            // 部屋を検索可能、入室可能、観戦可能として作成する
            var roomOpt = new RoomOption(MaxPlayers, WSNet2Helper.SearchGroup, pubProps, null)
                .Visible(true).Joinable(true).Watchable(true);
            client.Create(
                roomOpt,
                props,
                (room) =>
                {
                    room.OnClosed += msg =>
                    {
                        logger.Info("Close: {0}", msg);
                        cts.Cancel();
                    };
                    room.OnError += e =>
                    {
                        logger.Error(e, "OnError: {0}", e.Message);
                    };
                    room.OnErrorClosed += e =>
                    {
                        logger.Error(e, "OnErrorClosed: {0}", e.Message);
                        cts.Cancel();
                    };
                    room.OnOtherPlayerJoined += OnPlayerJoined;
                    this.room = room;
                    this.rpc = new RPCBridge(room, this);
                },
                (e) => throw e,
                logger);
        }

        /// <summary>
        ///   プレイヤーの入室を待機しているときの入室通知を受け取る
        /// </summary>
        void OnPlayerJoined(Player player)
        {
            if (room.PlayerCount == 2) // Master + Player
            {
/*
                room.OnOtherPlayerLeft += OnWaitingPlayerLeft; // 二人目を待つ間に退室したときの処理
                var prop = new Dictionary<string, object>
                {
                    {WSNet2Helper.PubKey.PlayerNum, (byte)1},
                    {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                };
                room.ChangeRoomProperty(publicProps: prop);
                */
                newEvents.Add(new PlayerEvent()
                    {
                        Code = PlayerEventCode.Join1,
                        PlayerId = player.Id,
                        Tick = timer.NowTick,
                    });
            }
            else if (room.PlayerCount == 3) // Master + Player + Player
            {
                // 参加者が揃ったのでゲームを開始する
                //room.OnOtherPlayerLeft -= OnWaitingPlayerLeft;
                room.OnOtherPlayerJoined -= OnPlayerJoined;

                // 状態をInGameにし、入室も受け付けない
                var prop = new Dictionary<string, object>
                {
/*
                    {WSNet2Helper.PubKey.State, GameStateCode.InGame.ToString()},
                    {WSNet2Helper.PubKey.PlayerNum, (byte)2},
                    {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
*/
                };
                room.ChangeRoomProperty(joinable: false, publicProps: prop);

                newEvents.Add(new PlayerEvent()
                    {
                        Code = PlayerEventCode.Join2,
                        PlayerId = player.Id,
                        Tick = timer.NowTick,
                    });

                logger.Info("Game Start");
            }
        }

/*
        /// <summary>
        ///   二人目を待っている間に退室したときの処理
        /// </summary>
        void OnWaitingPlayerLeft(Player player)
        {
            room.OnOtherPlayerLeft -= OnWaitingPlayerLeft;
            logger.Info("Player left: {0}", player.Id);

            var prop = new Dictionary<string, object>
            {
                {WSNet2Helper.PubKey.PlayerNum, (byte)0},
                {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
            };
            room.ChangeRoomProperty(publicProps: prop);
        }
        */

        /// <summary>
        ///   IMasterClientの実装：Playerからの入力のRPCが届く
        /// </summary>
        /// <remarks>
        ///   GameUpdateと同一Taskで駆動されるのでロック不要
        /// </remarks>
        public void OnPlayerEvent(string sender, PlayerEvent ev)
        {
            ev.PlayerId = sender;
            newEvents.Add(ev);
        }

        /// <summary>
        ///   Gameロジックの駆動
        /// </summary>
        void GameUpdate()
        {
            // 前回のループから今回までの間にやってきた PlayerEvent が newEvents に格納されている.
            // 再計算可能なもののみを抽出する.
            long oldestTick = timer.NowTick;
            bool newEventExist = 0 < newEvents.Count;
            if (newEventExist)
            {
                foreach (var ev in newEvents)
                {
                    if (!room.Players.ContainsKey(ev.PlayerId))
                    {
                        // プレイヤー以外のイベントは無視する
                        continue;
                    }

                    if (states[0].Tick < ev.Tick)
                    {
                        events.Add(ev);
                        oldestTick = Math.Min(oldestTick, ev.Tick);
                    }
                    else
                    {
                        logger.Warning("Discard PlayerEvent: too past tick. Code:{0} Player:{1} ServerTick{2} EventTick:{3}",
                                       ev.Code, ev.PlayerId, states[0].Tick, ev.Tick); // TODO どうハンドルするべきか
                    }
                }
                events.Sort((a, b) => a.Tick.CompareTo(b.Tick));
                newEvents.Clear();
            }

            // 再計算可能な直近の GameState を探しつつ、それよりも新しいものは破棄する.
            while (oldestTick <= states[states.Count - 1].Tick)
            {
                states.RemoveAt(states.Count - 1);
            }

            var state = states[states.Count - 1].Copy();

            var now = timer.NowTick;
            var targetEvents = events.Where(ev => oldestTick <= ev.Tick && ev.Tick <= now);
            var tooFutureEvents = events.Where(ev => now < ev.Tick);

            if (0 < tooFutureEvents.Count())
            {
                foreach (var ev in tooFutureEvents)
                {
                    logger.Warning("Too future event. Room: {0} State: {1} Events: {2}", room.Id, state.Code.ToString(), targetEvents.Count());
                }
            }

            if (0 < targetEvents.Count())
            {
                logger.Debug("Room: {0} State: {1} Events: {2}", room.Id, state.Code.ToString(), targetEvents.Count());
            }

            var prevStateCode = state.Code;
            bool forceSync = sim.UpdateGame(now, state, targetEvents);

            if (prevStateCode != state.Code)
            {
                // ステートの更新が発生したので、以前の状態には戻さない
                states.Clear();
            }

            states.Add(state);

            if (50 < states.Count)
            {
                // 一番古い GameState を破棄する.
                // O(n) だが要素数少ないのでよいだろう
                states.RemoveAt(0);

                // 残ったもののうち一番古い State よりも古い PlayerEvent はもう復元に使えないので削除する.
                long t = states[0].Tick;
                int idx = events.FindIndex(ev => t < ev.Tick);
                if (idx != -1)
                {
                    events.RemoveRange(0, idx);
                }
            }

            // 0.1秒ごとにゲーム状態の同期メッセージを送信する
            if (forceSync || 100.0 <= new TimeSpan(now - lastSync).TotalMilliseconds)
            {
                rpc?.SyncServerTick(timer.NowTick);
                rpc?.SyncGameState(state);
                lastSync = now;
            }

            // ステートが更新されていたら public props に反映
            if (room != null && prevStateCode != state.Code)
            {
                if (state.Code == GameStateCode.End)
                {
                    // 終了ステートに変更したら部屋から抜ける
                    room.OnRoomPropertyChanged += (_,_,_,_,_,_,_,_) => room.Leave();
                }
                room.ChangeRoomProperty(publicProps: new Dictionary<string, object> {
                        {WSNet2Helper.PubKey.State, state.Code.ToString()}
                    });
            }
        }
    }
}
