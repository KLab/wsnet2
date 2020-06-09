using System;
using System.Text;
using System.Collections.Generic;

namespace WSNet2.Core
{
    public class SerialWriter
    {
        const int MINSIZE = 1024;

        UTF8Encoding utf8 = new UTF8Encoding();
        Dictionary<System.Type, byte> typeCodes;
        int pos;
        byte[] buf;

        /// <summary>
        /// コンストラクタ
        /// </summary>
        public SerialWriter(int size, Dictionary<System.Type, byte> typeCodes)
        {
            var s = MINSIZE;
            while (s < size)
            {
                s *= 2;
            }

            this.pos = 0;
            this.buf = new byte[s];
            this.typeCodes = typeCodes;
        }

        public void Reset()
        {
            pos = 0;
        }

        public ArraySegment<byte> ArraySegment()
        {
            return new ArraySegment<byte>(buf, 0, pos);
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
            buf[pos+1] = (byte)((int)v - (int)sbyte.MinValue);
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
            buf[pos+1] = v;
            pos += 2;
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
        ///   文字列を書き込む
        /// </summary>
        /// <param name="v">値</param>
        public void Write(string v)
        {
            var len = utf8.GetByteCount(v);
            if (len <= byte.MaxValue)
            {
                expand(len+2);
                buf[pos] = (byte)Type.Str8;
                pos++;
                Put8(len);
            }
            else if (len <= ushort.MaxValue)
            {
                expand(len+3);
                buf[pos] = (byte)Type.Str16;
                pos++;
                Put16(len);
            }
            else
            {
                var msg = string.Format("string too long: {0}", len);
                throw new SerializationException(msg);
            }

            utf8.GetBytes(v, 0, v.Length, buf, pos);
            pos += len;
        }

        /// <summary>
        ///   登録された型のオブジェクトを書き込む
        /// </summary>
        /// <typeparam name="T">型</param>
        /// <param name="v">値</param>
        public void Write<T>(T v) where T : IWSNetSerializable, new()
        {
            var t = typeof(T);
            if (!typeCodes.ContainsKey(t))
            {
                var msg = string.Format("Type {0} is not registered", t);
                throw new SerializationException(msg);
            }

            expand(4);
            buf[pos] = (byte)Type.Obj;
            buf[pos+1] = typeCodes[t];
            pos += 4;

            var start = pos;
            v.Serialize(this);

            var size = pos - start;
            if (size > ushort.MaxValue) {
                var msg = string.Format("Serialized data is too big: {0}", size);
                throw new SerializationException(msg);
            }

            buf[start-2] = (byte)((size & 0xff00) >> 8);
            buf[start-1] = (byte)(size & 0xff);
        }

        public void Put8(int v)
        {
            buf[pos] = (byte)(v & 0xff);
            pos++;
        }

        public void Put16(int v)
        {
            buf[pos] = (byte)((v & 0xff00) >> 8);
            buf[pos+1] = (byte)(v & 0xff);
            pos += 2;
        }

        public void Put24(int v)
        {
            buf[pos] = (byte)((v & 0xff0000) >> 16);
            buf[pos+1] = (byte)((v & 0xff00) >> 8);
            buf[pos+2] = (byte)(v & 0xff);
            pos += 3;
        }

        public void Put32(long v)
        {
            buf[pos] = (byte)((v & 0xff000000) >> 24);
            buf[pos+1] = (byte)((v & 0xff0000) >> 16);
            buf[pos+2] = (byte)((v & 0xff00) >> 8);
            buf[pos+3] = (byte)(v & 0xff);
            pos += 4;
        }

        public void Put64(ulong v)
        {
            buf[pos] = (byte)((v & 0xff00000000000000) >> 56);
            buf[pos+1] = (byte)((v & 0xff000000000000) >> 48);
            buf[pos+2] = (byte)((v & 0xff0000000000) >> 40);
            buf[pos+3] = (byte)((v & 0xff00000000) >> 32);
            buf[pos+4] = (byte)((v & 0xff000000) >> 24);
            buf[pos+5] = (byte)((v & 0xff0000) >> 16);
            buf[pos+6] = (byte)((v & 0xff00) >> 8);
            buf[pos+7] = (byte)(v & 0xff);
            pos += 8;
        }

        private void expand(int size)
        {
            int len = buf.Length;
            if (len < pos+size)
            {
                while (len < pos+size) {
                    len *= 2;
                }
                Array.Resize(ref buf, len);
            }
        }

    }

}
