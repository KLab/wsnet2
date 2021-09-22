using System;
using WSNet2.Core;

namespace WSNet2
{
    /// <summary>
    /// Unity環境でデフォルトで使用されるLogger
    /// </summary>
    public class DefaultUnityLogger : IWSNet2Logger<WSNet2LogPayload>
    {
        public WSNet2LogPayload Payload { get; } = new WSNet2LogPayload();

        public void Log(WSNet2LogLevel logLevel, Exception e, string format, params object[] param)
        {
            var msg = $"{string.Format(format, param)}\nPayload: User={Payload.UserId}, Room={Payload.RoomId}, RoomNum={Payload.RoomNum}";

            switch (logLevel)
            {
                case WSNet2LogLevel.Critical:
                case WSNet2LogLevel.Error:
                    UnityEngine.Debug.LogError(msg);
                    break;
                case WSNet2LogLevel.Warning:
                    UnityEngine.Debug.LogWarning(msg);
                    break;
                case WSNet2LogLevel.Information:
                case WSNet2LogLevel.Debug:
                case WSNet2LogLevel.Trace:
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
