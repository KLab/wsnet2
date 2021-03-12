using System;
using NUnit.Framework;

namespace WSNet2.Core.Test
{
    public class WSNet2LoggerTests
    {
        [Test]
        public void ExampleWSNet2Logger()
        {
            WSNet2Logger.Debug("Hello World");
            WSNet2Logger.Info("Hello World");
            WSNet2Logger.Warning("Hello World");
            WSNet2Logger.Error("Hello World");

            WSNet2Logger.DebugWithPayload(new { Debug = "Foo" }, "Hello With Payload");
            WSNet2Logger.InfoWithPayload(new { Info = "Foo" }, "Hello With Payload");
            WSNet2Logger.WarningWithPayload(new { Warning = "Foo" }, "Hello With Payload");
            WSNet2Logger.ErrorWithPayload(new { Error = "Foo" }, "Hello With Payload");
        }

        [Test]
        public void ExampleExceptionLog()
        {
            Exception e = null;
            try
            {
                throw new Exception("TestException");
            }
            catch (Exception e_)
            {
                e = e_;
            }

            WSNet2Logger.Debug(e, "Hello Exception");
            WSNet2Logger.Info(e, "Hello Exception");
            WSNet2Logger.Warning(e, "Hello Exception");
            WSNet2Logger.Error(e, "Hello Exception");

            WSNet2Logger.DebugWithPayload(e, new { Debug = "Foo" }, "Hello With Payload");
            WSNet2Logger.InfoWithPayload(e, new { Info = "Foo" }, "Hello With Payload");
            WSNet2Logger.WarningWithPayload(e, new { Warning = "Foo" }, "Hello With Payload");
            WSNet2Logger.ErrorWithPayload(e, new { Error = "Foo" }, "Hello With Payload");
        }
    }
}