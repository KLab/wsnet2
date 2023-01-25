using System;
using System.Collections.Generic;

namespace WSNet2
{
    public class EvPong : Event
    {
        public ulong PingTimestamp { get; private set; }
        public ulong RTT { get; private set; }
        public uint WatcherCount { get; private set; }

        Dictionary<string, ulong> lastMsgTimestamps;

        public EvPong(SerialReader reader) : base(EvType.Pong, reader)
        {
            var now = (ulong)((DateTimeOffset)DateTime.Now).ToUnixTimeMilliseconds();

            PingTimestamp = reader.ReadULong();
            WatcherCount = reader.ReadUInt();
            RTT = now - PingTimestamp;
        }

        public void GetLastMsgTimestamps(Dictionary<string, ulong> output)
        {
            if (lastMsgTimestamps == null)
            {
                lastMsgTimestamps = reader.ReadIntoULongDict(output);
                return;
            }

            foreach (var kv in lastMsgTimestamps)
            {
                output[kv.Key] = kv.Value;
            }
        }
    }
}
