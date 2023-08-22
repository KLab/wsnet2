using System;

namespace WSNet2
{
    /// <summary>
    /// Unity環境でデフォルトで使用されるLogger
    /// </summary>
    public class DefaultUnityLogger : IWSNet2Logger<WSNet2LogPayload>
    {
        public WSNet2LogPayload Payload { get; } = new WSNet2LogPayload();

        public void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args)
        {
            var msg = $"{string.Format(format, args)}\nPayload: User={Payload.UserId}, Room={Payload.RoomId}, RoomNum={Payload.RoomNum}";

            switch (logLevel)
            {
                case WSNet2LogLevel.Critical:
                case WSNet2LogLevel.Error:
                    UnityEngine.Debug.LogError(msg);
                    if (exception != null)
                    {
                        UnityEngine.Debug.LogException(exception);
                    }
                    break;
                case WSNet2LogLevel.Warning:
                    if (exception != null)
                    {
                        msg = $"{msg}: {exception.Message}";
                    }
                    UnityEngine.Debug.LogWarning(msg);
                    break;
                case WSNet2LogLevel.Information:
                case WSNet2LogLevel.Debug:
                case WSNet2LogLevel.Trace:
                    if (exception != null)
                    {
                        msg = $"{msg}: {exception.Message}";
                    }
                    UnityEngine.Debug.Log(msg);
                    break;
            }
        }
    }
}
