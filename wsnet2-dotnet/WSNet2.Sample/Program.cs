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

        const int MaxMasterClient = 1;

        static void Main(string[] args)
        {
            WSNet2Helper.RegisterTypes();

            var rand = new Random();
            var server = "http://localhost:8080";
            var appId = "testapp";
            var pKey = "testapppkey";
            var searchGroup = 1000;

            var tasks = new Task[MaxMasterClient];
            for (int i = 0; i < MaxMasterClient; i++)
            {
                var userId = "gamemaster" + rand.Next(1000, 9999).ToString();
                tasks[i] = Task.Run(async () =>
                    await new MasterClient().Serve(server, appId, pKey, searchGroup, userId));
            }

            Task.WaitAll(tasks);
        }
    }
}