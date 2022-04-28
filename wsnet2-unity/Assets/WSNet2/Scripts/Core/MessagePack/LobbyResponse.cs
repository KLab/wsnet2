using MessagePack;

namespace WSNet2
{
    [MessagePackObject]
    public class LobbyResponse
    {
        [Key("msg")]
        public string msg;

        [Key("type")]
        public LobbyResponseType type;

        [Key("room")]
        public JoinedRoom room;

        [Key("rooms")]
        public RoomInfo[] rooms;
    }

    public enum LobbyResponseType : byte
    {
        OK = 0,
        RoomLimit,
        NoRoomFound,
        RoomFull,
    }
}
