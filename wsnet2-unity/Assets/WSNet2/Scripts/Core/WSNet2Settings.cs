namespace WSNet2.Core
{
    public class WSNet2Settings
    {
        /// <summary>保持できるEventの数</summary>
        public int EvPoolSize = 16;

        /// <summary>各Eventのバッファサイズの初期値</summary>
        public int EvBufInitialSize = 256;

        /// <summary>保持できるMsgの数</summary>
        public int MsgPoolSize = 16;

        /// <summary>各Msgのバッファサイズの初期値</summary>
        public int MsgBufInitialSize = 256;

        /// <summary>最大連続再接続試行回数</summary>
        public int MaxReconnection = 5;

        /// <summary>接続タイムアウト</summary>
        public int ConnectTimeoutMilliSec = 5000;

        /// <summary>再接続インターバル (milli seconds)</summary>
        public int RetryIntervalMilliSec = 1000;

        /// <summary>最大Ping間隔 (milli seconds)</summary>
        /// Playerの最終Msg時刻のやりとりのため、ある程度で上限を設ける
        public int MaxPingIntervalMilliSec = 10000;

        /// <summary>最小Ping間隔 (milli seconds)</summary>
        public int MinPingIntervalMilliSec = 1000;
    }
}
