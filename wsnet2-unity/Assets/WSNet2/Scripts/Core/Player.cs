using System;
using System.Collections.Generic;

namespace WSNet2.Core
{
    public class Player
    {
        public string Id { get; private set; }

        public Dictionary<string, object> Props;

        public Player(ClientInfo info)
        {
            Id = info.Id;
            var reader = Serialization.NewReader(new ArraySegment<byte>(info.Props));
            Props = reader.ReadDict();
        }

        public Player(string id, Dictionary<string, object> props)
        {
            Id = id;
            Props = props;
        }
    }
}
