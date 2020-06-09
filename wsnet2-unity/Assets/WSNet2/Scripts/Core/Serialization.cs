using System;
using System.Collections.Generic;
using System.Runtime.Serialization;

namespace WSNet2.Core
{
    public class Serialization
    {
        const int WRITER_BUFSIZE = 1024;

        static Dictionary<byte, System.Type> registeredTypes = new Dictionary<byte, System.Type>();
        static Dictionary<System.Type, byte> registeredIDs = new Dictionary<System.Type, byte>();

        public static SerialWriter NewWriter(int size = WRITER_BUFSIZE)
        {
            return new SerialWriter(size, registeredIDs);
        }

        public static SerialReader NewReader(ArraySegment<byte> buf)
        {
            return new SerialReader(buf, registeredTypes);
        }

        /// <summary>
        /// カスタム型を登録する
        /// </summary>
        /// <typeparam name="T">型</typeparam>
        /// <param name="classID">クラス識別子</param>
        public static void Register<T>(byte classID) where T : IWSNetSerializable, new()
        {
            if (registeredTypes.ContainsKey(classID))
            {
                var msg = string.Format("ClassID '{0}' is aleady used for {1}", classID, typeof(T));
                throw new ArgumentException(msg);
            }

            var t = typeof(T);
            if (registeredIDs.ContainsKey(t))
            {
                var msg = string.Format("Type '{0}' is aleady registered as {1}", typeof(T), classID);
                throw new ArgumentException(msg);
            }

            registeredIDs[t] = classID;
            registeredTypes[classID] = t;
        }
    }

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
        /// <param name="size">size</param>
        void Deserialize(SerialReader reader, int size);
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
    public class SerializationException : Exception
    {
        public SerializationException() : base()
        {
        }

        public SerializationException(string message) : base(message)
        {
        }

        public SerializationException(string message, Exception innerException) : base(message, innerException)
        {
        }

        protected SerializationException(SerializationInfo info, StreamingContext context) : base (info, context)
        {
        }
    }

}
