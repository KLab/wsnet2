using System;
using System.Text;
using System.Collections;
using System.Collections.Generic;
using System.Security.Cryptography;

namespace WSNet2
{
    /// <summary>
    ///   方を保存するシリアライザ
    /// </summary>
    public class SerialWriter
    {
        const int MINSIZE = 1024;

        UTF8Encoding utf8 = new UTF8Encoding();
        Hashtable types;
        int pos;
        byte[] buf;

        /// <summary>
        /// コンストラクタ
        /// </summary>
        public SerialWriter(int size, Hashtable types)//Dictionary<System.Type, byte> types)
        {
            var s = MINSIZE;
            while (s < size)
            {
                s *= 2;
            }

            this.pos = 0;
            this.buf = new byte[s];
            this.types = types;
        }

        public void Reset()
        {
            pos = 0;
        }

        public ArraySegment<byte> ArraySegment()
        {
            return new ArraySegment<byte>(buf, 0, pos);
        }

        public byte[] ToArray()
        {
            var arr = new byte[pos];
            Buffer.BlockCopy(buf, 0, arr, 0, pos);
            return arr;
        }

        /// <summary>
        /// Nullを書き込む
        /// </summary>
        public void Write()
        {
            expand(1);
            buf[pos] = (byte)Type.Null;
            pos++;
        }

        /// <summary>
        ///   Bool値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(bool v)
        {
            expand(1);
            buf[pos] = (byte)(v ? Type.True : Type.False);
            pos++;
        }

        /// <summary>
        ///   SByte値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(sbyte v)
        {
            expand(2);
            buf[pos] = (byte)Type.SByte;
            buf[pos + 1] = (byte)((int)v - (int)sbyte.MinValue);
            pos += 2;
        }

        /// <summary>
        ///   Byte値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(byte v)
        {
            expand(2);
            buf[pos] = (byte)Type.Byte;
            buf[pos + 1] = v;
            pos += 2;
        }

        /// <summary>
        ///   Char値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(char v)
        {
            expand(3);
            buf[pos] = (byte)Type.Char;
            pos++;
            Put16(v);
        }

        /// <summary>
        ///   Short値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(short v)
        {
            expand(3);
            buf[pos] = (byte)Type.Short;
            pos++;
            var n = (int)v - (int)short.MinValue;
            Put16(n);
        }

        /// <summary>
        ///   UShort値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(ushort v)
        {
            expand(3);
            buf[pos] = (byte)Type.UShort;
            pos++;
            Put16(v);
        }

        /// <summary>
        ///   Int値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(int v)
        {
            expand(5);
            buf[pos] = (byte)Type.Int;
            pos++;
            var n = (long)v - (long)int.MinValue;
            Put32(n);
        }

        /// <summary>
        ///   UInt値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(uint v)
        {
            expand(5);
            buf[pos] = (byte)Type.UInt;
            pos++;
            Put32(v);
        }

        /// <summary>
        ///   Long値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(long v)
        {
            expand(9);
            buf[pos] = (byte)Type.Long;
            pos++;
            ulong n = (ulong)(v - long.MinValue);
            Put64(n);
        }

        /// <summary>
        ///   ULong値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(ulong v)
        {
            expand(9);
            buf[pos] = (byte)Type.ULong;
            pos++;
            Put64(v);
        }

        /// <summary>
        ///   Float値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(float v)
        {
            expand(5);
            buf[pos] = (byte)Type.Float;
            pos++;
            var b = BitConverter.ToInt32(BitConverter.GetBytes(v), 0);
            if ((b & (1 << 31)) == 0)
            {
                b ^= 1 << 31;
            }
            else
            {
                b = ~b;
            }

            Put32(b);
        }

        /// <summary>
        ///   Double値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(double v)
        {
            expand(9);
            buf[pos] = (byte)Type.Double;
            pos++;
            var b = BitConverter.DoubleToInt64Bits(v);
            if ((b & (1L << 63)) == 0)
            {
                b ^= 1L << 63;
            }
            else
            {
                b = ~b;
            }

            Put64((ulong)b);
        }

        /// <summary>
        ///   文字列を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(string v)
        {
            if (v == null)
            {
                Write();
                return;
            }

            var len = utf8.GetByteCount(v);
            if (len <= byte.MaxValue)
            {
                expand(len + 2);
                buf[pos] = (byte)Type.Str8;
                pos++;
                Put8(len);
            }
            else if (len <= ushort.MaxValue)
            {
                expand(len + 3);
                buf[pos] = (byte)Type.Str16;
                pos++;
                Put16(len);
            }
            else
            {
                var msg = string.Format("string too long: {0}", len);
                throw new WSNet2SerializerException(msg);
            }

            utf8.GetBytes(v, 0, v.Length, buf, pos);
            pos += len;
        }

        /// <summary>
        ///   登録された型のオブジェクトを書き込む
        /// </summary>
        /// <Typeparam name="T">型</param>
        /// <param name="v">値</param>
        public void Write<T>(T v) where T : class, IWSNet2Serializable
        {
            if (v == null)
            {
                Write();
                return;
            }

            var t = v.GetType();
            var id = types[t];
            if (id == null)
            {
                var msg = string.Format("Type {0} is not registered", t);
                throw new WSNet2SerializerException(msg);
            }

            expand(4);
            buf[pos] = (byte)Type.Obj;
            buf[pos + 1] = (byte)id;
            pos += 4;

            var start = pos;
            v.Serialize(this);

            var size = pos - start;
            if (size > ushort.MaxValue)
            {
                var msg = string.Format("Serialized data is too big: {0}", size);
                throw new WSNet2SerializerException(msg);
            }

            buf[start - 2] = (byte)((size & 0xff00) >> 8);
            buf[start - 1] = (byte)(size & 0xff);
        }

        /// <summary>
        /// シリアライズ可能な値のリストを書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(IEnumerable v)
        {
            if (v == null)
            {
                Write();
                return;
            }

            expand(2);
            buf[pos] = (byte)Type.List;
            var countpos = pos + 1;
            pos += 2;

            var count = 0;
            foreach (var elem in v)
            {
                count++;
                if (count > byte.MaxValue)
                {
                    throw new WSNet2SerializerException("Too many list content");
                }

                writeElement(elem);
            }

            buf[countpos] = (byte)count;
        }

        /// <summary>
        ///   辞書型の値を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(IDictionary<string, object> v)
        {
            if (v == null)
            {
                Write();
                return;
            }

            expand(2);
            buf[pos] = (byte)Type.Dict;
            pos++;

            var count = v.Count;
            if (count > byte.MaxValue)
            {
                var msg = string.Format("Too many dictionary content: {0}", count);
                throw new WSNet2SerializerException(msg);
            }
            Put8(count);

            foreach (var kv in v)
            {
                var klen = utf8.GetByteCount(kv.Key);
                if (klen > byte.MaxValue)
                {
                    var msg = string.Format("Too long key: \"{0}\"", kv.Key);
                    throw new WSNet2SerializerException(msg);
                }
                expand(klen + 1);
                Put8(klen);
                utf8.GetBytes(kv.Key, 0, kv.Key.Length, buf, pos);
                pos += klen;

                writeElement(kv.Value);
            }
        }

        /// <summary>
        ///   bool配列を書き込む
        /// </summary>
        /// <param name="vals">値</param>
        public void Write(bool[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            var len = (count + 7) / 8;
            expand(3 + len);
            buf[pos] = (byte)Type.Bools;
            pos++;
            Put16(count);

            for (var i = 0; i < count; i++)
            {
                if (i % 8 == 0)
                {
                    buf[pos + i / 8] = 0;
                }

                if (vals[i])
                {
                    buf[pos + i / 8] += (byte)(1 << (7 - (i % 8)));
                }
            }

            pos += len;
        }

        /// <summary>
        ///   sbyte配列を書き込む
        /// </summary>
        public void Write(sbyte[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count);
            buf[pos] = (byte)Type.SBytes;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                var v = (int)val - sbyte.MinValue;
                Put8(v);
            }
        }

        /// <summary>
        ///   byte配列を書き込む
        /// </summary>
        public void Write(byte[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count);
            buf[pos] = (byte)Type.Bytes;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                Put8(val);
            }
        }

        /// <summary>
        ///   char配列を書き込む
        /// </summary>
        public void Write(char[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 2);
            buf[pos] = (byte)Type.Chars;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                Put16(val);
            }
        }

        /// <summary>
        ///   short配列を書き込む
        /// </summary>
        public void Write(short[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 2);
            buf[pos] = (byte)Type.Shorts;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                var v = (int)val - short.MinValue;
                Put16(v);
            }
        }

        /// <summary>
        ///   ushort配列を書き込む
        /// </summary>
        public void Write(ushort[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 2);
            buf[pos] = (byte)Type.UShorts;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                Put16(val);
            }
        }

        /// <summary>
        ///   int配列を書き込む
        /// </summary>
        public void Write(int[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 4);
            buf[pos] = (byte)Type.Ints;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                var v = (long)val - int.MinValue;
                Put32(v);
            }
        }

        /// <summary>
        ///   uint配列を書き込む
        /// </summary>
        public void Write(uint[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 4);
            buf[pos] = (byte)Type.UInts;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                Put32(val);
            }
        }

        /// <summary>
        ///   long配列を書き込む
        /// </summary>
        public void Write(long[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 8);
            buf[pos] = (byte)Type.Longs;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                Put64((ulong)(val - long.MinValue));
            }
        }

        /// <summary>
        ///   ulong配列を書き込む
        /// </summary>
        public void Write(ulong[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 8);
            buf[pos] = (byte)Type.ULongs;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                Put64(val);
            }
        }

        /// <summary>
        ///   float配列を書き込む
        /// </summary>
        public void Write(float[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 4);
            buf[pos] = (byte)Type.Floats;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                var b = BitConverter.ToInt32(BitConverter.GetBytes(val), 0);
                if ((b & (1 << 31)) == 0)
                {
                    b ^= 1 << 31;
                }
                else
                {
                    b = ~b;
                }

                Put32(b);
            }
        }

        /// <summary>
        ///   double配列を書き込む
        /// </summary>
        public void Write(double[] vals)
        {
            if (vals == null)
            {
                Write();
                return;
            }

            var count = vals.Length;
            if (count > ushort.MaxValue)
            {
                var msg = string.Format("Too long array: {0}", count);
                throw new WSNet2SerializerException(msg);
            }

            expand(3 + count * 4);
            buf[pos] = (byte)Type.Doubles;
            pos++;
            Put16(count);

            foreach (var val in vals)
            {
                var b = BitConverter.DoubleToInt64Bits(val);
                if ((b & (1L << 63)) == 0)
                {
                    b ^= 1L << 63;
                }
                else
                {
                    b = ~b;
                }

                Put64((ulong)b);
            }
        }

        /// <summary>
        ///   Bool型のみの辞書を書き込む
        /// </summary>
        public void Write(IDictionary<string, bool> v)
        {
            if (v == null)
            {
                Write();
                return;
            }

            expand(2);
            buf[pos] = (byte)Type.Dict;
            pos++;

            var count = v.Count;
            if (count > byte.MaxValue)
            {
                var msg = string.Format("Too many dictionary content: {0}", count);
                throw new WSNet2SerializerException(msg);
            }
            Put8(count);

            foreach (var kv in v)
            {
                var klen = utf8.GetByteCount(kv.Key);
                if (klen > byte.MaxValue)
                {
                    var msg = string.Format("Too long key: \"{0}\"", kv.Key);
                    throw new WSNet2SerializerException(msg);
                }
                expand(klen + 4);
                Put8(klen);
                utf8.GetBytes(kv.Key, 0, kv.Key.Length, buf, pos);
                pos += klen;
                Put16(1);
                Write(kv.Value);
            }
        }

        /// <summary>
        ///   ulongのみの辞書を書き込む
        /// </summary>
        public void Write(IDictionary<string, ulong> v)
        {
            if (v == null)
            {
                Write();
                return;
            }

            expand(2);
            buf[pos] = (byte)Type.Dict;
            pos++;

            var count = v.Count;
            if (count > byte.MaxValue)
            {
                var msg = string.Format("Too many dictionary content: {0}", count);
                throw new WSNet2SerializerException(msg);
            }
            Put8(count);

            foreach (var kv in v)
            {
                var klen = utf8.GetByteCount(kv.Key);
                if (klen > byte.MaxValue)
                {
                    var msg = string.Format("Too long key: \"{0}\"", kv.Key);
                    throw new WSNet2SerializerException(msg);
                }
                expand(klen + 3 + 9);
                Put8(klen);
                utf8.GetBytes(kv.Key, 0, kv.Key.Length, buf, pos);
                pos += klen;
                Put16(9);
                Write(kv.Value);
            }
        }

        public void Put8(int v)
        {
            buf[pos] = (byte)(v & 0xff);
            pos++;
        }

        public void Put16(int v)
        {
            buf[pos] = (byte)((v & 0xff00) >> 8);
            buf[pos + 1] = (byte)(v & 0xff);
            pos += 2;
        }

        public void Put24(int v)
        {
            buf[pos] = (byte)((v & 0xff0000) >> 16);
            buf[pos + 1] = (byte)((v & 0xff00) >> 8);
            buf[pos + 2] = (byte)(v & 0xff);
            pos += 3;
        }

        public void Put32(long v)
        {
            buf[pos] = (byte)((v & 0xff000000) >> 24);
            buf[pos + 1] = (byte)((v & 0xff0000) >> 16);
            buf[pos + 2] = (byte)((v & 0xff00) >> 8);
            buf[pos + 3] = (byte)(v & 0xff);
            pos += 4;
        }

        public void Put64(ulong v)
        {
            buf[pos] = (byte)((v & 0xff00000000000000) >> 56);
            buf[pos + 1] = (byte)((v & 0xff000000000000) >> 48);
            buf[pos + 2] = (byte)((v & 0xff0000000000) >> 40);
            buf[pos + 3] = (byte)((v & 0xff00000000) >> 32);
            buf[pos + 4] = (byte)((v & 0xff000000) >> 24);
            buf[pos + 5] = (byte)((v & 0xff0000) >> 16);
            buf[pos + 6] = (byte)((v & 0xff00) >> 8);
            buf[pos + 7] = (byte)(v & 0xff);
            pos += 8;
        }

        public void AppendHMAC(HMAC hmac)
        {
            byte[] hash;
            lock (hmac)
            {
                hash = hmac.ComputeHash(buf, 0, pos);
            }

            expand(hash.Length);
            Buffer.BlockCopy(hash, 0, buf, pos, hash.Length);
            pos += hash.Length;
        }

        private void expand(int size)
        {
            int len = buf.Length;
            if (len < pos + size)
            {
                while (len < pos + size)
                {
                    len *= 2;
                }
                Array.Resize(ref buf, len);
            }
        }

        private void writeElement(object elem)
        {
            expand(2);
            pos += 2;
            var start = pos;

            switch (elem)
            {
                case null:
                    Write();
                    break;
                case bool e:
                    Write(e);
                    break;
                case sbyte e:
                    Write(e);
                    break;
                case byte e:
                    Write(e);
                    break;
                case short e:
                    Write(e);
                    break;
                case ushort e:
                    Write(e);
                    break;
                case int e:
                    Write(e);
                    break;
                case uint e:
                    Write(e);
                    break;
                case long e:
                    Write(e);
                    break;
                case ulong e:
                    Write(e);
                    break;
                case float e:
                    Write(e);
                    break;
                case double e:
                    Write(e);
                    break;
                case string e:
                    Write(e);
                    break;
                case IWSNet2Serializable e:
                    Write(e);
                    break;
                case IDictionary<string, object> e:
                    Write(e);
                    break;
                case bool[] e:
                    Write(e);
                    break;
                case sbyte[] e:
                    Write(e);
                    break;
                case byte[] e:
                    Write(e);
                    break;
                case short[] e:
                    Write(e);
                    break;
                case ushort[] e:
                    Write(e);
                    break;
                case int[] e:
                    Write(e);
                    break;
                case uint[] e:
                    Write(e);
                    break;
                case long[] e:
                    Write(e);
                    break;
                case ulong[] e:
                    Write(e);
                    break;
                case float[] e:
                    Write(e);
                    break;
                case double[] e:
                    Write(e);
                    break;
                case IDictionary<string, bool> e:
                    Write(e);
                    break;
                case IDictionary<string, ulong> e:
                    Write(e);
                    break;
                case IEnumerable e:
                    Write(e);
                    break;
                default:
                    throw new WSNet2SerializerException($"unknown element type: {elem.GetType()}");
            }

            var size = pos - start;
            buf[start - 2] = (byte)((size & 0xff00) >> 8);
            buf[start - 1] = (byte)(size & 0xff);
        }

    }

}
