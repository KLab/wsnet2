using System;
using System.Linq;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;
using UnityEngine.InputSystem;
using Sample.Logic;

namespace Sample
{
    /// <summary>
    /// ゲームシーンのコントローラ
    /// </summary>
    public class GameScene : MonoBehaviour, IClient, IMasterClient
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

        GameSimulator simulator;
        GameState state;
        GameTimer timer;
        List<PlayerEvent> events;

        bool isOnlineMode;
        bool isWatcher;
        long lastSyncTick;
        float sinceEnd;
        bool closed;

        RPCBridge rpc;

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
            if (!roomText.IsDestroyed())
            {
                roomText.text += s + "\n";
            }

            Debug.LogFormat("Room {0}", s);
        }

        public void OnSyncGameState(GameState msg)
        {
            state = msg;
            events = events.Where(ev => state.Tick <= ev.Tick).ToList();
        }

        public void OnSyncServerTick(long tick)
        {
            timer.UpdateServerTick(tick);
            var ms = new TimeSpan(timer.NowTick - tick).TotalMilliseconds;
            if (64 <= ms)
            {
                Debug.LogWarningFormat("Packet jam {0}ms", ms);
            }
        }

        public void OnPlayerEvent(string sender, PlayerEvent ev)
        {
            if (simulator.IsMaster)
            {
                events.Add(ev);
            }
        }

        void Awake()
        {
            Application.targetFrameRate = 60;

            bar1 = Instantiate(barAsset);
            bar2 = Instantiate(barAsset);
            balls = new List<BallView>();

            bar1.gameObject.SetActive(false);
            bar2.gameObject.SetActive(false);

            moveInput.Enable();

            isOnlineMode = G.GameRoom != null;
            isWatcher = isOnlineMode && G.GameRoom.Me == null;
            simulator = new GameSimulator(!isOnlineMode || G.GameRoom.Master == G.GameRoom.Me);
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
                    G.GameRoom = null;
                    closed = true;
                };

                room.OnJoined += (me) =>
                {
                    RoomLog($"OnJoined: {me.Id}");

                    if (room.Master == room.Me)
                    {
                        room.ChangeRoomProperty(publicProps: new Dictionary<string, object> {
                             { WSNet2Helper.PubKey.PlayerNum, (byte)room.PlayerCount},
                             { WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                         });
                    }
                };

                room.OnClosed += (p) =>
                {
                    RoomLog($"OnClosed: {p}");
                    G.GameRoom = null;
                    closed = true;
                };

                room.OnOtherPlayerJoined += (p) =>
                {
                    RoomLog("OnOtherPlayerJoined:" + p.Id);

                    if (room.Master == room.Me)
                    {
                        room.ChangeRoomProperty(publicProps: new Dictionary<string, object> {
                             { WSNet2Helper.PubKey.PlayerNum, (byte)room.PlayerCount},
                             { WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                         });
                    }
                };

                room.OnOtherPlayerLeft += (p, msg) =>
                {
                    RoomLog("OnOtherPlayerLeft:" + p.Id + " msg:" + msg);

                    if (room.Master == room.Me)
                    {
                        room.ChangeRoomProperty(publicProps: new Dictionary<string, object> {
                             { WSNet2Helper.PubKey.PlayerNum, (byte)room.PlayerCount},
                             { WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                         });
                    }
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

                if (simulator.IsMaster)
                {
                    rpc = new RPCBridge(room, (IMasterClient)this);
                }
                else
                {
                    rpc = new RPCBridge(room, (IClient)this);
                }

                room.Restart();
            }
        }

        void Update()
        {
            if (closed)
            {
                sinceEnd += Time.deltaTime;
                if (3.0 < sinceEnd)
                {
                    SceneManager.LoadScene("Title");
                }

                return;
            }

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

                if (G.GameRoom != null)
                {
                    G.GameRoom.Leave();
                }
                else
                {
                    closed = true;
                }
            }

            if (state.Code == GameStateCode.WaitingPlayer)
            {
                if (Time.frameCount % Application.targetFrameRate == 0)
                {
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
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Join,
                            PlayerId = cpuPlayerId,
                            Tick = timer.NowTick,
                        });
                    }
                }
            }

            if (state.Code == GameStateCode.ReadyToStart)
            {
                if (Time.frameCount % Application.targetFrameRate == 0)
                {
                    bar1.gameObject.SetActive(true);
                    bar2.gameObject.SetActive(true);

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

            if (state.Code == GameStateCode.InGame)
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

                        prevMoveInput = value;
                    }
                }
            }

            if (isOnlineMode)
            {
                if (G.GameRoom == null)
                {
                    // タイトルシーンに戻る際にここに到達する可能性がある
                    return;
                }

                if (simulator.IsMaster)
                {
                    // 自分がマスタクライアントの場合相手にstateを同期する
                    var prevState = state.Code;
                    var forceSync = simulator.UpdateGame(timer.NowTick, state, events);
                    if (forceSync || 100 <= new TimeSpan(timer.NowTick - lastSyncTick).TotalMilliseconds)
                    {
                        rpc.SyncServerTick(timer.NowTick);
                        rpc.SyncGameState(state);
                        lastSyncTick = timer.NowTick;
                    }

                    if (prevState != state.Code)
                    {
                        var room = G.GameRoom;
                        room.ChangeRoomProperty(publicProps: new Dictionary<string, object> {
                            { WSNet2Helper.PubKey.PlayerNum, (byte)room.PlayerCount},
                            { WSNet2Helper.PubKey.State, state.Code.ToString()},
                            { WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
                        });
                    }
                }
                else
                {
                    // イベントをRPCで送信する
                    foreach (var ev in events)
                    {
                        rpc.PlayerEvent(ev);
                    }

                    // ローカルシミュレータを投機的に更新する
                    simulator.UpdateGame(timer.NowTick, state, events);
                }
            }
            else
            {
                // ローカルシミュレータを更新する
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
                bar1.gameObject.SetActive(true);
                bar2.gameObject.SetActive(true);
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
