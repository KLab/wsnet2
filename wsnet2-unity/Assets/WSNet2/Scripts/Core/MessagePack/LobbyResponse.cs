using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class LobbyResponse
    {
        [Key("msg")]
        public string msg;

        [Key("room")]
        public JoinedRoom room;

        [Key("rooms")]
        public RoomInfo[] rooms;
    }
}
