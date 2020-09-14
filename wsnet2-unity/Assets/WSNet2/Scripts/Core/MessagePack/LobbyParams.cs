using System.Collections.Generic;
using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class CreateParam
    {
        [Key("room")]
        public RoomOption roomOption;

        [Key("client")]
        public ClientInfo clientInfo;
    }

    [MessagePackObject]
    public class JoinParam
    {
        [Key("query")]
        public List<List<Query.Condition>> queries;

        [Key("client")]
        public ClientInfo clientInfo;
    }

    // TODO: SearchParam
}
