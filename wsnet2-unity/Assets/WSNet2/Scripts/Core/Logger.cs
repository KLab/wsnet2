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
        public class DefaultLogger : ILogger
        {
            public void Log(LogLevel logLevel, string message)
            {
#if UNITY_5_3_OR_NEWER
                switch (logLevel)
                {
                    case LogLevel.Error:
                        UnityEngine.Debug.LogError(message);
                        break;
                    case LogLevel.Warning:
                        UnityEngine.Debug.LogWarning(message);
                        break;
                    case LogLevel.Info:
                    case LogLevel.Debug:
                        UnityEngine.Debug.Log(message);
                        break;
                }
#else
                Console.WriteLine($"{logLevel,-8} {message}");
#endif
            }
        }

        private static ILogger logger = new DefaultLogger();
        private static LogLevel level = LogLevel.Info;

        private static string logPrefix = "[WSNet2]";

        /// <summary>
        /// WSNet2 が使用する Logger を設定します
        /// </summary>
        public static void SetLogger(ILogger logger)
        {
            WSNet2Logger.logger = logger;
        }

        /// <summary>
        /// WSNet2 が出力するログレベルを指定します
        /// </summary>
        public static void SetLogLevel(LogLevel level)
        {
            WSNet2Logger.level = level;
        }

        /// <summary>
        /// WSNet2 が出力するログのプレフィックスを指定します
        /// </summary>
        public static void SetPrefix(string prefix)
        {
            WSNet2Logger.logPrefix = prefix;
        }

        /// <summary>
        /// Debugレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Debug(string format, params object[] args)
        {
            if (logger == null || level < LogLevel.Debug)
            {
                return;
            }
            logger.Log(LogLevel.Debug, logPrefix + string.Format(format, args));
        }

        /// <summary>
        /// Infoレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Info(string format, params object[] args)
        {
            if (logger == null || level < LogLevel.Info)
            {
                return;
            }
            logger.Log(LogLevel.Info, logPrefix + string.Format(format, args));
        }

        /// <summary>
        /// Warningレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Warning(string format, params object[] args)
        {
            if (logger == null || level < LogLevel.Warning)
            {
                return;
            }
            logger.Log(LogLevel.Warning, logPrefix + string.Format(format, args));
        }

        /// <summary>
        /// Errorレベルのログを出力します
        /// </summary>
        /// <param name="format"></param>
        /// <param name="args"></param>
        public static void Error(string format, params object[] args)
        {
            if (logger == null || level < LogLevel.Error)
            {
                return;
            }
            logger.Log(LogLevel.Error, logPrefix + string.Format(format, args));
        }
    }
}