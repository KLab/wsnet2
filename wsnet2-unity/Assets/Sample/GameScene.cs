using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.InputSystem;
using WSNet2.Core;
using Sample.Logic;

namespace Sample
{
    public class GameScene : MonoBehaviour
    {
        public Text roomText;
        public BallView ballAsset;
        public BarView barAsset;
        public float prevMoveInput;
        public InputAction moveInput;

        BarView bar1;
        BarView bar2;
        BallView ball;

        BarView playerBar;
        BarView opponentBar;

        GameSimulator simulator;
        GameState state;
        List<PlayerEvent> events;

        bool isOnlineMode;
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
                if (WSNet2Runner.Instance != null && WSNet2Runner.Instance.GameRoom != null)
                {
                    return WSNet2Runner.Instance.GameRoom.Me.Id;
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
            }
        }

        void Awake()
        {
            bar1 = Instantiate(barAsset);
            bar2 = Instantiate(barAsset);
            ball = Instantiate(ballAsset);

            bar1.gameObject.SetActive(false);
            bar2.gameObject.SetActive(false);
            ball.gameObject.SetActive(false);

            moveInput.Enable();

            simulator = new GameSimulator();
            state = new GameState();
            events = new List<PlayerEvent>();
            simulator.Init(state);
            isOnlineMode = WSNet2Runner.Instance != null && WSNet2Runner.Instance.GameRoom != null;
        }

        // Start is called before the first frame update
        void Start()
        {
            if (isOnlineMode)
            {
                var room = WSNet2Runner.Instance.GameRoom;
                roomText.text = "Room:" + room.Id + "\n";

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
                    RoomLog($"OnJoined: {p}");
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

                room.OnRoomPropertyChanged += (_, __) =>
                {
                    RoomLog("OnRoomPropertyChanged");
                };

                room.OnPlayerPropertyChanged += (p, _) =>
                {
                    RoomLog($"OnPlayerPropertyChanged: {p.Id}");
                };

                /// 使用するRPCを登録する
                /// MasterClientと同じ順番で同じRPCを登録する必要がある
                room.RegisterRPC<GameState>(RPCSyncGameState);
                room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
                room.Restart();
            }

            events.Add(new PlayerEvent
            {
                Code = PlayerEventCode.Join,
                PlayerId = myPlayerId,
            });

            if (!isOnlineMode)
            {
                events.Add(new PlayerEvent
                {
                    Code = PlayerEventCode.Join,
                    PlayerId = cpuPlayerId,
                });
            }
        }

        void Update()
        {
            Debug.Log(state.Code);
            if (state.Code == GameStateCode.WaitingPlayer)
            {
                if (Time.frameCount % 10 == 0)
                {
                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Join,
                        PlayerId = myPlayerId,
                    });
                }
            }
            else if (state.Code == GameStateCode.ReadyToStart)
            {
                if (Time.frameCount % 10 == 0)
                {
                    bar1.gameObject.SetActive(true);
                    bar2.gameObject.SetActive(true);
                    ball.gameObject.SetActive(true);

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

                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Ready,
                        PlayerId = myPlayerId,
                    });

                    if (!isOnlineMode)
                    {
                        events.Add(new PlayerEvent
                        {
                            Code = PlayerEventCode.Ready,
                            PlayerId = cpuPlayerId,
                        });
                    }
                }
            }
            else if (state.Code == GameStateCode.InGame)
            {

                var value = moveInput.ReadValue<float>();
                if (value != prevMoveInput)
                {
                    MoveInputCode move = MoveInputCode.Stop;
                    if (0 < value) move = MoveInputCode.Up;
                    if (value < 0) move = MoveInputCode.Down;

                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Move,
                        PlayerId = myPlayerId,
                        MoveInput = move,
                    });
                }
                prevMoveInput = value;
            }
            else if (state.Code == GameStateCode.Goal)
            {
                events.Add(new PlayerEvent
                {
                    Code = PlayerEventCode.Ready,
                    PlayerId = myPlayerId,
                });

                if (!isOnlineMode)
                {
                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Ready,
                        PlayerId = cpuPlayerId,
                    });
                }
            }

            // オンラインモードならイベントをRPCで送信
            // オフラインモードならローカルのシミュレータに入力
            if (isOnlineMode)
            {
                foreach (var ev in events)
                {
                    WSNet2Runner.Instance.GameRoom.RPC(RPCPlayerEvent, ev);
                }
            }
            else
            {
                Bar cpuBar = null;
                if (state.Player1 == cpuPlayerId) cpuBar = state.Bar1;
                if (state.Player2 == cpuPlayerId) cpuBar = state.Bar2;

                if (cpuBar != null)
                {
                    MoveInputCode move = MoveInputCode.Stop;
                    if (state.Ball.Position.y < cpuBar.Position.y) move = MoveInputCode.Up;
                    if (state.Ball.Position.y > cpuBar.Position.y) move = MoveInputCode.Down;

                    events.Add(new PlayerEvent
                    {
                        Code = PlayerEventCode.Move,
                        PlayerId = cpuPlayerId,
                        MoveInput = move,
                    });
                }

                simulator.UpdateGame(state, events);
            }

            events.Clear();

            // FIXME: オンラインモードだと同期するまでにズレが発生するはずなので、座標を補完する機能が必要.
            // オンラインモード時もローカルでもシミュレータ動作させる?
            if (state.Code == GameStateCode.InGame)
            {
                bar1.UpdatePosition(state.Bar1);
                bar2.UpdatePosition(state.Bar2);
                ball.UpdatePosition(state.Ball);
            }
        }
    }
}
