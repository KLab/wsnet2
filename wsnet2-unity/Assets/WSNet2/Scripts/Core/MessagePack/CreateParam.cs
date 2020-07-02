using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class CreateParam
    {
        [Key("RoomOption")]
        public RoomOption roomOption;

        [Key("ClientInfo")]
        public ClientInfo clientInfo;
    }
}
