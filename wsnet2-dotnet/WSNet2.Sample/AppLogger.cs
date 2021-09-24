using WSNet2.Core;
using Microsoft.Extensions.Logging;
using ZLogger;
using System;
using System.Diagnostics;

namespace WSNet2.Sample
{
    /// <summary>
    /// 構造化ログのためのPayload
    /// </summary>
    class Payload : WSNet2LogPayload
    {
        public string ClientType;
        public string Server;
    }

    /// <summary>
    /// ZLoggerを使ったLoggerの例
    /// </summary>
    class AppLogger : IWSNet2Logger<Payload>
    {
        public Payload Payload { get; } = new Payload();

        ILogger logger;

        /// <summary>
        /// コンストラクタ
        /// </summary>
        /// <remarks>
        /// ZLoggerを設定してください
        /// </remarks>
        public AppLogger(ILogger logger)
        {
            this.logger = logger;
        }

        /// <summary>
        /// IWSNet2Logger<T>.Logの実装
        /// </summary>
        /// <remarks>
        /// boxing回避(パフォーマンス対策)のためのジェネリックメソッドも定義しています
        /// </remarks>
        public void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args)
        {
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, string.Format(format, args));
        }
        public void Log(WSNet2LogLevel logLevel, Exception exception, string message) =>
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, message);
        public void Log<T1>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1) =>
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1);
        public void Log<T1, T2>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2) =>
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2);
        public void Log<T1, T2, T3>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3) =>
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2, a3);
        public void Log<T1, T2, T3, T4>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2, a3, a4);
        public void Log<T1, T2, T3, T4, T5>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <remarks>
        /// Conditional属性により、DEBUGシンボルが定義されていない場合呼び出しは無視されます
        /// </remarks>
        [Conditional("DEBUG")]
        public void Debug(string format, params object[] args) =>
            Log(WSNet2LogLevel.Debug, null, format, args);

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        public void Info(string format, params object[] args) =>
            Log(WSNet2LogLevel.Information, null, format, args);

        public void Info(string format) =>
            Log(WSNet2LogLevel.Information, null, format);

        public void Info<T1>(string format, T1 a1) =>
            Log(WSNet2LogLevel.Information, null, format, a1);

        public void Info<T1, T2>(string format, T1 a1, T2 a2) =>
            Log(WSNet2LogLevel.Information, null, format, a1, a2);

        public void Info<T1, T2, T3>(string format, T1 a1, T2 a2, T3 a3) =>
            Log(WSNet2LogLevel.Information, null, format, a1, a2, a3);

        public void Info<T1, T2, T3, T4>(string format, T1 a1, T2 a2, T3 a3, T4 a4) =>
            Log(WSNet2LogLevel.Information, null, format, a1, a2, a3, a4);

        public void Info<T1, T2, T3, T4, T5>(string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5) =>
            Log(WSNet2LogLevel.Information, null, format, a1, a2, a3, a4, a5);

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        public void Warning(string format, params object[] args) =>
            Log(WSNet2LogLevel.Warning, null, format, args);

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        public void Error(Exception exception, string format, params object[] args) =>
            Log(WSNet2LogLevel.Error, exception, format, args);
    }
}
