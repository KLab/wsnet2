using System;
using System.Collections.Generic;

namespace WSNet2.Core
{
    public class EvPong : Event
    {
        public ulong PingTimestamp { get; private set; }
        public ulong RTT { get; private set; }
        public uint WatcherCount { get; private set; }
        public Dictionary<string, ulong> lastMsgTimestamps { get; private set; }

        public EvPong(SerialReader reader) : base(EvType.Pong, reader)
        {
            var now = (ulong)((DateTimeOffset)DateTime.Now).ToUnixTimeMilliseconds();

            PingTimestamp = reader.ReadULong();
            WatcherCount = reader.ReadUInt();
            lastMsgTimestamps = reader.ReadULongDict();

            RTT = now - PingTimestamp;
        }
    }
}
