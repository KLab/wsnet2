using System.Collections.Generic;
using MessagePack;

namespace WSNet2
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

        public ClientInfo(string id, IDictionary<string, object> props = null)
        {
            this.Id = id;

            var writer = WSNet2Serializer.GetWriter();
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
