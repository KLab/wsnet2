using System;

namespace WSNet2.Core
{
    public class EvSucceeded : Event, IEvResponse
    {
        public MsgType MsgType { get; private set; }
        public int MsgSeqNum { get; private set; }
        public ArraySegment<byte> Payload { get => null; }

        public EvSucceeded(SerialReader reader) : base(EvType.Succeeded, reader)
        {
            MsgType = (MsgType)reader.Get8();
            MsgSeqNum = reader.Get24();
        }
    }
}
