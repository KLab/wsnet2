using System;
using NUnit.Framework;

namespace WSNet2.Core.Test
{
    public class WSNet2LoggerTests
    {
        class Payload : WSNet2LogPayload
        {
            public int Foo { get; set; }

            public override string ToString()
            {
                return $"App={AppId} User={UserId} Foo={Foo}";
            }
        }

        class Logger : IWSNet2Logger<Payload>
        {
            public string output { get; private set; }

            public Payload Payload { get; } = new Payload();

            public void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args)
            {
                output = $"{logLevel}[{Payload}] {string.Format(format, args)}";
            }
        }

        [Test]
        public void ExampleWSNet2Logger()
        {
            var logger = new Logger();
            logger.Payload.Foo = 100;

            var cli = new WSNet2Client("https://example.com", "TestAppId", "TestUser", new AuthData("", "", ""), logger);

            logger.Log(WSNet2LogLevel.Warning, null, "Hello {0}", "World");

            Assert.AreEqual("Warning[App=TestAppId User=TestUser Foo=100] Hello World", logger.output);
        }
    }
}
