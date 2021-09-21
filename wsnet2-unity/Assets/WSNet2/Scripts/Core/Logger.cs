using System;
using System.Diagnostics;

namespace WSNet2.Core
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
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG")]
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
        [Conditional("DEBUG")]
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
        public void Info<T1>(string format, T1 p1) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, p1);
        public void Info<T1, T2>(string format, T1 p1, T2 p2) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, p1, p2);
        public void Info<T1, T2, T3>(string format, T1 p1, T2 p2, T3 p3) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, p1, p2, p3);
        public void Info<T1, T2, T3, T4>(string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, p1, p2, p3, p4);
        public void Info<T1, T2, T3, T4, T5>(string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger?.Log(WSNet2LogLevel.Information, null, format, p1, p2, p3, p4, p5);

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Info(Exception e, string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Information, e, format, args);
        public void Info(Exception e, string message) =>
            logger?.Log(WSNet2LogLevel.Information, e, message);
        public void Info<T1>(Exception e, string format, T1 p1) =>
            logger?.Log(WSNet2LogLevel.Information, e, format, p1);
        public void Info<T1, T2>(Exception e, string format, T1 p1, T2 p2) =>
            logger?.Log(WSNet2LogLevel.Information, e, format, p1, p2);
        public void Info<T1, T2, T3>(Exception e, string format, T1 p1, T2 p2, T3 p3) =>
            logger?.Log(WSNet2LogLevel.Information, e, format, p1, p2, p3);
        public void Info<T1, T2, T3, T4>(Exception e, string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger?.Log(WSNet2LogLevel.Information, e, format, p1, p2, p3, p4);
        public void Info<T1, T2, T3, T4, T5>(Exception e, string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger?.Log(WSNet2LogLevel.Information, e, format, p1, p2, p3, p4, p5);

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Warning(string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, args);
        public void Warning(string message) =>
            logger?.Log(WSNet2LogLevel.Warning, null, message);
        public void Warning<T1>(string format, T1 p1) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, p1);
        public void Warning<T1, T2>(string format, T1 p1, T2 p2) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, p1, p2);
        public void Warning<T1, T2, T3>(string format, T1 p1, T2 p2, T3 p3) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, p1, p2, p3);
        public void Warning<T1, T2, T3, T4>(string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, p1, p2, p3, p4);
        public void Warning<T1, T2, T3, T4, T5>(string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger?.Log(WSNet2LogLevel.Warning, null, format, p1, p2, p3, p4, p5);

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Warning(Exception e, string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Warning, e, format, args);
        public void Warning(Exception e, string message) =>
            logger?.Log(WSNet2LogLevel.Warning, e, message);
        public void Warning<T1>(Exception e, string format, T1 p1) =>
            logger?.Log(WSNet2LogLevel.Warning, e, format, p1);
        public void Warning<T1, T2>(Exception e, string format, T1 p1, T2 p2) =>
            logger?.Log(WSNet2LogLevel.Warning, e, format, p1, p2);
        public void Warning<T1, T2, T3>(Exception e, string format, T1 p1, T2 p2, T3 p3) =>
            logger?.Log(WSNet2LogLevel.Warning, e, format, p1, p2, p3);
        public void Warning<T1, T2, T3, T4>(Exception e, string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger?.Log(WSNet2LogLevel.Warning, e, format, p1, p2, p3, p4);
        public void Warning<T1, T2, T3, T4, T5>(Exception e, string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger?.Log(WSNet2LogLevel.Warning, e, format, p1, p2, p3, p4, p5);

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Error(string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, args);
        public void Error(string message) =>
            logger?.Log(WSNet2LogLevel.Error, null, message);
        public void Error<T1>(string format, T1 p1) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, p1);
        public void Error<T1, T2>(string format, T1 p1, T2 p2) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, p1, p2);
        public void Error<T1, T2, T3>(string format, T1 p1, T2 p2, T3 p3) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, p1, p2, p3);
        public void Error<T1, T2, T3, T4>(string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, p1, p2, p3, p4);
        public void Error<T1, T2, T3, T4, T5>(string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger?.Log(WSNet2LogLevel.Error, null, format, p1, p2, p3, p4, p5);

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public void Error(Exception e, string format, params object[] args) =>
            logger?.Log(WSNet2LogLevel.Error, e, format, args);
        public void Error(Exception e, string message) =>
            logger?.Log(WSNet2LogLevel.Error, e, message);
        public void Error<T1>(Exception e, string format, T1 p1) =>
            logger?.Log(WSNet2LogLevel.Error, e, format, p1);
        public void Error<T1, T2>(Exception e, string format, T1 p1, T2 p2) =>
            logger?.Log(WSNet2LogLevel.Error, e, format, p1, p2);
        public void Error<T1, T2, T3>(Exception e, string format, T1 p1, T2 p2, T3 p3) =>
            logger?.Log(WSNet2LogLevel.Error, e, format, p1, p2, p3);
        public void Error<T1, T2, T3, T4>(Exception e, string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger?.Log(WSNet2LogLevel.Error, e, format, p1, p2, p3, p4);
        public void Error<T1, T2, T3, T4, T5>(Exception e, string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger?.Log(WSNet2LogLevel.Error, e, format, p1, p2, p3, p4, p5);
    }
}
