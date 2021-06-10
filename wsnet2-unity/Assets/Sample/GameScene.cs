using System;
using System.Linq;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;
using UnityEngine.InputSystem;
using WSNet2.Core;
using Sample.Logic;

namespace Sample
{
    /// <summary>
    /// ゲームシーンのコントローラ
    /// </summary>
    public class GameScene : MonoBehaviour
    {
        /// <summary>
        /// 画面背景の文字
        /// </summary>
        public Text roomText;

        /// <summary>
        /// Player1側の文字
        /// </summary>
        public Text playerText1;

        /// <summary>
        /// Player2側の文字
        /// </summary>
        public Text playerText2;

        /// <summary>
        /// ボールのアセット
        /// </summary>
        public BallView ballAsset;

        /// <summary>
        /// バーのアセット
        /// </summary>
        public BarView barAsset;

        /// <summary>
        /// 1fr前の入力
        /// </summary>
        public float prevMoveInput;

        /// <summary>
        /// 移動入力
        /// </summary>
        public InputAction moveInput;

        BarView bar1;
        BarView bar2;
        List<BallView> balls;

        BarView playerBar;
        BarView opponentBar;

        GameSimulator simulator;
        GameState state;
        GameTimer timer;
        List<PlayerEvent> events;

        bool isOnlineMode;
        bool isWatcher;
        float nextSyncTime;

        string cpuPlayerId
        {
            get
            {
                return "CPU";
            }
        }

        string myPlayerId
        {
            get
            {
                if (G.GameRoom != null)
                {
                    return G.GameRoom.Me?.Id;
                }
                else
                {
                    return "YOU";
                }
            }
        }

        void RoomLog(string s)
        {
            roomText.text += s + "\n";
            Debug.LogFormat("Room {0}", s);
        }

        void RPCPlayerEvent(string sender, PlayerEvent msg)
        {
            // only master client handle this.
        }

        void RPCSyncGameState(string sender, GameState msg)
        {
            // TODO: How to check if the sender is valid master client?
            if (msg.MasterId == sender)
            {
                state = msg;
                events = events.Where(ev => state.Tick <= ev.Tick).ToList();
            }
        }

        void Awake()
        {
            bar1 = Instantiate(barAsset);
            bar2 = Instantiate(barAsset);
            balls = new List<BallView>();

            bar1.gameObject.SetActive(false);
            bar2.gameObject.SetActive(false);

            moveInput.Enable();

            isOnlineMode = G.GameRoom != null;
            isWatcher = isOnlineMode && G.GameRoom.Me == null;
            simulator = new GameSimulator(!isOnlineMode);
            state = new GameState();
            timer = new GameTimer();
            events = new List<PlayerEvent>();
            simulator.Init(state);
        }

        // Start is called before the first frame update
        void Start()
        {
            if (isOnlineMode)
            {
                var room = G.GameRoom;
                roomText.text = "Room:" + room.Id + "\n";
                RoomLog($"isOnlineMode: {isOnlineMode} isWatcher: {isWatcher}");

                room.OnError += (e) =>
                {
                    Debug.LogError(e.ToString());
                    RoomLog($"OnError: {e}");
                };

                room.OnErrorClosed += (e) =>
                 {
                     Debug.LogError(e.ToString());
                     RoomLog($"OnErrorClosed: {e}");
                 };

                room.OnJoined += (me) =>
                {
                    RoomLog($"OnJoined: {me.Id}");
                };

                room.OnClosed += (p) =>
                {
                    RoomLog($"OnClosed: {p}");
                    G.GameRoom = null;
                    SceneManager.LoadScene("Title");
                };

                room.OnOtherPlayerJoined += (p) =>
                 {
                     RoomLog("OnOtherPlayerJoined:" + p.Id);
                 };

                room.OnOtherPlayerLeft += (p) =>
                {
                    RoomLog("OnOtherPlayerLeft:" + p.Id);
                };

                room.OnMasterPlayerSwitched += (prev, cur) =>
                {
                    RoomLog("OnMasterPlayerSwitched:" + prev.Id + " -> " + cur.Id);
                };

                room.OnPlayerPropertyChanged += (p, _) =>
                {
                    RoomLog($"OnPlayerPropertyChanged: {p.Id}");
                };

                room.OnRoomPropertyChanged += (visible, joinable, watchable, searchGroup, maxPlayers, clientDeadline, publicProps, privateProps) =>
                {
                    RoomLog($"OnRoomPropertyChanged");
                    if (publicProps != null)
                    {
                        foreach (var kv in publicProps)
                        {
                            Debug.LogFormat("(public) {0}:{1}", kv.Key, kv.Value.ToString());
                        }
                    }
                    if (privateProps != null)
                    {
                        foreach (var kv in privateProps)
                        {
                            Debug.LogFormat("(private) {0}:{1}", kv.Key, kv.Value.ToString());
                        }
                    }
                };


                var RPCSyncServerTick = new Action<string, long>((sender, tick) =>
                {
                    if (sender == G.GameRoom?.Master.Id)
                    {
                        timer.UpdateServerTick(tick);
                        var ms = new TimeSpan(timer.NowTick - tick).TotalMilliseconds;
                        if (64 <= ms)
                        {
                            Debug.LogWarningFormat("Packet jam {0}ms", ms);
                        }
                    }
                });

                /// 使用するRPCを登録する
                /// MasterClientと同じ順番で同じRPCを登録する必要がある
                room.RegisterRPC<GameState>(RPCSyncGameState);
                room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
                room.RegisterRPC(RPCSyncServerTick);
                room.Restart();
            }

            if (!isWatcher)
            {
                events.Add(new PlayerEvent
                {
                    Code = PlayerEventCode.Join,
                    PlayerId = myPlayerId,
                    Tick = timer.NowTick,
                });
            }

            if (!isOnlineMode)
            {
                // オフラインモードのときは WaitingPlayer から始める
                state.Code = GameStateCode.WaitingPlayer;
                events.Add(new PlayerEvent
                {
                    Code = PlayerEventCode.Join,
                    PlayerId = cpuPlayerId,
                    Tick = timer.NowTick,
                });
            }
        }

        void Update()
        {
            playerText1.text = $"Name: {state.Player1}\n Score: {state.Score1}\n";
            playerText2.text = $"Name: {state.Player2}\n Score: {state.Score2}\n";
            if (state.Code == GameStateCode.End)
            {
                if (state.Score2 < state.Score1)
                {
                    playerText1.text += $"\n WIN";
                    playerText2.text += $"\n LOSE";
                }
                else if (state.Score1 < state.Score2)
                {
                    playerText1.text += $"\n LOSE";
                    playerText2.text += $"\n WIN";
                }
                else
                {
                    playerText1.text += $"\n DRAW";
                    playerText2.text += $"\n DRAW";
                }
            }

            if (state.Code == GameStateCode.WaitingGameMaster)
            {
                if (Time.frameCount % 10 == 0)
                {
                    var room = G.GameRoom;
                    // 本当はルームマスタがルームを作成するシーケンスを想定しているが, サンプルは簡単のため,
                    // マスタークライアントがJoinしてきたら, ルームマスタを委譲する
                    if (room.Me == room.Master)
                    {
                        foreach (var p in room.Players.Values)
                        {
                            if (p.Id.StartsWith("gamemaster"))
                            {
                                RoomLog("Switch master to" + p.Id);
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
            }
            else if (state.Code == GameStateCode.WaitingPlayer)
            {
                if (!isWatcher && Time.frameCount % 10 == 0)
                {
                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Join,
                        PlayerId = myPlayerId,
                        Tick = timer.NowTick,
                    });
                }
            }
            else if (state.Code == GameStateCode.ReadyToStart)
            {
                if (Time.frameCount % 10 == 0)
                {
                    bar1.gameObject.SetActive(true);
                    bar2.gameObject.SetActive(true);

                    if (state.Player1 == myPlayerId)
                    {
                        playerBar = bar1;
                        opponentBar = bar2;
                    }
                    if (state.Player2 == myPlayerId)
                    {
                        playerBar = bar2;
                        opponentBar = bar1;
                    }

                    if (!isWatcher)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Ready,
                            PlayerId = myPlayerId,
                            Tick = timer.NowTick,
                        });
                    }

                    if (!isOnlineMode)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Ready,
                            PlayerId = cpuPlayerId,
                            Tick = timer.NowTick,
                        });
                    }
                }
            }
            else if (state.Code == GameStateCode.InGame)
            {
                if (!isWatcher)
                {
                    var value = moveInput.ReadValue<float>();
                    if (value != prevMoveInput)
                    {
                        MoveInputCode move = MoveInputCode.Stop;
                        if (0 < value) move = MoveInputCode.Down;
                        if (value < 0) move = MoveInputCode.Up;

                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Move,
                            PlayerId = myPlayerId,
                            MoveInput = move,
                            Tick = timer.NowTick,
                        });
                    }
                    prevMoveInput = value;
                }
            }
            else if (state.Code == GameStateCode.End)
            {
                return;
            }

            // オンラインモードならイベントをRPCで送信
            // オフラインモードならローカルのシミュレータに入力
            if (isOnlineMode)
            {
                if (G.GameRoom == null)
                {
                    // タイトルシーンに戻る際にここに到達する可能性がある
                    return;
                }

                foreach (var ev in events)
                {
                    G.GameRoom.RPC(RPCPlayerEvent, ev);
                }
                simulator.UpdateGame(timer.NowTick, state, events.Where(ev => state.Tick <= ev.Tick));
            }
            else
            {
                Bar cpuBar = null;
                if (state.Player1 == cpuPlayerId) cpuBar = state.Bar1;
                if (state.Player2 == cpuPlayerId) cpuBar = state.Bar2;

                if (cpuBar != null)
                {
                    MoveInputCode move = MoveInputCode.Stop;
                    if (state.Balls[0].Position.y < cpuBar.Position.y) move = MoveInputCode.Up;
                    if (state.Balls[0].Position.y > cpuBar.Position.y) move = MoveInputCode.Down;

                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Move,
                        PlayerId = cpuPlayerId,
                        MoveInput = move,
                        Tick = timer.NowTick,
                    });
                }
                simulator.UpdateGame(timer.NowTick, state, events);
            }

            if (state.Code == GameStateCode.InGame)
            {
                bar1.UpdatePosition(state.Bar1);
                bar2.UpdatePosition(state.Bar2);

                // ボールの数を揃える
                while (state.Balls.Count < balls.Count)
                {
                    Destroy(balls[balls.Count - 1].gameObject);
                    balls.RemoveAt(balls.Count - 1);
                }
                while (balls.Count < state.Balls.Count)
                {
                    balls.Add(Instantiate(ballAsset));
                }

                for (int i = 0; i < state.Balls.Count; ++i)
                {
                    balls[i].UpdatePosition(state.Balls[i]);
                }
            }
            events.Clear();
        }
    }
}
