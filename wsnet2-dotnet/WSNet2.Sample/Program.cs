using System;
using System.Collections.Generic;
using System.Linq;
using System.Security.Cryptography;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using WSNet2.Core;
using Sample.Logic;

namespace WSNet2.Sample
{
    public class Program
    {
        static void Main(string[] args)
        {
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
                    tasks[i] = Task.Run(async () =>
                        await new BotClient().Serve(server, appId, pKey, searchGroup, userId));
                    Thread.Sleep(1000);
                }
            }

            Task.WaitAll(tasks);
        }
    }
}