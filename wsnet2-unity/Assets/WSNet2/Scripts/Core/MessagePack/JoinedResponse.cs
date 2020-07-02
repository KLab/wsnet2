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
    }
}
