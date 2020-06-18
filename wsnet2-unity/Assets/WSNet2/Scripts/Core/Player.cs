using System.Collections.Generic;

namespace WSNet2.Core
{
    public class Player
    {
        public string Id { get { return info.Id; } }

        ClientInfo info;
        Dictionary<string, object> props;

        public Player(ClientInfo info)
        {
            this.info = info;

            var reader = Serialization.NewReader(info.Props);
            props = reader.ReadDict();
        }

    }

}
