using System;
using System.Collections.Generic;
using System.Diagnostics;
using WSNet2.Core;

#if UNITY_5_3_OR_NEWER
using Vector2 = UnityEngine.Vector2;
using Vector3 = UnityEngine.Vector3;
#endif

namespace Sample.Logic
{
    /// <summary>
    /// GameTimer
    /// </summary>
    public class GameTimer
    {

        /// <summary>
        /// 最後に受信したサーバ時間
        /// </summary>
        private long serverTick;
        private Stopwatch stopwatch;

        public GameTimer()
        {
            serverTick = 0;
            stopwatch = new Stopwatch();
            stopwatch.Start();
        }

        /// <summary>
        /// サーバ時間を更新する
        /// </summary>
        public void UpdateServerTick(long newServerTick)
        {
            if (1.0 < new TimeSpan(Math.Abs(newServerTick - serverTick)).TotalSeconds || (newServerTick < serverTick))
            {
                serverTick = newServerTick;
                stopwatch.Restart();
            }
        }

        /// <summary>
        /// サーバ時間を基準に現在のTickを取得
        /// </summary>
        public long NowTick
        {
            get
            {
                return serverTick + stopwatch.Elapsed.Ticks;
            }
        }
    }


    /// <summary>
    /// Pongゲームの状態を表すデータ
    /// </summary>
    public class GameState : IWSNet2Serializable
    {
        /// <summary>
        /// ゲームの状態の種類
        /// </summary>
        public GameStateCode Code;

        /// <summary>
        /// 最後に更新した時間
        /// </summary>
        public long Tick;

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
        public List<Ball> Balls;

        public void Serialize(SerialWriter writer)
        {
            writer.Write((int)Code);
            writer.Write(Tick);
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
            writer.Write(Balls);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            Code = (GameStateCode)reader.ReadInt();
            Tick = reader.ReadLong();
            GameCount = reader.ReadInt();
            MasterId = reader.ReadString();
            Player1 = reader.ReadString();
            Player2 = reader.ReadString();
            Player1Ready = reader.ReadInt();
            Player2Ready = reader.ReadInt();
            Score1 = reader.ReadInt();
            Score2 = reader.ReadInt();
            Bar1 = reader.ReadObject(Bar1);
            Bar2 = reader.ReadObject(Bar2);
            Balls = reader.ReadList(Balls);
        }

        public GameState Copy()
        {
            var o = new GameState();
            o.Code = Code;
            o.Tick = Tick;
            o.GameCount = GameCount;
            o.MasterId = MasterId;
            o.Player1 = Player1;
            o.Player2 = Player2;
            o.Player1Ready = Player1Ready;
            o.Player2Ready = Player2Ready;
            o.Score1 = Score1;
            o.Score2 = Score2;
            o.Bar1 = Bar1.Copy();
            o.Bar2 = Bar2.Copy();
            o.Balls = new List<Ball>();
            foreach (var ball in Balls)
            {
                o.Balls.Add(ball.Copy());
            }
            return o;
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
        /// ゲーム終了 プレイヤーの退室を待っている
        /// </summary>
        End,
    }

    /// <summary>
    /// ボールのデータ
    /// </summary>
    public class Ball : IWSNet2Serializable
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

        public Ball Copy()
        {
            var o = new Ball();
            o.Position.x = Position.x;
            o.Position.y = Position.y;
            o.Direction.x = Direction.x;
            o.Direction.y = Direction.y;
            o.Speed = Speed;
            o.Radius = Radius;
            return o;
        }
    }

    /// <summary>
    /// バーのデータ
    /// </summary>
    public class Bar : IWSNet2Serializable
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

        public Bar Copy()
        {
            var o = new Bar();
            o.Position.x = Position.x;
            o.Position.y = Position.y;
            o.Direction.x = Direction.x;
            o.Direction.y = Direction.y;
            o.Speed = Speed;
            o.Width = Width;
            o.Height = Height;
            return o;
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
    public class PlayerEvent : IWSNet2Serializable
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

        /// <summary>
        /// タイムスタンプ
        /// </summary>
        public long Tick;

        public void Serialize(SerialWriter writer)
        {
            writer.Write((int)Code);
            writer.Write(PlayerId);
            writer.Write((int)MoveInput);
            writer.Write(Tick);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            Code = (PlayerEventCode)reader.ReadInt();
            PlayerId = string.Intern(reader.ReadString());
            MoveInput = (MoveInputCode)reader.ReadInt();
            Tick = reader.ReadLong();
        }
    }

    /// <summary>
    /// Pongゲームのシミュレータ
    /// </summary>
    public class GameSimulator
    {
        /// <summary>
        /// ゲームのステートを変更する権利があるか
        /// </summary>
        /// <value></value>
        public bool IsMaster { get; private set; }

        public int BoardWidth { get; private set; }
        public int BoardHeight { get; private set; }
        public Vector2 Center { get; private set; }
        public int MinX { get; private set; }
        public int MaxX { get; private set; }
        public int MinY { get; private set; }
        public int MaxY { get; private set; }

        public GameSimulator(bool isMaster)
        {
            this.IsMaster = isMaster;
        }

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
            state.Balls = new List<Ball>();
            for (int i = 0; i < 22; ++i)
            {
                state.Balls.Add(new Ball());
            }
        }

        void ResetBallPosition(Ball ball, Random rnd)
        {
            var x = (double)rnd.Next(-1000, 1000);
            var y = (double)rnd.Next(-100, 100);
            var s = Math.Sqrt(x * x + y * y);
            ball.Position.x = 0;
            ball.Position.y = 0;
            ball.Direction.x = (float)(x / s);
            ball.Direction.y = (float)(y / s);
            ball.Speed = 600f + (float)(100 * rnd.NextDouble());
            ball.Radius = 10;
        }

        void ResetPositions(GameState state)
        {
            state.Bar1.Width = 10;
            state.Bar1.Height = 200;
            state.Bar1.Position.x = MinX + 50;
            state.Bar1.Position.y = 0;
            state.Bar1.Speed = 400f;

            state.Bar2.Width = 10;
            state.Bar2.Height = 200;
            state.Bar2.Position.x = MaxX - 50;
            state.Bar2.Position.y = 0;
            state.Bar2.Speed = 400f;

            var rnd = new Random();
            foreach (var ball in state.Balls)
            {
                ResetBallPosition(ball, rnd);
            }
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

        bool UpdateGameInternal(long tick, GameState state, PlayerEvent ev)
        {
            long prevTick = state.Tick;
            state.Tick = tick;
            float dt = (float)new TimeSpan(tick - prevTick).TotalSeconds;
            bool forceSync = false;

            if (state.Code == GameStateCode.WaitingGameMaster)
            {
                // マスタクライアントの参加を待っている
                // 何もしない
                return false;
            }
            if (state.Code == GameStateCode.WaitingPlayer)
            {
                // プレイヤーの参加を待っている
                // 先に参加したプレイヤーを1P, あとに参加したプレイヤーを2Pとする
                if (ev?.Code == PlayerEventCode.Join)
                {
                    if (ev.PlayerId == state.Player1 || ev.PlayerId == state.Player2)
                    {
                        return false;
                    }

                    if (string.IsNullOrEmpty(state.Player1))
                    {
                        state.Player1 = ev.PlayerId;
                    }
                    else if (string.IsNullOrEmpty(state.Player2))
                    {
                        state.Player2 = ev.PlayerId;

                        if (IsMaster)
                        {
                            // プレイヤーが集まったのでスタート準備へ
                            state.Code = GameStateCode.ReadyToStart;
                            return true;
                        }
                    }
                }
            }
            else if (state.Code == GameStateCode.ReadyToStart)
            {

                if (ev?.Code == PlayerEventCode.Ready)
                {
                    // 1P, 2Pが Ready 入力を送ってくるのを待っている
                    if (ev.PlayerId == state.Player1)
                    {
                        state.Player1Ready = 1;
                    }
                    if (ev.PlayerId == state.Player2)
                    {
                        state.Player2Ready = 1;
                    }
                }

                if (state.Player1Ready == 1 && state.Player2Ready == 1)
                {
                    if (IsMaster)
                    {
                        // 準備ができたのでゲーム開始
                        StartNextGame(state);
                        return true;
                    }
                }
            }
            else if (state.Code == GameStateCode.InGame)
            {
                // 移動入力を処理する
                if (ev?.Code == PlayerEventCode.Move)
                {
                    float dirY = 0;
                    switch (ev.MoveInput)
                    {
                        case MoveInputCode.Stop: dirY = 0; break;
                        case MoveInputCode.Up: dirY = 1; break;
                        case MoveInputCode.Down: dirY = -1; break;
                    }

                    if (ev.PlayerId == state.Player1 && state.Bar1.Direction.y != dirY)
                    {
                        state.Bar1.Direction.y = dirY;
                        forceSync = true;
                    }
                    if (ev.PlayerId == state.Player2 && state.Bar2.Direction.y != dirY)
                    {
                        state.Bar2.Direction.y = dirY;
                        forceSync = true;
                    }
                }
            }
            else if (state.Code == GameStateCode.End)
            {
                return false;
            }

            if (state.Code != GameStateCode.InGame)
            {
                return false;
            }

            // 1Pのバーの移動
            state.Bar1.Position.x = Math.Min(MaxX - state.Bar1.Width / 2, Math.Max(MinX + state.Bar1.Width / 2, state.Bar1.Position.x + state.Bar1.Direction.x * state.Bar1.Speed * dt));
            state.Bar1.Position.y = Math.Min(MaxY - state.Bar1.Height / 2, Math.Max(MinY + state.Bar1.Height / 2, state.Bar1.Position.y + state.Bar1.Direction.y * state.Bar1.Speed * dt));

            // 2Pのバーの移動
            state.Bar2.Position.x = Math.Min(MaxX - state.Bar2.Width / 2, Math.Max(MinX + state.Bar2.Width / 2, state.Bar2.Position.x + state.Bar2.Direction.x * state.Bar2.Speed * dt));
            state.Bar2.Position.y = Math.Min(MaxY - state.Bar2.Height / 2, Math.Max(MinY + state.Bar2.Height / 2, state.Bar2.Position.y + state.Bar2.Direction.y * state.Bar2.Speed * dt));

            // ボールの移動
            foreach (var ball in state.Balls)
            {
                ball.Position.x = ball.Position.x + ball.Direction.x * ball.Speed * dt;
                ball.Position.y = ball.Position.y + ball.Direction.y * ball.Speed * dt;
            }

            // ボールが上下の壁にあたってたら移動方向を反射
            foreach (var ball in state.Balls)
            {
                if (ball.Position.y < MinY + ball.Radius || ball.Position.y > MaxY - ball.Radius)
                {
                    ball.Direction.y *= -1;
                }
            }

            // ボールが壁にめり込んでいたら戻す
            foreach (var ball in state.Balls)
            {
                ball.Position.y = Math.Min(MaxY - ball.Radius, Math.Max(MinY + ball.Radius, ball.Position.y));
            }

            // 1Pのバーにボールが当たってたら反射
            foreach (var ball in state.Balls)
            {
                var bx = ball.Position.x;
                var by = ball.Position.y;
                var br = ball.Radius;
                var px = state.Bar1.Position.x;
                var py = state.Bar1.Position.y;
                var pw = state.Bar1.Width;
                var ph = state.Bar1.Height;

                if (bx - br <= px + pw / 2f && bx + br >= px + pw / 2f)
                {
                    if (py - ph / 2f <= by && py + ph / 2f >= by)
                    {
                        if (ball.Direction.x < 0)
                        {
                            ball.Direction.x *= -1;
                        }
                    }
                }
            }

            // 2Pのバーにボールが当たってたら反射
            foreach (var ball in state.Balls)
            {
                var bx = ball.Position.x;
                var by = ball.Position.y;
                var br = ball.Radius;
                var px = state.Bar2.Position.x;
                var py = state.Bar2.Position.y;
                var pw = state.Bar2.Width;
                var ph = state.Bar2.Height;

                if (bx - br <= px + pw / 2f && bx + br >= px + pw / 2f)
                {
                    if (py - ph / 2f <= by && py + ph / 2f >= by)
                    {
                        if (ball.Direction.x > 0)
                        {
                            ball.Direction.x *= -1;
                        }
                    }
                }
            }

            // 1Pのゴールに入っていたら2Pに得点
            foreach (var ball in state.Balls)
            {
                if (ball.Position.x < MinX + ball.Radius)
                {
                    if (IsMaster)
                    {
                        state.Score2 += 1;
                        ResetBallPosition(ball, new Random());
                        forceSync = true;
                    }
                }
            }

            // 2Pのゴールに入っていたら1Pに得点
            foreach (var ball in state.Balls)
            {
                if (MaxX + ball.Radius < ball.Position.x)
                {
                    if (IsMaster)
                    {
                        state.Score1 += 1;
                        ResetBallPosition(ball, new Random());
                        forceSync = true;
                    }
                }
            }

            // 100点とったら終わり
            if (100 <= state.Score1 || 100 <= state.Score2)
            {
                if (IsMaster)
                {
                    state.Code = GameStateCode.End;
                }
            }

            return forceSync;
        }

        /// <summary>
        /// </summary>

        /// <summary>
        /// ゲームをシミュレーションする
        /// events は Tick で昇順ソートされており, eventsの中で最小の Tick は state の最終更新よりも後である必要がある
        /// </summary>
        /// <param name="nowTick">最新のサーバ時刻</param>
        /// <param name="state">ゲームステート</param>
        /// <param name="events">プレイヤーの入力イベント</param>
        /// <returns>プレイヤーの入力や、ステートの変更が反映され、同期をすぐに行ったほうが良い場合 true</returns>
        public bool UpdateGame(long nowTick, GameState state, IEnumerable<PlayerEvent> events)
        {
            bool forceSync = false;
            foreach (var ev in events)
            {
                var prevState = state.Code;
                forceSync |= UpdateGameInternal(ev.Tick, state, ev);
                if (prevState != state.Code)
                {
                    return forceSync;
                }
            }
            forceSync |= UpdateGameInternal(nowTick, state, null);
            return forceSync;
        }
    }
}
