using System;
using WSNet2.Core;

namespace WSNet2
{
    /// <summary>
    /// ILoggerを実装したデフォルトで使用されるLogger
    /// </summary>
    public class DefaultLogger : WSNet2Logger.ILogger
    {
        public void Log(WSNet2Logger.LogLevel logLevel, string message)
        {
#if UNITY_5_3_OR_NEWER
                switch (logLevel)
                {
                    case WSNet2Logger.LogLevel.Error:
                        UnityEngine.Debug.LogError(message);
                        break;
                    case WSNet2Logger.LogLevel.Warning:
                        UnityEngine.Debug.LogWarning(message);
                        break;
                    case WSNet2Logger.LogLevel.Info:
                    case WSNet2Logger.LogLevel.Debug:
                        UnityEngine.Debug.Log(message);
                        break;
                }
#else
            Console.WriteLine($"{logLevel,-8} {message}");
#endif
        }
    }
}