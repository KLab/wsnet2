using System;
using System.Collections.Generic;
using System.Runtime.Serialization;

namespace WSNet2.Core
{
    public class Serialization
    {
        public delegate object ReadFunc(SerialReader reader, object recycle);

        const int WRITER_BUFSIZE = 1024;

        static Dictionary<System.Type, byte> registeredTypes = new Dictionary<System.Type, byte>();
        static Dictionary<byte, ReadFunc> readFuncs = new Dictionary<byte, ReadFunc>();
        static SerialWriter writer;

        public static SerialWriter NewWriter(int size = WRITER_BUFSIZE)
        {
            var w = new SerialWriter(size, registeredTypes);
            if (writer == null)
            {
                writer = w;
            }

            return w;
        }

        public static SerialWriter GetWriter()
        {
            if (writer == null)
            {
                writer = new SerialWriter(WRITER_BUFSIZE, registeredTypes);
            }

            return writer;
        }

        public static SerialReader NewReader(ArraySegment<byte> buf)
        {
            return new SerialReader(buf, registeredTypes, readFuncs);
        }

        /// <summary>
        /// カスタム型を登録する
        /// </summary>
        /// <typeparam name="T">型</typeparam>
        /// <param name="classID">クラス識別子</param>
        public static void Register<T>(byte classID) where T : class, IWSNetSerializable, new()
        {
            var t = typeof(T);
            if (registeredTypes.ContainsKey(t))
            {
                var msg = string.Format("Type '{0}' is aleady registered as {1}", t, classID);
                throw new ArgumentException(msg);
            }

            if (readFuncs.ContainsKey(classID))
            {
                var msg = string.Format("ClassID '{0}' is aleady used for {1}", classID, t);
                throw new ArgumentException(msg);
            }

            registeredTypes[t] = classID;
            readFuncs[classID] = (reader, obj) => reader.ReadObject<T>(obj as T);
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
