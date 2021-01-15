using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using WSNet2.Core;
using Sample.Logic;
using ZLogger;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.DependencyInjection;

namespace WSNet2.Sample
{
    public class Program
    {

        static void Main(string[] args)
        {
            var host = Host.CreateDefaultBuilder().ConfigureLogging(logging =>
            {
                // optional(MS.E.Logging):clear default providers.
                logging.ClearProviders();

                // optional(MS.E.Logging): default is Info, you can use this or AddFilter to filtering log.
                logging.SetMinimumLevel(LogLevel.Debug);

                // Add Console Logging.
                logging.AddZLoggerConsole();

                // Add File Logging.
                logging.AddZLoggerFile("wsnet2-dotnet.log");

                // Add Rolling File Logging.
                // logging.AddZLoggerRollingFile((dt, x) => $"logs/{dt.ToLocalTime():yyyy-MM-dd}_{x:000}.log", x => x.ToLocalTime().Date, 1024);

                // Enable Structured Logging
                logging.AddZLoggerConsole(options =>
                {
                    options.EnableStructuredLogging = true;
                });
            }).Build();
            var loggerFactory = host.Services.GetRequiredService<ILoggerFactory>();
            var logger = loggerFactory.CreateLogger("Global");
            WSNet2Logger.Logger = new AppLogger(logger);

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
                    tasks[i] = Task.Run(async () =>
                        await new MasterClient().Serve(server, appId, pKey, searchGroup, userId));
                }
                else if (runAs == "bot")
                {
                    var userId = "bot" + rand.Next(1000, 9999).ToString();
                    var bot = new BotClient();
                    tasks[i] = bot.Serve(server, appId, pKey, searchGroup, userId);
                }
            }

            Task.WaitAll(tasks);
        }
    }
}