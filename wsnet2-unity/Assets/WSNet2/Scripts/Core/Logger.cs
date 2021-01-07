using System;

namespace WSNet2.Core
{
    /// <summary>
    /// WSNet2 が使用するLogger
    /// </summary>
    public class WSNet2Logger
    {
        public enum LogLevel
        {
            Quiet = 0,
            Error,
            Warning,
            Info,
            Debug,
        }

        /// <summary>
        /// アプリケーション側が提供するログ出力のインターフェース
        /// </summary>
        public interface ILogger
        {
            void Log(LogLevel logLevel, string message);
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
        public static void Debug(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Debug, string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Info(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Info, string.Format(format, args));
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
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Error(string format, params object[] args)
        {
            Logger?.Log(LogLevel.Error, string.Format(format, args));
        }
    }
}