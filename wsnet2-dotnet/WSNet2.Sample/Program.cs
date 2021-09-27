using System;
using System.Threading;
using System.Threading.Tasks;
using Sample.Logic;
using ZLogger;
using Microsoft.Extensions.Logging;

namespace WSNet2.Sample
{
    public class Program
    {

        static void Main(string[] args)
        {
            try
            {
                main(args);
            }
            catch
            {
                throw;
            }
        }

        static void main(string[] args)
        {
            // ThreadPool を詰まりにくくするおまじない
            ThreadPool.SetMinThreads(200, 200);

            using var loggerFactory = LoggerFactory.Create(builder =>
            {
                builder.ClearProviders();
                builder.SetMinimumLevel(LogLevel.Debug);

                builder.AddZLoggerConsole(options =>
                {
                    options.EnableStructuredLogging = false;
                });

                // Add File Logging.
                builder.AddZLoggerFile("wsnet2-dotnet.log", options =>
                {
                    options.EnableStructuredLogging = true;
                });
            });

            WSNet2Helper.RegisterTypes();

            string runAs = "master";
            var numberOfClient = 1;
            var rand = new Random();
            var server = "http://localhost:8080";
            var appId = "testapp";
            var pKey = "testapppkey";
            var searchGroup = 1000;

            for (int i = 0; i < args.Length; i++)
            {
                if ((args[i] == "-s" || args[i] == "--server") && i + 1 < args.Length)
                {
                    server = args[i + 1];
                }

                if ((args[i] == "-n" || args[i] == "--num") && i + 1 < args.Length)
                {
                    numberOfClient = int.Parse(args[i + 1]);
                }

                if (args[i] == "-m" || args[i] == "--master")
                {
                    runAs = "master";
                }

                if (args[i] == "-b" || args[i] == "--bot")
                {
                    runAs = "bot";
                }
            }

            var tasks = new Task[numberOfClient];
            for (int i = 0; i < numberOfClient; i++)
            {
                if (runAs == "master")
                {
                    var userId = "gamemaster" + rand.Next(1000, 9999).ToString();
                    var logger = new AppLogger(loggerFactory.CreateLogger(userId));
                    tasks[i] = Task.Run(async () =>
                        await new MasterClient(logger).Serve(server, appId, pKey, searchGroup, userId));
                }
                else if (runAs == "bot")
                {
                    var userId = "bot" + rand.Next(1000, 9999).ToString();
                    var logger = new AppLogger(loggerFactory.CreateLogger(userId));
                    var bot = new BotClient(logger);
                    tasks[i] = bot.Serve(server, appId, pKey, searchGroup, userId);
                }
            }

            Task.WaitAll(tasks);
        }
    }
}
