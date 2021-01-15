using WSNet2.Core;
using Microsoft.Extensions.Logging;
using ZLogger;

namespace WSNet2.Sample
{
    class AppLogger : WSNet2.Core.WSNet2Logger.ILogger
    {
        ILogger logger;

        public AppLogger(ILogger logger)
        {
            this.logger = logger;
        }

        public void Log(WSNet2Logger.LogLevel logLevel, string message)
        {
            logger.ZLog((LogLevel)logLevel, message);
        }

        public void Log<TPayload>(WSNet2Logger.LogLevel logLevel, TPayload payload, string message)
        {
            logger.ZLogWithPayload((LogLevel)logLevel, payload, message);
        }
    }
}