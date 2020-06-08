using System;
using System.Runtime.Serialization;

namespace WSNet2.Core
{
    /// <summary>
    /// Websocketで送受信するカスタム型はこのインターフェイスを実装する
    /// </summary>
    public interface IWSNetSerializable
    {
        /// <summary>
        /// Serializeする.
        /// </summary>
        /// <param name="writer">writer</param>
        void Serialize(SerialWriter writer);

        /// <summary>
        /// Deserializeする.
        /// </summary>
        /// <param name="reader">reader</param>
        void Deserialize(SerialReader reader);
    }

    enum Type : byte
    {
        Null = 0,
        False,
        True,
        SByte,
        Byte,
        Short,
        UShort,
        Int,
        UInt,
        Long,
        ULong,
        Float,
        Double,
        Str8,
        Str16,
        Obj,
        List,
        Dict,
        Bytes,
    }

    [Serializable()]
    public class DeserializeException : Exception
    {
        public DeserializeException() : base()
        {
        }

        public DeserializeException(string message) : base(message)
        {
        }

        public DeserializeException(string message, Exception innerException) : base(message, innerException)
        {
        }

        protected DeserializeException(SerializationInfo info, StreamingContext context) : base (info, context)
        {
        }
    }

}
