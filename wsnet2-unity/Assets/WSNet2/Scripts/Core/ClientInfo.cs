using System.Collections.Generic;
using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class ClientInfo
    {
        [Key("id")]
        public string Id;

        [Key("props")]
        public byte[] Props;

        public ClientInfo()
        {
        }

        public ClientInfo(string id, IDictionary<string, object> props)
        {
            this.Id = id;

            var writer = Serialization.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(props);
                this.Props = writer.ArraySegment().ToArray();
            }
        }
    }
}
