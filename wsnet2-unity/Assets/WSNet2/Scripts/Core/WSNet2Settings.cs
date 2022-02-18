namespace WSNet2.Core
{
    public class WSNet2Settings
    {
        /// <summary>保持できるEventの数</summary>
        public static int EvPoolSize = 16;

        /// <summary>各Eventのバッファサイズの初期値</summary>
        public static int EvBufInitialSize = 256;

        /// <summary>保持できるMsgの数</summary>
        public static int MsgPoolSize = 16;

        /// <summary>各Msgのバッファサイズの初期値</summary>
        public static int MsgBufInitialSize = 256;

        /// <summary>最大連続再接続試行回数</summary>
        public static int MaxReconnection = 5;

        /// <summary>接続タイムアウト</summary>
        public static int ConnectTimeoutMilliSec = 5000;

        /// <summary>再接続インターバル (milli seconds)</summary>
        public static int RetryIntervalMilliSec = 1000;

        /// <summary>最大Ping間隔 (milli seconds)</summary>
        /// Playerの最終Msg時刻のやりとりのため、ある程度で上限を設ける
        public static int MaxPingIntervalMilliSec = 10000;

        /// <summary>最小Ping間隔 (milli seconds)</summary>
        public static int MinPingIntervalMilliSec = 1000;
    }
}
