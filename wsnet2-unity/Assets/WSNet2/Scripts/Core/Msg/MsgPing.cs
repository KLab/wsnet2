using System;
using System.Security.Cryptography;

namespace WSNet2.Core
{
    public class MsgPing
    {
        public ArraySegment<byte> Value { get; private set; }

        byte[] buf;
        HMAC hmac;
        int hsize;

        public MsgPing(HMAC hmac)
        {
            this.hmac = hmac;
            this.hsize = hmac.HashSize / 8;
            this.buf = new byte[9 + hsize];
            buf[0] = (byte)MsgType.Ping;

            Value = new ArraySegment<byte>(buf);
        }

        public ulong SetTimestamp()
        {
            var now = DateTime.UtcNow;
            var unix = (ulong)((DateTimeOffset)now).ToUnixTimeMilliseconds();

            buf[1] = (byte)((unix & 0xff00000000000000) >> 56);
            buf[2] = (byte)((unix & 0xff000000000000) >> 48);
            buf[3] = (byte)((unix & 0xff0000000000) >> 40);
            buf[4] = (byte)((unix & 0xff00000000) >> 32);
            buf[5] = (byte)((unix & 0xff000000) >> 24);
            buf[6] = (byte)((unix & 0xff0000) >> 16);
            buf[7] = (byte)((unix & 0xff00) >> 8);
            buf[8] = (byte)(unix & 0xff);

            byte[] hash;
            lock (hmac)
            {
                hash = hmac.ComputeHash(buf, 0, 9);
            }

            Buffer.BlockCopy(hash, 0, buf, 9, hsize);

            return unix;
        }
    }
}
