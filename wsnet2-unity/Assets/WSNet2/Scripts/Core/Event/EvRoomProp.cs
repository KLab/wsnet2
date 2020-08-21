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
        public Dictionary<string, object> PublicProps;
        public Dictionary<string, object> PrivateProps;

        public EvRoomProp(SerialReader reader) : base(EvType.RoomProp, reader)
        {
            var flags = reader.ReadByte();
            Visible = (flags & 1) != 0;
            Joinable = (flags & 2) != 0;
            Watchable = (flags & 4) != 0;
            SearchGroup = reader.ReadUInt();
            MaxPlayers = reader.ReadUShort();
            ClientDeadline = reader.ReadUShort();
            PublicProps = reader.ReadDict();
            PrivateProps = reader.ReadDict();
        }
    }
}
