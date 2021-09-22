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
        public void Log(WSNet2LogLevel logLevel, Exception e, string format, params object[] param)
        {
            logger.ZLogWithPayload((LogLevel)logLevel, e, Payload, string.Format(format, param));
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <remarks>
        /// Conditional属性により、DEBUGシンボルが定義されていない場合呼び出しは無視されます
        /// </remarks>
        [Conditional("DEBUG")]
        public void Debug(string format, params object[] param) =>
            Log(WSNet2LogLevel.Debug, null, format, param);

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <remarks>
        /// boxing回避(パフォーマンス対策)のためのジェネリックメソッドも定義しています
        /// </remarks>
        public void Info(string format, params object[] param) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, string.Format(format, param));

        public void Info(string format) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, format);

        public void Info<T1>(string format, T1 p1) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, format, p1);

        public void Info<T1, T2>(string format, T1 p1, T2 p2) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, format, p1, p2);

        public void Info<T1, T2, T3>(string format, T1 p1, T2 p2, T3 p3) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, format, p1, p2, p3);

        public void Info<T1, T2, T3, T4>(string format, T1 p1, T2 p2, T3 p3, T4 p4) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, format, p1, p2, p3, p4);

        public void Info<T1, T2, T3, T4, T5>(string format, T1 p1, T2 p2, T3 p3, T4 p4, T5 p5) =>
            logger.ZLogWithPayload(LogLevel.Information, null, Payload, format, p1, p2, p3, p4, p5);

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        public void Warning(string format, params object[] param) =>
            Log(WSNet2LogLevel.Warning, null, format, param);

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        public void Error(Exception e, string format, params object[] param) =>
            Log(WSNet2LogLevel.Error, e, format, param);
    }
}
