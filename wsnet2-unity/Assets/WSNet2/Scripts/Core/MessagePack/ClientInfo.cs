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
                var arr = (IList<byte>)writer.ArraySegment();
                this.Props = new byte[arr.Count];
                arr.CopyTo(this.Props, 0);
            }
        }
    }
}
