using WSNet2.Core;
using Microsoft.Extensions.Logging;
using ZLogger;
using System;

namespace WSNet2.Sample
{
    class AppLogger : WSNet2.Core.WSNet2Logger.ILogger
    {
        ILogger logger;

        public AppLogger(ILogger logger)
        {
            this.logger = logger;
        }

        public void Log<TPayload>(WSNet2Logger.LogLevel logLevel, Exception e, TPayload payload, string message)
        {
            logger.ZLogWithPayload((LogLevel)logLevel, e, payload, message);
        }
    }
}