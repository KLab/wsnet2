using System;

namespace WSNet2.Core
{
    public class SerialReader
    {
        ArraySegment<byte> buf;
        int pos;

        public SerialReader(ArraySegment<byte> array)
        {
            buf = array;
            pos = 0;
        }

        void checkLength(int want)
        {
            var rest = buf.Count - pos;
            if (rest < want)
            {
                var msg = String.Format("Not enough data: {0} < {1}", rest, want);
                throw new DeserializeException(msg);
            }
        }

        Type checkType(Type want)
        {
            var t = (Type)buf[pos];
            if (t != want) {
                var msg = String.Format("Type mismatch: {0} wants {1}", t, want);
                throw new DeserializeException(msg);
            }

            return t;
        }

        Type checkType(Type want1, Type want2)
        {
            var t = (Type)buf[pos];
            if (t != want1 && t != want2) {
                var msg = String.Format("Type mismatch: {0} wants {1} or {2}", t, want1, want2);
                throw new DeserializeException(msg);
            }

            return t;
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
            checkLength(2);
            checkType(Type.SByte);
            var b = (sbyte)((int)buf[pos+1] + (int)sbyte.MinValue);
            pos += 2;
            return b;
        }

        public byte ReadByte()
        {
            checkLength(2);
            checkType(Type.Byte);
            var b = buf[pos+1];
            pos += 2;
            return b;
        }

        public short ReadShort()
        {
            checkLength(3);
            checkType(Type.Short);
            var n = (int)buf[pos+1] << 8;
            n += (int)buf[pos+2];
            pos += 3;
            return (short)(n + (int)short.MinValue);
        }

        public ushort ReadUShort()
        {
            checkLength(3);
            checkType(Type.UShort);
            var n = (int)buf[pos+1] << 8;
            n += (int)buf[pos+2];
            pos += 3;
            return (ushort)n;
        }

        public int ReadInt()
        {
            checkLength(5);
            checkType(Type.Int);
            var n = (long)buf[pos+1] << 24;
            n += (long)buf[pos+2] << 16;
            n += (long)buf[pos+3] << 8;
            n += (long)buf[pos+4];
            pos += 5;
            return (int)(n + (long)int.MinValue);
        }

        public uint ReadUInt()
        {
            checkLength(5);
            checkType(Type.UInt);
            var n = (uint)buf[pos+1] << 24;
            n += (uint)buf[pos+2] << 16;
            n += (uint)buf[pos+3] << 8;
            n += (uint)buf[pos+4];
            pos += 5;
            return n;
        }

        public long ReadLong()
        {
            checkLength(9);
            checkType(Type.Long);
            var n = (long)buf[pos+1] << 56;
            n += (long)buf[pos+2] << 48;
            n += (long)buf[pos+3] << 40;
            n += (long)buf[pos+4] << 32;
            n += (long)buf[pos+5] << 24;
            n += (long)buf[pos+6] << 16;
            n += (long)buf[pos+7] << 8;
            n += (long)buf[pos+8];
            return n - long.MinValue;
        }


        public ulong ReadULong()
        {
            checkLength(9);
            checkType(Type.ULong);
            var n = (ulong)buf[pos+1] << 56;
            n += (ulong)buf[pos+2] << 48;
            n += (ulong)buf[pos+3] << 40;
            n += (ulong)buf[pos+4] << 32;
            n += (ulong)buf[pos+5] << 24;
            n += (ulong)buf[pos+6] << 16;
            n += (ulong)buf[pos+7] << 8;
            n += (ulong)buf[pos+8];
            return n;
        }
    }
}
