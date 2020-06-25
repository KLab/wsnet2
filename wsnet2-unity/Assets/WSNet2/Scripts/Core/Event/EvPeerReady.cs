namespace WSNet2.Core
{
    public class EvPeerReady : Event
    {
        public int LastSeqNum { get; private set; }

        public EvPeerReady(SerialReader reader) : base(EvType.PeerReady, reader)
        {
            LastSeqNum = reader.Get24();
        }
    }
}
