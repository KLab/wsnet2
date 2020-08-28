using System;

namespace WSNet2.Core
{
    public class EvPermissionDeny : Event, EvMsgError
    {
        public MsgType MsgType { get; private set; }
        public int MsgSeqNum { get; private set; }
        public ArraySegment<byte> Payload { get; private set; }

        public EvPermissionDeny(SerialReader reader) : base(EvType.PermissionDeny, reader)
        {
            MsgType = (MsgType)reader.Get8();
            MsgSeqNum = reader.Get24();
            Payload = reader.GetRest();
        }
    }
}
