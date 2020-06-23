using System;

namespace WSNet2.Core
{
    public class Event
    {
        const int regularEvType = 30;

        public enum EvType
        {
            PeerReady = 1,
            Pong,

            Joined = regularEvType,
            Leave,
            RomProp,
            ClientProp,
            Message,
        }

        public byte[] BufferArray { get; private set; }
        public EvType Type { get; private set; }
        public bool IsRegular { get{ return (int)Type >= regularEvType; } }
        public uint SequenceNum { get; private set; }

        protected SerialReader reader;

        public static Event Parse(ArraySegment<byte> buf)
        {
            var reader = Serialization.NewReader(buf);
            var type = (EvType)reader.Get8();

            Event ev;
            switch (type)
            {
                case EvType.PeerReady:
                    ev = new EvPeerReady(type, reader);
                    break;
                case EvType.Joined:
                    ev = new EvJoined(type, reader);
                    break;

                default:
                    throw new Exception($"unknown event type: {type}");
            }

            ev.BufferArray = buf.Array;
            return ev;
        }

        public Event(EvType type, SerialReader reader)
        {
            this.Type = type;
            this.reader = reader;

            if (IsRegular)
            {
                SequenceNum = reader.Get32();
            }
        }
    }
}
