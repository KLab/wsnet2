namespace WSNet2.Core
{
    public class EvMessage : Event
    {
        public string SenderID { get; private set; }

        IWSNetSerializable body;

        public EvMessage(SerialReader reader) : base(EvType.Message, reader)
        {
            body = null;
            SenderID = reader.ReadString();
        }

        public T Body<T>(T recycle = null) where T : class, IWSNetSerializable, new()
        {
            if (body == null)
            {
                body = reader.ReadObject(recycle);
            }

            return body as T;
        }
    }
}
