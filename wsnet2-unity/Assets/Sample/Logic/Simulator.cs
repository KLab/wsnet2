using System;
using System.Collections.Generic;
using WSNet2.Core;

#if UNITY_5_3_OR_NEWER
using Vector2 = UnityEngine.Vector2;
using Vector3 = UnityEngine.Vector3;
#endif

namespace Sample.Logic
{
    /// <summary>
    /// Pongゲームの状態を表すデータ
    /// </summary>
    public class GameState : IWSNetSerializable
    {
        /// <summary>
        /// ゲームの状態の種類
        /// </summary>
        public GameStateCode Code;

        /// <summary>
        /// 何回ゲームを遊んだか
        /// </summary>
        public int GameCount;

        /// <summary>
        /// マスタークライアントのプレイヤーID
        /// </summary>
        public string MasterId;

        /// <summary>
        /// 1PのプレイヤーID
        /// </summary>
        public string Player1;

        /// <summary>
        /// 2PのプレイヤーID
        /// </summary>
        public string Player2;

        /// <summary>
        /// 1Pが準備完了しているか
        /// </summary>
        public int Player1Ready;

        /// <summary>
        /// 2Pが準備完了しているか
        /// </summary>
        public int Player2Ready;

        /// <summary>
        /// 1Pのスコア
        /// </summary>
        public int Score1;

        /// <summary>
        /// 2Pのスコア
        /// </summary>
        public int Score2;

        /// <summary>
        /// 1Pのバーの状態
        /// </summary>
        public Bar Bar1;

        /// <summary>
        /// 2Pのバーの状態
        /// </summary>
        public Bar Bar2;

        /// <summary>
        /// ボールの状態
        /// </summary>
        public Ball Ball;

        public void Serialize(SerialWriter writer)
        {
            writer.Write((int)Code);
            writer.Write(GameCount);
            writer.Write(MasterId);
            writer.Write(Player1);
            writer.Write(Player2);
            writer.Write(Player1Ready);
            writer.Write(Player2Ready);
            writer.Write(Score1);
            writer.Write(Score2);
            writer.Write(Bar1);
            writer.Write(Bar2);
            writer.Write(Ball);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            Code = (GameStateCode)reader.ReadInt();
            GameCount = reader.ReadInt();
            MasterId = reader.ReadString();
            Player1 = reader.ReadString();
            Player2 = reader.ReadString();
            // Player1Ready = reader.ReadBool();
            // Player2Ready = reader.ReadBool();
            Player1Ready = reader.ReadInt();
            Player2Ready = reader.ReadInt();
            Score1 = reader.ReadInt();
            Score2 = reader.ReadInt();
            Bar1 = reader.ReadObject(Bar1);
            Bar2 = reader.ReadObject(Bar2);
            Ball = reader.ReadObject(Ball);
        }
    }

    /// <summary>
    /// ゲーム状態を表す定数
    /// </summary>
    public enum GameStateCode
    {

        /// <summary>
        /// 未使用
        /// </summary>
        None,

        /// <summary>
        /// マスタークライアントが参加するのを待っている
        /// </summary>
        WaitingGameMaster,

        /// <summary>
        /// プレイヤーが参加するのを待っている
        /// </summary>
        WaitingPlayer,

        /// <summary>
        /// プレイヤーの準備完了を待っている
        /// </summary>
        ReadyToStart,

        /// <summary>
        /// ゲーム進行中
        /// </summary>
        InGame,

        /// <summary>
        /// ゴール演出 プレイヤーの準備完了を待っている
        /// </summary>
        Goal,
    }

    /// <summary>
    /// ボールのデータ
    /// </summary>
    public class Ball : IWSNetSerializable
    {
        /// <summary>
        /// 位置
        /// </summary>
        public Vector2 Position;

        /// <summary>
        /// 移動方向
        /// </summary>
        public Vector2 Direction;

        /// <summary>
        /// 速さ
        /// </summary>
        public float Speed;

        /// <summary>
        /// 半径
        /// </summary>
        public float Radius;

        public void Serialize(SerialWriter writer)
        {
            writer.Write(Position.x);
            writer.Write(Position.y);
            writer.Write(Direction.x);
            writer.Write(Direction.y);
            writer.Write(Speed);
            writer.Write(Radius);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            Position.x = reader.ReadFloat();
            Position.y = reader.ReadFloat();
            Direction.x = reader.ReadFloat();
            Direction.y = reader.ReadFloat();
            Speed = reader.ReadFloat();
            Radius = reader.ReadFloat();
        }
    }

    /// <summary>
    /// バーのデータ
    /// </summary>
    public class Bar : IWSNetSerializable
    {
        /// <summary>
        /// 位置
        /// </summary>
        public Vector2 Position;

        /// <summary>
        /// 移動方向
        /// </summary>
        public Vector2 Direction;

        /// <summary>
        /// 移動スピード
        /// </summary>
        public float Speed;

        /// <summary>
        /// 幅
        /// </summary>
        public float Width;

        /// <summary>
        /// 高さ
        /// </summary>
        public float Height;

        public void Serialize(SerialWriter writer)
        {
            writer.Write(Position.x);
            writer.Write(Position.y);
            writer.Write(Direction.x);
            writer.Write(Direction.y);
            writer.Write(Speed);
            writer.Write(Width);
            writer.Write(Height);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            Position.x = reader.ReadFloat();
            Position.y = reader.ReadFloat();
            Direction.x = reader.ReadFloat();
            Direction.y = reader.ReadFloat();
            Speed = reader.ReadFloat();
            Width = reader.ReadFloat();
            Height = reader.ReadFloat();
        }
    }

    /// <summary>
    /// プレイヤー入力イベントの種類を表す定数
    /// </summary>
    public enum PlayerEventCode
    {
        /// <summary>
        /// 未使用
        /// </summary>
        None,

        /// <summary>
        /// ゲームに参加表明
        /// </summary>
        Join,

        /// <summary>
        /// ゲーム開始準備完了
        /// </summary>
        Ready,

        /// <summary>
        /// 移動入力
        /// </summary>
        Move,
    }

    /// <summary>
    /// プレイヤーの移動入力の種類を表す定数
    /// </summary>
    public enum MoveInputCode
    {
        /// <summary>
        /// 停止
        /// </summary>
        Stop,

        /// <summary>
        /// 上移動
        /// </summary>
        Up,
        
        /// <summary>
        /// 下移動
        /// </summary>
        Down,
    }

    /// <summary>
    /// プレイヤー入力を表す
    /// Code の種類によって入力の種類を識別する
    /// </summary>
    public class PlayerEvent : IWSNetSerializable
    {
        /// <summary>
        /// プレイヤー入力の種類
        /// </summary>
        public PlayerEventCode Code;

        /// <summary>
        /// イベントを発生させたプレイヤーID
        /// </summary>
        /// <remarks>
        /// オンラインモードでは自己申告ではなくマスタークライアントが送信元のユーザIDで上書きすること
        /// </remarks>
        public string PlayerId;

        /// <summary>
        /// 移動入力の種類
        /// </summary>
        public MoveInputCode MoveInput;

        public void Serialize(SerialWriter writer)
        {
            writer.Write((int)Code);
            writer.Write(PlayerId);
            writer.Write((int)MoveInput);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            Code = (PlayerEventCode)reader.ReadInt();
            PlayerId = string.Intern(reader.ReadString());
            MoveInput = (MoveInputCode)reader.ReadInt();
        }
    }

    /// <summary>
    /// Pongゲームのシミュレータ
    /// </summary>
    public class GameSimulator
    {
        public int BoardWidth { get; private set; }
        public int BoardHeight { get; private set; }
        public Vector2 Center { get; private set; }
        public int MinX { get; private set; }
        public int MaxX { get; private set; }
        public int MinY { get; private set; }
        public int MaxY { get; private set; }

        /// <summary>
        /// stateを初期状態に設定する
        /// </summary>
        /// <param name="state"></param>
        public void Init(GameState state)
        {
            BoardWidth = 960;
            BoardHeight = 540;
            Center = new Vector2(BoardWidth / 2f, BoardHeight / 2f);
            MinX = -BoardWidth / 2;
            MaxX = BoardWidth / 2;
            MinY = -BoardHeight / 2;
            MaxY = BoardHeight / 2;

            state.Code = GameStateCode.WaitingGameMaster;
            state.GameCount = 0;
            state.Player1 = "";
            state.Player2 = "";
            state.Player1Ready = 0;
            state.Player2Ready = 0;
            state.Score1 = 0;
            state.Score2 = 0;
            state.Bar1 = new Bar();
            state.Bar2 = new Bar();
            state.Ball = new Ball();
        }

        void ResetPositions(GameState state)
        {
            state.Bar1.Width = 10;
            state.Bar1.Height = 200;
            state.Bar1.Position.x = MinX + 50;
            state.Bar1.Position.y = 0;
            state.Bar1.Speed = 3f;

            state.Bar2.Width = 10;
            state.Bar2.Height = 200;
            state.Bar2.Position.x = MaxX - 50;
            state.Bar2.Position.y = 0;
            state.Bar2.Speed = 3f;

            var rnd = new Random();
            var x = (double)rnd.Next(-1000, 1000);
            var y = (double)rnd.Next(-100, 100);
            var s = Math.Sqrt(x * x + y * y);
            state.Ball.Position.x = 0;
            state.Ball.Position.y = 0;
            state.Ball.Direction.x = (float)(x / s);
            state.Ball.Direction.y = (float)(y / s);
            state.Ball.Speed = 3f;
            state.Ball.Radius = 10;
        }


        void StartNextGame(GameState state)
        {
            if (string.IsNullOrEmpty(state.Player1)) throw new Exception("Player1 is not joind");
            if (string.IsNullOrEmpty(state.Player2)) throw new Exception("Player2 is not joind");

            state.Code = GameStateCode.InGame;
            if (state.GameCount == 0)
            {
                state.Score1 = 0;
                state.Score2 = 0;
            }
            state.GameCount++;

            ResetPositions(state);
        }

        /// <summary>
        /// ゲームを1ステップ進める
        /// </summary>
        /// <param name="state">ゲームステート</param>
        /// <param name="events">プレイヤーの入力イベント</param>
        public void UpdateGame(GameState state, IEnumerable<PlayerEvent> events)
        {
            if (state.Code == GameStateCode.WaitingGameMaster)
            {
                // マスタクライアントの参加を待っている
                // 何もしない
            }
            if (state.Code == GameStateCode.WaitingPlayer)
            {
                // プレイヤーの参加を待っている
                // 先に参加したプレイヤーを1P, あとに参加したプレイヤーを2Pとする
                if (events != null)
                {
                    foreach (var ev in events)
                    {
                        if (ev.Code == PlayerEventCode.Join)
                        {
                            if (ev.PlayerId == state.Player1 || ev.PlayerId == state.Player2)
                            {
                                continue;
                            }

                            if (string.IsNullOrEmpty(state.Player1))
                            {
                                state.Player1 = ev.PlayerId;
                            }
                            else if (string.IsNullOrEmpty(state.Player2))
                            {
                                state.Player2 = ev.PlayerId;
                                // プレイヤーが集まったのでスタート準備へ
                                state.Code = GameStateCode.ReadyToStart;
                            }
                        }
                    }
                }
            }
            else if (state.Code == GameStateCode.ReadyToStart || state.Code == GameStateCode.Goal)
            {
                // 1P, 2Pが Ready 入力を送ってくるのを待っている
                if (events != null)
                {
                    foreach (var ev in events)
                    {
                        if (ev.PlayerId == state.Player1 && ev.Code == PlayerEventCode.Ready)
                        {
                            state.Player1Ready = 1;
                        }
                        if (ev.PlayerId == state.Player2 && ev.Code == PlayerEventCode.Ready)
                        {
                            state.Player2Ready = 1;
                        }

                        if (state.Player1Ready == 1 && state.Player2Ready == 1)
                        {
                            // 準備ができたのでゲーム開始
                            StartNextGame(state);
                        }
                    }
                }
            }
            else if (state.Code == GameStateCode.InGame)
            {
                if (events != null)
                {
                    foreach (var ev in events)
                    {
                        // 移動入力を処理する
                        if (ev.Code == PlayerEventCode.Move)
                        {
                            float dirY = 0;
                            switch (ev.MoveInput)
                            {
                                case MoveInputCode.Stop: dirY = 0; break;
                                case MoveInputCode.Up: dirY = -1; break;
                                case MoveInputCode.Down: dirY = 1; break;
                            }

                            if (ev.PlayerId == state.Player1)
                            {
                                state.Bar1.Direction.y = dirY;
                            }
                            if (ev.PlayerId == state.Player2)
                            {
                                state.Bar2.Direction.y = dirY;
                            }
                        }
                    }
                }

                // 1Pのバーの移動
                state.Bar1.Position.x = Math.Min(MaxX - state.Bar1.Width / 2, Math.Max(MinX + state.Bar1.Width / 2, state.Bar1.Position.x + state.Bar1.Direction.x * state.Bar1.Speed));
                state.Bar1.Position.y = Math.Min(MaxY - state.Bar1.Height / 2, Math.Max(MinY + state.Bar1.Height / 2, state.Bar1.Position.y + state.Bar1.Direction.y * state.Bar1.Speed));

                // 2Pのバーの移動
                state.Bar2.Position.x = Math.Min(MaxX - state.Bar2.Width / 2, Math.Max(MinX + state.Bar2.Width / 2, state.Bar2.Position.x + state.Bar2.Direction.x * state.Bar2.Speed));
                state.Bar2.Position.y = Math.Min(MaxY - state.Bar2.Height / 2, Math.Max(MinY + state.Bar2.Height / 2, state.Bar2.Position.y + state.Bar2.Direction.y * state.Bar2.Speed));

                // ボールの移動
                state.Ball.Position.x = state.Ball.Position.x + state.Ball.Direction.x * state.Ball.Speed;
                state.Ball.Position.y = state.Ball.Position.y + state.Ball.Direction.y * state.Ball.Speed;

                // ボールが上下の壁にあたってたら移動方向を反射
                if (state.Ball.Position.y < MinY + state.Ball.Radius || state.Ball.Position.y > MaxY - state.Ball.Radius)
                {
                    state.Ball.Direction.y *= -1;
                }

                // ボールが壁にめり込んでいたら戻す
                state.Ball.Position.y = Math.Min(MaxY - state.Ball.Radius, Math.Max(MinY + state.Ball.Radius, state.Ball.Position.y));

                // 1Pのバーにボールが当たってたら反射
                {
                    var bx = state.Ball.Position.x;
                    var by = state.Ball.Position.y;
                    var br = state.Ball.Radius;
                    var px = state.Bar1.Position.x;
                    var py = state.Bar1.Position.y;
                    var pw = state.Bar1.Width;
                    var ph = state.Bar1.Height;

                    if (bx - br <= px + pw / 2f && bx + br >= px + pw / 2f)
                    {
                        if (py - ph / 2f <= by && py + ph / 2f >= by)
                        {
                            if (state.Ball.Direction.x < 0)
                            {
                                state.Ball.Direction.x *= -1;
                            }
                        }
                    }
                }

                // 2Pのバーにボールが当たってたら反射
                {
                    var bx = state.Ball.Position.x;
                    var by = state.Ball.Position.y;
                    var br = state.Ball.Radius;
                    var px = state.Bar2.Position.x;
                    var py = state.Bar2.Position.y;
                    var pw = state.Bar2.Width;
                    var ph = state.Bar2.Height;

                    if (bx - br <= px + pw / 2f && bx + br >= px + pw / 2f)
                    {
                        if (py - ph / 2f <= by && py + ph / 2f >= by)
                        {
                            if (state.Ball.Direction.x > 0)
                            {
                                state.Ball.Direction.x *= -1;
                            }
                        }
                    }
                }

                // 1Pのゴールに入っていたら2Pに得点
                if (state.Ball.Position.x < MinX + state.Ball.Radius)
                {
                    state.Code = GameStateCode.Goal;
                    state.Score2 += 1;
                    state.Player1Ready = 0;
                    state.Player2Ready = 0;
                }

                // 2Pのゴールに入っていたら1Pに得点
                if (MaxX + state.Ball.Radius < state.Ball.Position.x)
                {
                    state.Code = GameStateCode.Goal;
                    state.Score1 += 1;
                    state.Player1Ready = 0;
                    state.Player2Ready = 0;
                }
            }
        }
    }
}