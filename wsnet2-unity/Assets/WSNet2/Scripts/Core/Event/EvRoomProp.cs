using System.Collections.Generic;

namespace WSNet2.Core
{
    public class EvRoomProp : Event
    {
        public bool Visible;
        public bool Joinable;
        public bool Watchable;
        public uint SearchGroup;
        public ushort MaxPlayers;
        public ushort ClientDeadline;

        Dictionary<string, object> publicProps;
        Dictionary<string, object> privateProps;
        bool gotPublicProps;
        bool gotPrivateProps;

        public EvRoomProp(SerialReader reader) : base(EvType.RoomProp, reader)
        {
            var flags = reader.ReadByte();
            Visible = (flags & 1) != 0;
            Joinable = (flags & 2) != 0;
            Watchable = (flags & 4) != 0;
            SearchGroup = reader.ReadUInt();
            MaxPlayers = reader.ReadUShort();
            ClientDeadline = reader.ReadUShort();
            gotPublicProps = false;
            gotPrivateProps = false;
        }

        public Dictionary<string, object> GetPublicProps(IDictionary<string, object> recycle = null)
        {
            if (!gotPublicProps)
            {
                publicProps = reader.ReadDict(recycle);
                gotPublicProps = true;
            }

            return publicProps;
        }

        public Dictionary<string, object> GetPrivateProps(IDictionary<string, object> recycle = null)
        {
            if (!gotPrivateProps)
            {
                _ = GetPublicProps();
                privateProps = reader.ReadDict(recycle);
                gotPrivateProps = true;
            }

            return privateProps;
        }
    }
}
