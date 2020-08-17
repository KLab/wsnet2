using System;
using System.Collections.Generic;
using WSNet2.Core;

#if UNITY_5_3_OR_NEWER
using Vector2 = UnityEngine.Vector2;
using Vector3 = UnityEngine.Vector3;
#endif

namespace Sample.Logic
{
    public enum GameStateCode
    {
        None,
        WaitingGameMaster,
        WaitingPlayer,
        ReadyToStart,
        InGame,
        Goal,
    }

    public class Ball : IWSNetSerializable
    {
        public Vector2 Position;
        public Vector2 Direction;
        public float Speed;
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

    public class Bar : IWSNetSerializable
    {
        public Vector2 Position;
        public Vector2 Direction;
        public float Speed;
        public float Width;
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

    public class GameState : IWSNetSerializable
    {
        public GameStateCode Code;
        public int GameCount;
        public string MasterId;
        public string Player1;
        public string Player2;
        public int Player1Ready;
        public int Player2Ready;
        public int Score1;
        public int Score2;
        public Bar Bar1;
        public Bar Bar2;
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

    public enum PlayerEventCode
    {
        None,
        Join,
        Ready,
        Start,
        Move,
    }

    public enum MoveInputCode
    {
        Stop,
        Up,
        Down,
    }

    public class PlayerEvent : IWSNetSerializable
    {
        public PlayerEventCode Code;
        public string PlayerId;
        public MoveInputCode MoveInput; // 0:stop, 1:up, 2:down


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

    public class GameSimulator
    {
        public int BoardWidth { get; private set; }
        public int BoardHeight { get; private set; }
        public Vector2 Center { get; private set; }
        public int MinX { get; private set; }
        public int MaxX { get; private set; }
        public int MinY { get; private set; }
        public int MaxY { get; private set; }

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

        public void UpdateGame(GameState state, IEnumerable<PlayerEvent> events)
        {
            if (state.Code == GameStateCode.WaitingGameMaster)
            {
                // Do nothing 
            }
            if (state.Code == GameStateCode.WaitingPlayer)
            {
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
                                state.Code = GameStateCode.ReadyToStart;
                            }
                        }
                    }
                }
            }
            else if (state.Code == GameStateCode.ReadyToStart || state.Code == GameStateCode.Goal)
            {
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

                state.Bar1.Position.x = Math.Min(MaxX - state.Bar1.Width / 2, Math.Max(MinX + state.Bar1.Width / 2, state.Bar1.Position.x + state.Bar1.Direction.x * state.Bar1.Speed));
                state.Bar1.Position.y = Math.Min(MaxY - state.Bar1.Height / 2, Math.Max(MinY + state.Bar1.Height / 2, state.Bar1.Position.y + state.Bar1.Direction.y * state.Bar1.Speed));

                state.Bar2.Position.x = Math.Min(MaxX - state.Bar2.Width / 2, Math.Max(MinX + state.Bar2.Width / 2, state.Bar2.Position.x + state.Bar2.Direction.x * state.Bar2.Speed));
                state.Bar2.Position.y = Math.Min(MaxY - state.Bar2.Height / 2, Math.Max(MinY + state.Bar2.Height / 2, state.Bar2.Position.y + state.Bar2.Direction.y * state.Bar2.Speed));

                state.Ball.Position.x = state.Ball.Position.x + state.Ball.Direction.x * state.Ball.Speed;
                state.Ball.Position.y = state.Ball.Position.y + state.Ball.Direction.y * state.Ball.Speed;

                if (state.Ball.Position.y < MinY + state.Ball.Radius || state.Ball.Position.y > MaxY - state.Ball.Radius)
                {
                    state.Ball.Direction.y *= -1;
                }

                state.Ball.Position.y = Math.Min(MaxY - state.Ball.Radius, Math.Max(MinY + state.Ball.Radius, state.Ball.Position.y));

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

                if (state.Ball.Position.x < MinX + state.Ball.Radius)
                {
                    state.Code = GameStateCode.Goal;
                    state.Score2 += 1;
                    state.Player1Ready = 0;
                    state.Player2Ready = 0;
                }
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