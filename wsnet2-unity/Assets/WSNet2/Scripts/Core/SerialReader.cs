using System;
using System.Collections;
using System.Collections.Generic;
using System.Text;

namespace WSNet2.Core
{
    using ReadFunc = Serialization.ReadFunc;

    /// <summary>
    ///   型を保存するデシリアライザ
    /// </summary>
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

        public ArraySegment<byte> GetRest()
        {
            var start = arrSeg.Offset + pos;
            var len = arrSeg.Count - start;
            return new ArraySegment<byte>(arrSeg.Array, start, len);
        }

        /// <summary>
        ///   bool値を取り出す
        /// </summary>
        public bool ReadBool()
        {
            checkLength(1);
            var t = checkType(Type.True, Type.False);
            return t == Type.True;
        }

        /// <summary>
        ///   sbyte値を取り出す
        /// </summary>
        public sbyte ReadSByte()
        {
            checkType(Type.SByte);
            return (sbyte)(Get8() + (int)sbyte.MinValue);
        }

        /// <summary>
        ///   byte値を取り出す
        /// </summary>
        public byte ReadByte()
        {
            checkType(Type.Byte);
            return (byte)Get8();
        }

        /// <summary>
        ///   char値を取り出す
        /// </summary>
        public char ReadChar()
        {
            checkType(Type.Char);
            return (char)Get16();
        }

        /// <summary>
        ///   short値を取り出す
        /// </summary>
        public short ReadShort()
        {
            checkType(Type.Short);
            return (short)(Get16() + (int)short.MinValue);
        }

        /// <summary>
        ///   ushort値を取り出す
        /// </summary>
        public ushort ReadUShort()
        {
            checkType(Type.UShort);
            return (ushort)Get16();
        }

        /// <summary>
        ///   int値を取り出す
        /// </summary>
        public int ReadInt()
        {
            checkType(Type.Int);
            return (int)((long)Get32() + (long)int.MinValue);
        }

        /// <summary>
        ///   uint値を取り出す
        /// </summary>
        public uint ReadUInt()
        {
            checkType(Type.UInt);
            return Get32();
        }

        /// <summary>
        ///   long値を取り出す
        /// </summary>
        public long ReadLong()
        {
            checkType(Type.Long);
            return (long)Get64() + long.MinValue;
        }

        /// <summary>
        ///   ulong値を取り出す
        /// </summary>
        public ulong ReadULong()
        {
            checkType(Type.ULong);
            return Get64();
        }

        /// <summary>
        ///   float値を取り出す
        /// </summary>
        public float ReadFloat()
        {
            checkType(Type.Float);
            var b = (int)Get32();
            if ((b & (1 << 31)) != 0)
            {
                b ^= 1 << 31;
            }
            else
            {
                b = ~b;
            }

            return BitConverter.ToSingle(BitConverter.GetBytes(b), 0);
        }

        /// <summary>
        ///   double値を取り出す
        /// </summary>
        public double ReadDouble()
        {
            checkType(Type.Double);
            var b = (long)Get64();
            if ((b & ((long)1 << 63)) != 0)
            {
                b ^= (long)1 << 63;
            }
            else
            {
                b = ~b;
            }

            return BitConverter.Int64BitsToDouble(b);
        }

        /// <summary>
        ///   string値を取り出す
        /// </summary>
        public string ReadString()
        {
            var t = checkType(Type.Str8, Type.Str16, Type.Null);
            if (t == Type.Null)
            {
                return null;
            }

            var len = (t == Type.Str8) ? Get8() : Get16();
            var str = utf8.GetString(arrSeg.Array, arrSeg.Offset + pos, len);
            pos += len;
            return str;
        }

        /// <summary>
        ///   登録された型の値を取り出す
        /// </summary>
        public T ReadObject<T>(T recycle = default) where T : class, IWSNet2Serializable, new()
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
            if (obj == null)
            {
                obj = new T();
            }

            var start = pos;
            obj.Deserialize(this, size);
            pos = start + size;

            return obj;
        }

        /// <summary>
        ///   シリアライズ可能な型のリストを取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public List<object> ReadList(List<object> recycle = null)
        {
            if (checkType(Type.List, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get8();
            var list = recycle;
            if (list == null)
            {
                list = new List<object>(count);
            }
            else if (list.Count > count)
            {
                list.RemoveRange(count, list.Count - count);
            }

            var recycleCount = (recycle != null) ? recycle.Count : 0;

            for (var i = 0; i < count; i++)
            {
                var elem = readElement((i < recycleCount) ? recycle[i] : null);

                if (list.Count > i)
                {
                    list[i] = elem;
                }
                else
                {
                    list.Add(elem);
                }
            }

            return list;
        }

        /// <summary>
        ///   シリアライズ可能な型の配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public object[] ReadArray(object[] recycle = null)
        {
            if (checkType(Type.List, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get8();
            var list = recycle;
            if (list == null || list.Length != count)
            {
                list = new object[count];
            }

            var recycleCount = (recycle != null) ? recycle.Length : 0;

            for (var i = 0; i < count; i++)
            {
                var elem = readElement((i < recycleCount) ? recycle[i] : null);
                list[i] = elem;
            }

            return list;
        }

        /// <summary>
        ///   登録された型のリストを取り出す
        /// </summary>
        /// <typeparam name="T">登録された型</typeparam>
        /// <param name="recycle">再利用するオブジェクト</param>
        public List<T> ReadList<T>(List<T> recycle = null) where T : class, IWSNet2Serializable, new()
        {
            if (checkType(Type.List, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get8();
            var list = recycle;
            if (list == null)
            {
                list = new List<T>(count);
            }
            else if (list.Count > count)
            {
                list.RemoveRange(count, list.Count - count);
            }

            var recycleCount = (recycle != null) ? recycle.Count : 0;

            for (var i = 0; i < count; i++)
            {
                var len = Get16();
                var st = pos;
                var elem = ReadObject<T>((i < recycleCount) ? recycle[i] : null);

                if (list.Count > i)
                {
                    list[i] = elem;
                }
                else
                {
                    list.Add(elem);
                }

                pos = st + len;
            }

            return list;
        }

        /// <summary>
        ///   登録された型の配列を取り出す
        /// </summary>
        /// <typeparam name="T">登録された型</typeparam>
        /// <param name="recycle">再利用するオブジェクト</param>
        public T[] ReadArray<T>(T[] recycle = null) where T : class, IWSNet2Serializable, new()
        {
            if (checkType(Type.List, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get8();
            var list = recycle;
            if (list == null || list.Length != count)
            {
                list = new T[count];
            }

            var recycleCount = (recycle != null) ? recycle.Length : 0;

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

        /// <summary>
        ///   辞書を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public Dictionary<string, object> ReadDict(IDictionary<string, object> recycle = null)
        {
            if (checkType(Type.Dict, Type.Null) == Type.Null)
            {
                return null;
            }

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

        /// <summary>
        ///   boolの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public bool[] ReadBools(bool[] recycle = null)
        {
            if (checkType(Type.Bools, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new bool[count];
            }

            int b = 0;
            for (var i = 0; i < count; i++)
            {
                if (i % 8 == 0)
                {
                    b = Get8();
                }

                vals[i] = (b & (1 << (7 - i % 8))) != 0;
            }

            return vals;
        }

        /// <summary>
        ///   sbyteの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public sbyte[] ReadSBytes(sbyte[] recycle = null)
        {
            if (checkType(Type.SBytes, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new sbyte[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (sbyte)(Get8() + sbyte.MinValue);
            }

            return vals;
        }

        /// <summary>
        ///   byteの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public byte[] ReadBytes(byte[] recycle = null)
        {
            if (checkType(Type.Bytes, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new byte[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (byte)Get8();
            }

            return vals;
        }

        /// <summary>
        ///   charの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public char[] ReadChars(char[] recycle = null)
        {
            if (checkType(Type.Chars, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new char[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (char)Get16();
            }

            return vals;
        }

        /// <summary>
        ///   shortの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public short[] ReadShorts(short[] recycle = null)
        {
            if (checkType(Type.Shorts, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new short[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (short)(Get16() + short.MinValue);
            }

            return vals;
        }

        /// <summary>
        ///   ushortの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public ushort[] ReadUShorts(ushort[] recycle = null)
        {
            if (checkType(Type.UShorts, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new ushort[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (ushort)Get16();
            }

            return vals;
        }

        /// <summary>
        ///   intの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public int[] ReadInts(int[] recycle = null)
        {
            if (checkType(Type.Ints, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new int[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (int)((long)Get32() + int.MinValue);
            }

            return vals;
        }

        /// <summary>
        ///   uintの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public uint[] ReadUInts(uint[] recycle = null)
        {
            if (checkType(Type.UInts, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new uint[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = Get32();
            }

            return vals;
        }

        /// <summary>
        ///   longの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public long[] ReadLongs(long[] recycle = null)
        {
            if (checkType(Type.Longs, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new long[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = (long)Get64() + long.MinValue;
            }

            return vals;
        }

        /// <summary>
        ///   ulongの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public ulong[] ReadULongs(ulong[] recycle = null)
        {
            if (checkType(Type.ULongs, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new ulong[count];
            }

            for (var i = 0; i < count; i++)
            {
                vals[i] = Get64();
            }

            return vals;
        }

        /// <summary>
        ///   floatの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public float[] ReadFloats(float[] recycle = null)
        {
            if (checkType(Type.Floats, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new float[count];
            }

            for (var i = 0; i < count; i++)
            {
                var b = (int)Get32();
                if ((b & (1 << 31)) != 0)
                {
                    b ^= 1 << 31;
                }
                else
                {
                    b = ~b;
                }

                vals[i] = BitConverter.ToSingle(BitConverter.GetBytes(b), 0);
            }

            return vals;
        }

        /// <summary>
        ///   doubleの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public double[] ReadDoubles(double[] recycle = null)
        {
            if (checkType(Type.Doubles, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get16();
            var vals = recycle;
            if (vals == null || vals.Length != count)
            {
                vals = new double[count];
            }

            for (var i = 0; i < count; i++)
            {
                var b = (long)Get64();
                if ((b & ((long)1 << 63)) != 0)
                {
                    b ^= (long)1 << 63;
                }
                else
                {
                    b = ~b;
                }

                vals[i] = BitConverter.Int64BitsToDouble(b);
            }

            return vals;
        }

        /// <summary>
        ///   stringの配列を取り出す
        /// </summary>
        /// <param name="recycle">再利用するオブジェクト</param>
        public string[] ReadStrings(string[] recycle = null)
        {
            if (checkType(Type.List, Type.Null) == Type.Null)
            {
                return null;
            }

            var count = Get8();
            var list = recycle;
            if (list == null || list.Length != count)
            {
                list = new string[count];
            }

            for (var i = 0; i < count; i++)
            {
                var len = Get16();
                var st = pos;
                list[i] = ReadString();
                pos = st + len;
            }

            return list;
        }

        // experimental
        public Dictionary<string, bool> ReadBoolDict()
        {
            if (checkType(Type.Dict, Type.Null) == Type.Null)
            {
                return null;
            }

            var dict = new Dictionary<string, bool>();
            var count = Get8();

            for (var i = 0; i < count; i++)
            {
                var klen = Get8();
                var key = string.Intern(utf8.GetString(arrSeg.Array, arrSeg.Offset + pos, klen));
                pos += klen + 2;
                dict[key] = ReadBool();
            }

            return dict;
        }

        // TODO: implement other primitive type dict

        // experimental
        public Dictionary<string, ulong> ReadULongDict()
        {
            if (checkType(Type.Dict, Type.Null) == Type.Null)
            {
                return null;
            }

            var dict = new Dictionary<string, ulong>();
            var count = Get8();

            for (var i = 0; i < count; i++)
            {
                var klen = Get8();
                var key = string.Intern(utf8.GetString(arrSeg.Array, arrSeg.Offset + pos, klen));
                pos += klen + 2;
                dict[key] = ReadULong();
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
            n += (int)buf[pos + 1];
            pos += 2;
            return n;
        }

        public int Get24()
        {
            checkLength(3);
            var n = (int)buf[pos] << 16;
            n += (int)buf[pos + 1] << 8;
            n += (int)buf[pos + 2];
            pos += 3;
            return n;
        }

        public uint Get32()
        {
            checkLength(4);
            var n = (uint)buf[pos] << 24;
            n += (uint)buf[pos + 1] << 16;
            n += (uint)buf[pos + 2] << 8;
            n += (uint)buf[pos + 3];
            pos += 4;
            return n;
        }

        public ulong Get64()
        {
            checkLength(8);
            var n = (ulong)buf[pos] << 56;
            n += (ulong)buf[pos + 1] << 48;
            n += (ulong)buf[pos + 2] << 40;
            n += (ulong)buf[pos + 3] << 32;
            n += (ulong)buf[pos + 4] << 24;
            n += (ulong)buf[pos + 5] << 16;
            n += (ulong)buf[pos + 6] << 8;
            n += (ulong)buf[pos + 7];
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
            if (t != want)
            {
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
            if (t != want1 && t != want2)
            {
                var msg = String.Format("Type mismatch: {0} wants {1} or {2}", t, want1, want2);
                throw new SerializationException(msg);
            }

            pos++;
            return t;
        }

        Type checkType(Type want1, Type want2, Type want3)
        {
            checkLength(1);
            var t = (Type)buf[pos];
            if (t != want1 && t != want2 && t != want3)
            {
                var msg = String.Format("Type mismatch: {0} wants {1}, {2} or {3}", t, want1, want2, want3);
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
            switch (t)
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
                    var cid = buf[pos + 1];
                    var read = readFuncs[cid];
                    if (read == null)
                    {
                        throw new SerializationException(
                            string.Format("ClassID {0} is not registered", cid));
                    }
                    elem = read(this, recycle);
                    break;
                case Type.List:
                    elem = ReadList(recycle as List<object>);
                    break;
                case Type.Dict:
                    elem = ReadDict(recycle as IDictionary<string, object>);
                    break;
                case Type.Bools:
                    elem = ReadBools(recycle as bool[]);
                    break;
                case Type.SBytes:
                    elem = ReadSBytes(recycle as sbyte[]);
                    break;
                case Type.Bytes:
                    elem = ReadBytes(recycle as byte[]);
                    break;
                case Type.Shorts:
                    elem = ReadShorts(recycle as short[]);
                    break;
                case Type.UShorts:
                    elem = ReadUShorts(recycle as ushort[]);
                    break;
                case Type.Ints:
                    elem = ReadInts(recycle as int[]);
                    break;
                case Type.UInts:
                    elem = ReadUInts(recycle as uint[]);
                    break;
                case Type.Longs:
                    elem = ReadLongs(recycle as long[]);
                    break;
                case Type.ULongs:
                    elem = ReadULongs(recycle as ulong[]);
                    break;
                case Type.Floats:
                    elem = ReadFloats(recycle as float[]);
                    break;
                case Type.Doubles:
                    elem = ReadDoubles(recycle as double[]);
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
