using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;

namespace WSNet2.Core
{
    using ReadFunc = Serialization.ReadFunc;

    public class SerialReader
    {
        UTF8Encoding utf8 = new UTF8Encoding();
        Hashtable typeIDs;
        ReadFunc[] readFuncs;
        ArraySegment<byte> arrSeg;
        IList<byte> buf;
        int pos;

        public SerialReader(ArraySegment<byte> buf, Hashtable typeIDs, ReadFunc[] readFuncs)
        {
            this.arrSeg = buf;
            this.buf = (IList<byte>)buf;
            this.pos = 0;
            this.typeIDs = typeIDs;
            this.readFuncs = readFuncs;
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

        public float ReadFloat()
        {
            checkType(Type.Float);
            throw new NotImplementedException();
        }

        public float ReadDouble()
        {
            checkType(Type.Double);
            throw new NotImplementedException();
        }

        public string ReadString()
        {
            var t = checkType(Type.Str8, Type.Str16);
            var len = (t == Type.Str8) ? Get8() : Get16();
            var str = utf8.GetString(arrSeg.Array, arrSeg.Offset + pos, len);
            pos += len;
            return str;
        }

        public T ReadObject<T>(T recycle = default) where T : class, IWSNetSerializable, new()
        {
            if (checkType(Type.Obj, Type.Null) == Type.Null)
            {
                return null;
            }

            var t = typeof(T);
            var tid = typeIDs[t];
            if (tid == null)
            {
                var msg = string.Format("Type {0} is not registered", t);
                throw new SerializationException(msg);
            }

            var id = (byte)Get8();
            if (id != (byte)tid)
            {
                var msg = string.Format("Type mismatch {0} wants {1}", tid, id);
                throw new SerializationException(msg);
            }

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

        public List<object> ReadList(IReadOnlyList<object> recycle = null)
        {
            checkType(Type.List);
            var count = Get8();
            var list = new List<object>(count);
            var recycleCount = (recycle != null) ? recycle.Count : 0;

            for (var i = 0; i < count; i++)
            {
                var elem = readElement((i < recycleCount) ? recycle[i] : null);
                list.Add(elem);
            }

            return list;
        }

        public object[] ReadArray(IReadOnlyList<object> recycle = null)
        {
            checkType(Type.List);
            var count = Get8();
            var list = new object[count];
            var recycleCount = (recycle != null) ? recycle.Count : 0;

            for (var i = 0; i < count; i++)
            {
                var elem = readElement((i < recycleCount) ? recycle[i] : null);
                list[i] = elem;
            }

            return list;
        }

        public List<T> ReadList<T>(IReadOnlyList<T> recycle = null) where T : class, IWSNetSerializable, new()
        {
            checkType(Type.List);
            var count = Get8();
            var list = new List<T>(count);
            var recycleCount = (recycle != null) ? recycle.Count : 0;

            for (var i = 0; i < count; i++)
            {
                var len = Get16();
                var st = pos;
                var elem = ReadObject<T>((i < recycleCount) ? recycle[i] : null);
                list.Add(elem);
                pos = st + len;
            }

            return list;
        }

        public T[] ReadArray<T>(IReadOnlyList<T> recycle = null) where T : class, IWSNetSerializable, new()
        {
            checkType(Type.List);
            var count = Get8();
            var list = new T[count];
            var recycleCount = (recycle != null) ? recycle.Count : 0;

            for (var i = 0; i < count; i++)
            {
                var len = Get16();
                var st = pos;
                var elem = ReadObject<T>((i < recycleCount) ? recycle[i] : null);
                list[i] = elem;
                pos = st + len;
            }

            return list;
        }

        public Dictionary<string, object> ReadDict(IDictionary<string, object> recycle = null)
        {
            checkType(Type.Dict);
            var dict = new Dictionary<string, object>();
            var count = Get8();

            for (var i = 0; i < count; i++)
            {
                var klen = Get8();
                var key = string.Intern(utf8.GetString(arrSeg.Array, arrSeg.Offset + pos, klen));
                pos += klen;

                var val = readElement(
                    (recycle != null && recycle.ContainsKey(key)) ? recycle[key] : null);

                dict[key] = val;
            }

            return dict;
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

        object readElement(object recycle)
        {
            var len = Get16();
            var st = pos;
            checkLength(len);

            object elem = null;

            var t = (Type)Enum.ToObject(typeof(Type), buf[pos]);
            switch(t)
            {
                case Type.Null:
                    break;
                case Type.True:
                    elem = true;
                    break;
                case Type.False:
                    elem = false;
                    break;
                case Type.SByte:
                    elem = ReadSByte();
                    break;
                case Type.Byte:
                    elem = ReadByte();
                    break;
                case Type.Short:
                    elem = ReadShort();
                    break;
                case Type.UShort:
                    elem = ReadUShort();
                    break;
                case Type.Int:
                    elem = ReadInt();
                    break;
                case Type.UInt:
                    elem = ReadUInt();
                    break;
                case Type.Long:
                    elem = ReadLong();
                    break;
                case Type.ULong:
                    elem = ReadULong();
                    break;
                case Type.Float:
                    elem = ReadFloat();
                    break;
                case Type.Double:
                    elem = ReadDouble();
                    break;
                case Type.Str8:
                case Type.Str16:
                    elem = ReadString();
                    break;
                case Type.Obj:
                    var cid = buf[pos+1];
                    var read = readFuncs[cid];
                    if (read == null)
                    {
                        throw new SerializationException(
                            string.Format("ClassID {0} is not registered", cid));
                    }
                    elem = read(this, recycle);
                    break;
                case Type.List:
                    elem = ReadList(recycle as IReadOnlyList<object>);
                    break;
                case Type.Dict:
                    elem = ReadDict(recycle as IDictionary<string, object>);
                    break;
                default:
                    throw new SerializationException(
                        string.Format("Type {0} is not implemented", t));
            }

            pos = st + len;
            return elem;
        }

    }
}
