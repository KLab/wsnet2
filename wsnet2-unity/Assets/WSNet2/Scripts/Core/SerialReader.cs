using System;
using System.Collections.Generic;
using System.Text;

namespace WSNet2.Core
{
    public class SerialReader
    {
        UTF8Encoding utf8 = new UTF8Encoding();
        Dictionary<byte, System.Type> classTypes;
        ArraySegment<byte> buf;
        int pos;

        public SerialReader(ArraySegment<byte> buf, Dictionary<byte, System.Type> classTypes)
        {
            this.buf = buf;
            this.pos = 0;
            this.classTypes = classTypes;
        }

        public bool ReadBool()
        {
            checkLength(1);
            var t = checkType(Type.True, Type.False);
            pos++;
            return t == Type.True;
        }

        public sbyte ReadSByte()
        {
            checkType(Type.SByte);
            return (sbyte)(Get8() + (int)sbyte.MinValue);
        }

        public byte ReadByte()
        {
            checkType(Type.Byte);
            return (byte)Get8();
        }

        public short ReadShort()
        {
            checkType(Type.Short);
            return (short)(Get16() + (int)short.MinValue);
        }

        public ushort ReadUShort()
        {
            checkType(Type.UShort);
            return (ushort)Get16();
        }

        public int ReadInt()
        {
            checkType(Type.Int);
            return (int)((long)Get32() + (long)int.MinValue);
        }

        public uint ReadUInt()
        {
            checkType(Type.UInt);
            return Get32();
        }

        public long ReadLong()
        {
            checkType(Type.Long);
            return (long)Get64() + long.MinValue;
        }


        public ulong ReadULong()
        {
            checkType(Type.ULong);
            return Get64();
        }

        public string ReadString()
        {
            var t = checkType(Type.Str8, Type.Str16);
            var len = (t == Type.Str8) ? Get8() : Get16();
            var str = utf8.GetString(buf.Slice(pos, len));
            pos += len;
            return str;
        }

        public T ReadObject<T>(T recycle = default) where T : IWSNetSerializable, new()
        {
            checkType(Type.Obj);
            var code = buf[pos];
            if (!classTypes.ContainsKey(code))
            {
                var msg = string.Format("code 0x{0:X2} is not registered", code);
                throw new SerializationException(msg);
            }

            var t = classTypes[code];
            if (t != typeof(T))
            {
                var msg = string.Format("Type mismatch {0} wants {1}", typeof(T), t);
                throw new SerializationException(msg);
            }

            pos++;
            checkLength(2);
            var size = Get16();

            checkLength(size);

            var obj = recycle; 
            if (obj == null) {
                obj = new T();
            }

            var start = pos;
            obj.Deserialize(this, size);
            pos = start + size;

            return obj;
        }


        public int Get8()
        {
            checkLength(1);
            var n = (int)buf[pos];
            pos++;
            return n;
        }

        public int Get16()
        {
            checkLength(2);
            var n = (int)buf[pos] << 8;
            n += (int)buf[pos+1];
            pos += 2;
            return n;
        }

        public int Get24()
        {
            checkLength(3);
            var n = (int)buf[pos] << 16;
            n += (int)buf[pos+1] << 8;
            n += (int)buf[pos+2];
            pos += 3;
            return n;
        }

        public uint Get32()
        {
            checkLength(4);
            var n = (uint)buf[pos] << 24;
            n += (uint)buf[pos+1] << 16;
            n += (uint)buf[pos+2] << 8;
            n += (uint)buf[pos+3];
            pos += 4;
            return n;
        }

        public ulong Get64()
        {
            checkLength(8);
            var n = (ulong)buf[pos] << 56;
            n += (ulong)buf[pos+1] << 48;
            n += (ulong)buf[pos+2] << 40;
            n += (ulong)buf[pos+3] << 32;
            n += (ulong)buf[pos+4] << 24;
            n += (ulong)buf[pos+5] << 16;
            n += (ulong)buf[pos+6] << 8;
            n += (ulong)buf[pos+7];
            pos += 8;
            return n;
        }


        void checkLength(int want)
        {
            var rest = buf.Count - pos;
            if (rest < want)
            {
                var msg = String.Format("Not enough data: {0} < {1}", rest, want);
                throw new SerializationException(msg);
            }
        }

        Type checkType(Type want)
        {
            checkLength(1);
            var t = (Type)buf[pos];
            if (t != want) {
                var msg = String.Format("Type mismatch: {0} wants {1}", t, want);
                throw new SerializationException(msg);
            }

            pos++;
            return t;
        }

        Type checkType(Type want1, Type want2)
        {
            checkLength(1);
            var t = (Type)buf[pos];
            if (t != want1 && t != want2) {
                var msg = String.Format("Type mismatch: {0} wants {1} or {2}", t, want1, want2);
                throw new SerializationException(msg);
            }

            pos++;
            return t;
        }
    }
}
