using System.Collections.Generic;

namespace WSNet2.Core
{
    public class EvJoined : Event
    {
        public string ClientID { get; private set; }

        Dictionary<string, object> props;

        public EvJoined(SerialReader reader) : base(EvType.Joined, reader)
        {
            ClientID = reader.ReadString();
            props = null;
        }

        public Dictionary<string, object> GetProps(IDictionary<string, object> recycle = null)
        {
            if (props == null)
            {
                props = reader.ReadDict(recycle);
            }

            return props;
        }
    }
}
