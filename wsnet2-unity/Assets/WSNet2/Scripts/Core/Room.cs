using System;
using System.Collections.Generic;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;

namespace WSNet2.Core
{
    public class Room
    {
        public string Id { get { return info.id; } }
        public bool Visible { get { return info.visible; } }
        public bool Joinable { get { return info.joinable; } }

        public bool Running { get; set; }

        public Player Me { get; private set; }

        RoomInfo info;
        Uri uri;
        IEventReceiver receiver;
        Dictionary<string, object> publicProps;
        Dictionary<string, object> privateProps;
        List<Player> players;

        CallbackPool callbackPool = new CallbackPool();

        public Room(JoinedResponse joined, string myId, IEventReceiver receiver)
        {
            this.info = joined.roomInfo;
            this.uri = new Uri(joined.url);
            this.receiver = receiver;
            this.Running = true;

            var reader = Serialization.NewReader(info.publicProps);
            publicProps = reader.ReadDict();

            reader = Serialization.NewReader(info.privateProps);
            privateProps = reader.ReadDict();

            players = new List<Player>(joined.players.Length);
            foreach (var p in joined.players)
            {
                var player = new Player(p);
                players.Add(player);
                if (p.Id == myId)
                {
                    Me = player;
                }
            }
        }

        public void ProcessCallback()
        {
            if (Running)
            {
                callbackPool.Process();
            }
        }

        public async Task Start()
        {
            var cts = new CancellationTokenSource();

            // todo: leave前に切断したら再接続
            try
            {
                await Connect(cts.Token);
            }
            catch(Exception e)
            {
                callbackPool.Add(()=>{
                    throw e;
                });
            }
        }

        private async Task Connect(CancellationToken ct)
        {
            var ws = new ClientWebSocket();
            ws.Options.SetRequestHeader("X-Wsnet-App", info.appId);
            ws.Options.SetRequestHeader("X-Wsnet-User", Me.Id);
            ws.Options.SetRequestHeader("X-Wsnet-LastEventSeq", "0");

            await ws.ConnectAsync(uri, ct);

            // 最初にEvPeerReadyが来る
            var ev = await ReceiveEvent(ws, ct);

            // todo: send required msgs
            // todo: start sender task

            while(true)
            {
                ct.ThrowIfCancellationRequested();

                ev = await ReceiveEvent(ws, ct);

                callbackPool.Add(() => {
                    _ = Console.Out.WriteLineAsync("msg: " + BitConverter.ToString(ev.ToArray()));
                });
            }
        }

        byte[] buf = new byte[8];

        // TODO: return Event
        private async Task<ArraySegment<byte>> ReceiveEvent(WebSocket ws, CancellationToken ct)
        {
            var pos = 0;
            while(true){
                var seg = new ArraySegment<byte>(buf, pos, buf.Length-pos);
                var ret = await ws.ReceiveAsync(seg, ct);
                pos += ret.Count;

                if (ret.CloseStatus.HasValue)
                {
                    throw new Exception("ws status:("+ret.CloseStatus.Value+") "+ret.CloseStatusDescription);
                }

                if (ret.EndOfMessage) {
                    break;
                }

                Array.Resize(ref buf, buf.Length*2);
            }

            return new ArraySegment<byte>(buf, 0, pos);
        }
    }
}
