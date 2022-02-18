using System.Collections.Generic;

namespace WSNet2.Core
{
    public class EvClientProp : Event
    {
        public string ClientID;

        Dictionary<string, object> props;

        public EvClientProp(SerialReader reader) : base(EvType.ClientProp, reader)
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
