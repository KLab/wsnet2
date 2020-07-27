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

    [MessagePackObject]
    public class JoinParam
    {
        [Key("RoomId")]
        public string roomId;

        [Key("ClientInfo")]
        public ClientInfo clientInfo;
    }

    [MessagePackObject]
    public class JoinByNumberParam
    {
        [Key("RoomNumber")]
        public int roomNumber;

        [Key("ClientInfo")]
        public ClientInfo clientInfo;
    }

    public enum OpType : byte
    {
        Equal = 0,
        Not,
        LessThan,
        LessEqual,
        GreaterThan,
        GreaterEqual,
    }

    [MessagePackObject]
    public class PropQuery
    {
        [Key("Key")]
        public string key;

        [Key("Op")]
        public OpType op;

        [Key("Val")]
        public byte[] val;
    }

    [MessagePackObject]
    public class JoinByQueriesParam
    {
        [Key("SearchGroup")]
        public uint searchGroup;

        [Key("Queries")]
        public PropQuery[][] queries;

        [Key("ClientInfo")]
        public ClientInfo clientInfo;
    }

}
