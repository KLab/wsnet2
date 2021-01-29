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
            void Log<TPayload>(LogLevel logLevel, Exception e, TPayload payload, string message);
        }

        /// <summary>
        /// WSNet2 が使用する Logger を設定します
        /// </summary>
        public static ILogger Logger { get; set; } = new DefaultConsoleLogger();

        /// <summary>
        /// ILoggerを実装したデフォルトで使用されるLogger
        /// </summary>
        public class DefaultConsoleLogger : WSNet2Logger.ILogger
        {
            public void Log<TPayload>(WSNet2Logger.LogLevel logLevel, Exception e, TPayload payload, string message)
            {
                Console.WriteLine($"{logLevel,-12} {message} {payload}");

                if (e != null)
                {
                    Console.WriteLine(e.ToString());
                }
            }
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG")]
        public static void Debug(string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Debug, null, null, string.Format(format, args));
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG")]
        public static void Debug(Exception e, string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Debug, e, null, string.Format(format, args));
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="payload"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG")]
        public static void DebugWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Debug, null, payload, string.Format(format, args));
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="payload"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        [Conditional("DEBUG")]
        public static void DebugWithPayload<TPayload>(Exception e, TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Debug, e, payload, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Info(string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Information, null, null, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Info(Exception e, string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Information, e, null, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void InfoWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Information, null, payload, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void InfoWithPayload<TPayload>(Exception e, TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Information, e, payload, string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Warning(string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Warning, null, null, string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Warning(Exception e, string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Warning, e, null, string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void WarningWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Warning, null, payload, string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void WarningWithPayload<TPayload>(Exception e, TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Warning, e, payload, string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Error(string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Error, null, null, string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Error(Exception e, string format, params object[] args)
        {
            Logger?.Log<object>(LogLevel.Error, e, null, string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void ErrorWithPayload<TPayload>(TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Error, null, payload, string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="e"></param>
        /// <param name="format"></param>
        /// <param name="args"></param>
        /// <param name="payload"></param>
        public static void ErrorWithPayload<TPayload>(Exception e, TPayload payload, string format, params object[] args)
        {
            Logger?.Log(LogLevel.Error, e, payload, string.Format(format, args));
        }
    }
}