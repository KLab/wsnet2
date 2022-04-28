using System.Collections.Generic;
using MessagePack;

namespace WSNet2
{
    [MessagePackObject]
    public class CreateParam
    {
        [Key("room")]
        public RoomOption roomOption;

        [Key("client")]
        public ClientInfo clientInfo;

        [Key("emk")]
        public string encryptedMACKey;
    }

    [MessagePackObject]
    public class JoinParam
    {
        [Key("query")]
        public List<List<Query.Condition>> queries;

        [Key("client")]
        public ClientInfo clientInfo;

        [Key("emk")]
        public string encryptedMACKey;
    }

    [MessagePackObject]
    public class SearchParam
    {
        [Key("group")]
        public uint group;

        [Key("query")]
        public List<List<Query.Condition>> queries;

        [Key("limit")]
        public int limit;

        [Key("joinable")]
        public bool checkJoinable;

        [Key("watchable")]
        public bool checkWatchable;
    }

    [MessagePackObject]
    public class SearchByIdsParam
    {
        [Key("ids")]
        public string[] ids;

        [Key("query")]
        public List<List<Query.Condition>> queries;
    }
}
