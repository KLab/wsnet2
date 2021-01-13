using System;
using WSNet2.Core;

namespace WSNet2
{
    /// <summary>
    /// Unity環境でデフォルトで使用されるLogger
    /// </summary>
    public class DefaultUnityLogger : WSNet2Logger.ILogger
    {
        public void Log(WSNet2Logger.LogLevel logLevel, string message)
        {
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
        }

        public void Log<TPayload>(WSNet2Logger.LogLevel logLevel, TPayload payload, string message)
        {
            Log(logLevel, $"{message} payload = {payload}");
        }
    }
}