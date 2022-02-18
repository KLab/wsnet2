using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class JoinedRoom
    {
        [Key("room_info")]
        public RoomInfo roomInfo;

        [Key("players")]
        public ClientInfo[] players;

        [Key("url")]
        public string url;

        [Key("auth_key")]
        public string authKey;

        [Key("master_id")]
        public string masterId;

        [Key("deadline")]
        public uint deadline;
    }
}
