namespace WSNet2.Core
{
    public class EvPeerReady : Event
    {
        public int MsgSeqNum { get; private set; }

        public EvPeerReady(EvType type, SerialReader reader) : base(type, reader)
        {
            MsgSeqNum = reader.Get24();
        }
    }
}
