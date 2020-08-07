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
        static async Task callbackrunner(WSNet2Client cli, CancellationToken ct)
        {
            while (true)
            {
                ct.ThrowIfCancellationRequested();
                cli.ProcessCallback();
                await Task.Delay(1000);
            }
        }

        static async Task Main(string[] args)
        {
            WSNet2Helper.RegisterTypes();

            var rand = new Random();
            var server = "http://localhost:8080";
            var appId = "testapp";
            var pKey = "testapppkey";
            var userId = "gamemaster";
            var searchGroup = 1000;
            var maxPlayer = 3;

            var mc = new MasterClient(server, appId, pKey, searchGroup, maxPlayer, userId);
            await mc.Serve();
        }
    }
}