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
        /// <summary>1フレームの時間</summary>
        const int frameMilliSec = 1000/60;

        string userId;
        string pKey;
        AuthDataGenerator authgen;
        WSNet2Client client;

        Room room;
        RPCBridge rpc;
        GameTimer timer;
        GameSimulator simulator;
        GameState state;
        List<PlayerEvent> events;
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

            simulator = new GameSimulator(true);
            timer = new GameTimer();
            events = new List<PlayerEvent>();
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
                    state = new GameState();
                    state.Tick = timer.NowTick;
                    state.MasterId = userId;
                    simulator.Init(state);

                    CreateRoom(cts);
                    await Updater(cts.Token);
                }
                catch (OperationCanceledException)
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

            // clientを々利用するため、使う直前に認証情報を更新
            client.UpdateAuthData(authgen.Generate(pKey, userId));

            // 部屋を検索可能、入室可能、観戦可能として作成する
            var roomOpt = new RoomOption(MaxPlayers, WSNet2Helper.SearchGroup, pubProps, null)
                .Visible(true).Joinable(true).Watchable(true);

            client.Create(
                roomOpt,
                null,
                (room) =>
                {
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
                    room.OnOtherPlayerJoined += OnPlayerJoined;
                    room.OnOtherPlayerLeft += OnPlayerLeft;
                    this.room = room;
                    this.rpc = new RPCBridge(room, this);
                },
                (e) => throw e,
                logger);
        }

        /// <summary>
        ///   プレイヤーが入室したときの処理
        /// </summary>
        void OnPlayerJoined(Player player)
        {
            if (GameStateCode.ReadyToStart <= state.Code || 4 <= room.PlayerCount)
            {
                // 予定外の人が入室してきた
                room.Kick(player);
                return;
            }

            if (room.PlayerCount == 2) // Master + Player
            {
                room.ChangeRoomProperty(joinable: true, publicProps: new Dictionary<string, object>
                {
                    {WSNet2Helper.PubKey.State, GameStateCode.WaitingPlayer.ToString()},
                    {WSNet2Helper.PubKey.PlayerNum, room.PlayerCount - 1},
                    {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                });
            }
            else if (room.PlayerCount == 3) // Master + Player + Player
            {
                // 参加者が揃ったので募集を締め切る
                room.ChangeRoomProperty(joinable: false, publicProps: new Dictionary<string, object>
                {
                    {WSNet2Helper.PubKey.State, GameStateCode.InGame.ToString()},
                    {WSNet2Helper.PubKey.PlayerNum, room.PlayerCount - 1},
                    {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                });

                // 先に入室した方を1Pとする
                var joiner = room.Players.Where((p) => p.Value.Id != userId).ToArray();
                if (joiner[0].Value.Id == player.Id) {
                    (joiner[0], joiner[1]) = (joiner[1], joiner[0]);
                }

                events.Add(new PlayerEvent()
                {
                    Code = PlayerEventCode.Join,
                    PlayerId = joiner[0].Value.Id,
                    Tick = timer.NowTick,
                });

                events.Add(new PlayerEvent()
                {
                    Code = PlayerEventCode.Join,
                    PlayerId = joiner[1].Value.Id,
                    Tick = timer.NowTick,
                });

                logger.Info("Game Start");
            }
        }

        /// <summary>
        ///   プレイヤーが退室したときの処理
        /// </summary>
        void OnPlayerLeft(Player player, string msg)
        {
            logger.Info($"OnPlayerLeft player={player.Id} msg={msg}");

            if (state.Code < GameStateCode.ReadyToStart)
            {
                // まだゲームを開始していないので、募集を再開する
                room.ChangeRoomProperty(joinable: true, publicProps: new Dictionary<string, object>
                {
                    {WSNet2Helper.PubKey.State, GameStateCode.WaitingPlayer.ToString()},
                    {WSNet2Helper.PubKey.PlayerNum, (byte)room.PlayerCount - 1},
                    {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                });
            }
            else
            {
                // ゲーム中にプレイヤーが退室してしまった
                // 退室して終了する
                room.Leave($"Player {player} is disconnected");
            }
        }

        /// <summary>
        ///   IMasterClientの実装：Playerからの入力のRPCが届く
        /// </summary>
        /// <remarks>
        ///   GameUpdateと同一Taskで駆動されるのでロック不要
        /// </remarks>
        public void OnPlayerEvent(string sender, PlayerEvent ev)
        {
            ev.PlayerId = sender;
            events.Add(ev);
        }

        /// <summary>
        ///   Gameロジックの駆動
        /// </summary>
        void GameUpdate()
        {
            // 前回のループから今回までの間にやってきた PlayerEvent が events に格納されている
            // これを使ってゲーム状態を更新する
            var now = timer.NowTick;
            var prevStateCode = state.Code;
            events.Sort((a, b) => (int)(a.Tick - b.Tick));
            bool forceSync = simulator.UpdateGame(now, state, events);
            events.Clear();

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
                room.ChangeRoomProperty(publicProps: new Dictionary<string, object>
                {
                    {WSNet2Helper.PubKey.State, state.Code.ToString()}
                });
            }

            // public props のステートがEndになっているのを確認したら部屋から抜ける
            if (room.GameState() == GameStateCode.End)
            {
                room.Leave();
            }
        }
    }
}
