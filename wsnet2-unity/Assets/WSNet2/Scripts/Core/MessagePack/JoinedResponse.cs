using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class JoinedResponse
    {
        [Key("room_info")]
        public RoomInfo roomInfo;

        [Key("players")]
        public ClientInfo[] players;

        [Key("url")]
        public string url;

        [Key("token")]
        public AuthToken token;

        [Key("master_id")]
        public string masterId;

        [Key("deadline")]
        public uint deadline;
    }
}
