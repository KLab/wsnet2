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
        /// <param name="message"></param>
        public static void Debug(string message)
        {
            Logger?.Log(LogLevel.Debug, message);
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="message"></param>
        public static void Info(string message)
        {
            Logger?.Log(LogLevel.Info, message);
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="message"></param>
        public static void Warning(string message)
        {
            Logger?.Log(LogLevel.Warning, message);
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="message"></param>
        public static void Error(string message)
        {
            Logger?.Log(LogLevel.Error, message);
        }
    }
}