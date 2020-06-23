using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Net.WebSockets;
using System.Threading;
using System.Threading.Tasks;

namespace WSNet2.Core
{
    public class Room
    {
        const int EvBufPoolSize = 16;
        const int EvBufInitialSize = 256;
        const int MsgPoolSize = 16;

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

        BlockingCollection<byte[]> evBufPool;
        uint evSeqNum;

        //MessagePool


        CallbackPool callbackPool = new CallbackPool();

        public Room(JoinedResponse joined, string myId, IEventReceiver receiver)
        {
            this.info = joined.roomInfo;
            this.uri = new Uri(joined.url);
            this.receiver = receiver;
            this.Running = true;
            this.evSeqNum = 0;

            evBufPool = new BlockingCollection<byte[]>(
                new ConcurrentStack<byte[]>(), EvBufPoolSize);
            for (var i = 0; i<EvBufPoolSize; i++)
            {
                evBufPool.Add(new byte[EvBufInitialSize]);
            }

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
            Console.WriteLine("room.start");

            var cts = new CancellationTokenSource();

            // todo: leave前に切断したら再接続
            try
            {
                await Connect(cts.Token);

            }
            catch(Exception e)
            {
                callbackPool.Add(()=>{
                    receiver.OnError(e);
                });
            }
        }

        private async Task Connect(CancellationToken ct)
        {
            var ws = new ClientWebSocket();
            ws.Options.SetRequestHeader("X-Wsnet-App", info.appId);
            ws.Options.SetRequestHeader("X-Wsnet-User", Me.Id);
            ws.Options.SetRequestHeader("X-Wsnet-LastEventSeq", evSeqNum.ToString());

            await ws.ConnectAsync(uri, ct);

            while(true)
            {
                ct.ThrowIfCancellationRequested();
                var ev = await ReceiveEvent(ws, ct);
                Dispatch(ev, ct);
            }
        }

        private async Task<Event> ReceiveEvent(WebSocket ws, CancellationToken ct)
        {
            var buf = evBufPool.Take(ct);
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

            return Event.Parse(new ArraySegment<byte>(buf, 0, pos));
        }

        private void Dispatch(Event ev, CancellationToken ct)
        {
            if (ev.IsRegular)
            {
                if (ev.SequenceNum != evSeqNum+1)
                {
                    throw new Exception($"invalid event sequence number: {ev.SequenceNum} wants {evSeqNum+1}");
                }

                evSeqNum = ev.SequenceNum;
            }

            switch (ev)
            {
                case EvPeerReady evPeerReady:
                    OnEvPeerReady(evPeerReady, ct);
                    break;
                case EvJoined evJoined:
                    OnEvJoined(evJoined);
                    break;


                default:
                    throw new Exception($"unknown event: {ev}");
            }

            // Event受信に使ったバッファはcallbackで参照されるので
            // callbackが呼ばれて使い終わってから返却
            callbackPool.Add(() => evBufPool.Add(ev.BufferArray));
        }

        private void OnEvPeerReady(EvPeerReady ev, CancellationToken ct)
        {
            
        }

        private void OnEvJoined(EvJoined ev)
        {
            if (ev.ClientID == Me.Id)
            {
                Me.Props = ev.GetProps(Me.Props);
                callbackPool.Add(() => receiver.OnJoined(Me));
                return;
            }

            callbackPool.Add(()=>
            {
                var player = new Player(ev.ClientID, ev.GetProps());
                players.Add(player);
                callbackPool.Add(() => receiver.OnOtherPlayerJoined(player));
            });
        }

    }
}
