using System;
using WSNet2.Core;

namespace WSNet2
{
    /// <summary>
    /// Unity環境でデフォルトで使用されるLogger
    /// </summary>
    public class DefaultUnityLogger : WSNet2Logger.ILogger
    {
        public void Log<TPayload>(WSNet2Logger.LogLevel logLevel, Exception e, TPayload payload, string message)
        {
            var msg = message;
            if (payload != null)
            {
                msg = $"{message} Payload = {payload}";
            }

            switch (logLevel)
            {
                case WSNet2Logger.LogLevel.Error:
                    UnityEngine.Debug.LogError(msg);
                    break;
                case WSNet2Logger.LogLevel.Warning:
                    UnityEngine.Debug.LogWarning(msg);
                    break;
                case WSNet2Logger.LogLevel.Information:
                case WSNet2Logger.LogLevel.Debug:
                    UnityEngine.Debug.Log(msg);
                    break;
            }

            if (e != null)
            {
                UnityEngine.Debug.LogException(e);
            }
        }
    }
}