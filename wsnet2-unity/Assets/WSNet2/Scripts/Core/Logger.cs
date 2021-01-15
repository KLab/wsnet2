using System;
using System.Diagnostics;

namespace WSNet2.Core
{
    /// <summary>
    /// WSNet2 が使用するLogger
    /// </summary>
    public class WSNet2Logger
    {
        // Same of Microsoft.Extensions.Logging.LogLevel
        public enum LogLevel
        {
            Trace = 0,
            Debug = 1,
            Information = 2,
            Warning = 3,
            Error = 4,
            Critical = 5,
            None = 6
        }

        /// <summary>
        /// アプリケーション側が提供するログ出力のインターフェース
        /// </summary>
        public interface ILogger
        {
            void Log(LogLevel logLevel, string message);
            void Log<TPayload>(LogLevel logLevel, TPayload payload, string message);
        }

        /// <summary>
        /// ILoggerを実装したデフォルトで使用されるLogger
        /// </summary>
        public class DefaultConsoleLogger : WSNet2Logger.ILogger
        {
            public void Log(WSNet2Logger.LogLevel logLevel, string message)
            {
                Console.WriteLine($"{logLevel,-8} {message}");
            }
            public void Log<TPayload>(WSNet2Logger.LogLevel logLevel, TPayload payload, string message)
            {
                Console.WriteLine($"{logLevel,-8} {message} {payload}");
            }
        }

        /// <summary>
        /// WSNet2 が使用する Logger を設定します
        /// </summary>
        public static ILogger Logger { get; set; } = new DefaultConsoleLogger();

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG")]
        public static void Debug(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Debug, string.Format(format, args));
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        [Conditional("DEBUG")]
        public static void DebugWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Debug, payload, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Info(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Information, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void InfoWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Information, payload, string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Warning(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Warning, string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void WarningWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Warning, payload, string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Error(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Error, string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void ErrorWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Error, payload, string.Format(format, args));
        }
    }
}