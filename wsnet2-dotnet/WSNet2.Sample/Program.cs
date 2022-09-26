using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.Threading;
using System.Threading.Tasks;
using Sample.Logic;
using ZLogger;
using Microsoft.Extensions.Logging;

namespace WSNet2.Sample
{
    public class Program
    {
        static void PrintHelp()
        {
            Console.WriteLine(
                "usage: WSNet2.Sample [option]\n" +
                "options:\n" +
                "  -s, --server        wsnet2 lobby address (default: http://localhost:8080)\n" +
                "  -m, --master [num]  number of game master\n" +
                "  -b, --bot    [num]  number of bot.\n" +
                "  -?, -h, --help      show this message\n");
        }

        static async Task Main(string[] args)
        {
            // finallyを必ず実行するための例外ハンドリング
            try
            {
                await main(args);
            }
            catch
            {
                throw;
            }
        }

        static async Task main(string[] args)
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

            var masterCount = 0;
            var botCount = 0;
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

                if (args[i] == "-m" || args[i] == "--master")
                {
                    if (i + 1 < args.Length && args[i+1][0] != '-')
                    {
                        masterCount += int.Parse(args[++i]);
                    }
                    else
                    {
                        masterCount++;
                    }
                }

                if (args[i] == "-b" || args[i] == "--bot")
                {
                    if (i + 1 < args.Length && args[i+1][0] != '-')
                    {
                        botCount += int.Parse(args[++i]);
                    }
                    else
                    {
                        botCount++;
                    }
                }

                if (args[i] == "-h" || args[i] == "-?" || args[i] == "--help")
                {
                    PrintHelp();
                    return;
                }
            }

            if (masterCount + botCount == 0)
            {
                PrintHelp();
                return;
            }

            var pid = Process.GetCurrentProcess().Id;
            var tasks = new List<Task>();

            for (var i = 0; i < masterCount; i++)
            {
                var userId = $"master_{pid}_{i}";
                var logger = new AppLogger(loggerFactory.CreateLogger(userId));
                var master = new MasterClient(logger);
                var task = Task.Run(async () => await master.Serve(server, appId, pKey, searchGroup, userId));
                task.ConfigureAwait(false);
                tasks.Add(task);
            }
            for (var i = 0; i < botCount; i++)
            {
                var userId = $"bot_{pid}_{i}";
                var logger = new AppLogger(loggerFactory.CreateLogger(userId));
                var bot = new BotClient(logger);
                var task = Task.Run(async () => await bot.Serve(server, appId, pKey, searchGroup, userId));
                task.ConfigureAwait(false);
                tasks.Add(task);
            }

            await Task.WhenAll(tasks);
        }
    }
}
