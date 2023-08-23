using System;
using System.Diagnostics;

namespace WSNet2
{
    /// <summary>
    /// WSNet2内部で使用するLogger
    /// </summary>
    public class Logger
    {
        IWSNet2Logger<WSNet2LogPayload> logger;

        /// <summary>
        /// コンストラクタ
        /// </summary>
        public Logger(IWSNet2Logger<WSNet2LogPayload> logger)
        {
            this.logger = logger;
        }

        /// <summary>
        /// Room情報をPayloadに設定します
        /// </summary>
        /// <param name="roomId">Room ID</param>
        /// <param name="roomNum">Room Number</param>
        public void SetRoomInfo(string roomId, int roomNum)
        {
            logger.Payload.RoomId = roomId;
            logger.Payload.RoomNum = roomNum;
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG"), Conditional("WSNET2_LOG_DEBUG")]
        public void Debug(string format, params object[] args)
        {
            logger?.Log(WSNet2LogLevel.Debug, null, format, args);
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG"), Conditional("WSNET2_LOG_DEBUG")]
        public void Debug(Exception e, string format, params object[] args)
        {
            logger?.Log(WSNet2LogLevel.Debug, e, format, args);
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Info(string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, args);
        public void Info(string message) =>
            logger?.Log(WSNet2LogLevel.Information, null, message);
        public void Info<T1>(string format, T1 a1) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, a1);
        public void Info<T1, T2>(string format, T1 a1, T2 a2) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, a1, a2);
        public void Info<T1, T2, T3>(string format, T1 a1, T2 a2, T3 a3) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, a1, a2, a3);
        public void Info<T1, T2, T3, T4>(string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, a1, a2, a3, a4);
        public void Info<T1, T2, T3, T4, T5>(string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Info(Exception exception, string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Information, exception, format, args);
        public void Info(Exception exception, string message) =>
            logger?.Log(WSNet2LogLevel.Information, exception, message);
        public void Info<T1>(Exception exception, string format, T1 a1) =>
            logger?.Log(WSNet2LogLevel.Information, exception, format, a1);
        public void Info<T1, T2>(Exception exception, string format, T1 a1, T2 a2) =>
            logger?.Log(WSNet2LogLevel.Information, exception, format, a1, a2);
        public void Info<T1, T2, T3>(Exception exception, string format, T1 a1, T2 a2, T3 a3) =>
            logger?.Log(WSNet2LogLevel.Information, exception, format, a1, a2, a3);
        public void Info<T1, T2, T3, T4>(Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger?.Log(WSNet2LogLevel.Information, exception, format, a1, a2, a3, a4);
        public void Info<T1, T2, T3, T4, T5>(Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger?.Log(WSNet2LogLevel.Information, exception, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Warning(string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, args);
        public void Warning(string message) =>
            logger?.Log(WSNet2LogLevel.Warning, null, message);
        public void Warning<T1>(string format, T1 a1) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, a1);
        public void Warning<T1, T2>(string format, T1 a1, T2 a2) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, a1, a2);
        public void Warning<T1, T2, T3>(string format, T1 a1, T2 a2, T3 a3) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, a1, a2, a3);
        public void Warning<T1, T2, T3, T4>(string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, a1, a2, a3, a4);
        public void Warning<T1, T2, T3, T4, T5>(string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Warning(Exception exception, string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, format, args);
        public void Warning(Exception exception, string message) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, message);
        public void Warning<T1>(Exception exception, string format, T1 a1) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, format, a1);
        public void Warning<T1, T2>(Exception exception, string format, T1 a1, T2 a2) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, format, a1, a2);
        public void Warning<T1, T2, T3>(Exception exception, string format, T1 a1, T2 a2, T3 a3) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, format, a1, a2, a3);
        public void Warning<T1, T2, T3, T4>(Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, format, a1, a2, a3, a4);
        public void Warning<T1, T2, T3, T4, T5>(Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger?.Log(WSNet2LogLevel.Warning, exception, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Error(string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, args);
        public void Error(string message) =>
            logger?.Log(WSNet2LogLevel.Error, null, message);
        public void Error<T1>(string format, T1 a1) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, a1);
        public void Error<T1, T2>(string format, T1 a1, T2 a2) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, a1, a2);
        public void Error<T1, T2, T3>(string format, T1 a1, T2 a2, T3 a3) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, a1, a2, a3);
        public void Error<T1, T2, T3, T4>(string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, a1, a2, a3, a4);
        public void Error<T1, T2, T3, T4, T5>(string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Errorレベルのログを出力します(Exception付き)
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Error(Exception exception, string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Error, exception, format, args);
        public void Error(Exception exception, string message) =>
            logger?.Log(WSNet2LogLevel.Error, exception, message);
        public void Error<T1>(Exception exception, string format, T1 a1) =>
            logger?.Log(WSNet2LogLevel.Error, exception, format, a1);
        public void Error<T1, T2>(Exception exception, string format, T1 a1, T2 a2) =>
            logger?.Log(WSNet2LogLevel.Error, exception, format, a1, a2);
        public void Error<T1, T2, T3>(Exception exception, string format, T1 a1, T2 a2, T3 a3) =>
            logger?.Log(WSNet2LogLevel.Error, exception, format, a1, a2, a3);
        public void Error<T1, T2, T3, T4>(Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger?.Log(WSNet2LogLevel.Error, exception, format, a1, a2, a3, a4);
        public void Error<T1, T2, T3, T4, T5>(Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger?.Log(WSNet2LogLevel.Error, exception, format, a1, a2, a3, a4, a5);
    }
}
