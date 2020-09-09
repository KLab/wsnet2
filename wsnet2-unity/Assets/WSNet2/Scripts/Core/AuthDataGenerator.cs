using System;
using System.Security.Cryptography;
using System.Text;

namespace WSNet2.Core
{
    public class AuthDataGenerator
    {
        Random rand = new Random();

        public string Generate(string key, string clientId)
        {
            // clientId, nonce 64bit, timestamp 64bit, hmac 256bit
            var l = Encoding.UTF8.GetByteCount(clientId);
            var data = new byte[l+8+8+32];

            // clientId
            Encoding.UTF8.GetBytes(clientId, 0, clientId.Length, data, 0);

            // nonce
            data[l+0] = (byte)rand.Next(256);
            data[l+1] = (byte)rand.Next(256);
            data[l+2] = (byte)rand.Next(256);
            data[l+3] = (byte)rand.Next(256);
            data[l+4] = (byte)rand.Next(256);
            data[l+5] = (byte)rand.Next(256);
            data[l+6] = (byte)rand.Next(256);
            data[l+7] = (byte)rand.Next(256);

            // timestamp
            var now = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds();
            data[l+8] = (byte)((now >> 56) & 0xff);
            data[l+9] = (byte)((now >> 48) & 0xff);
            data[l+10] = (byte)((now >> 40) & 0xff);
            data[l+11] = (byte)((now >> 32) & 0xff);
            data[l+12] = (byte)((now >> 24) & 0xff);
            data[l+13] = (byte)((now >> 16) & 0xff);
            data[l+14] = (byte)((now >> 8) & 0xff);
            data[l+15] = (byte)(now & 0xff);

            var hmac = new HMACSHA256(Encoding.ASCII.GetBytes(key));
            var hash = hmac.ComputeHash(data, 0, l+16);

            Buffer.BlockCopy(hash, 0, data, l+16, 32);

            return Convert.ToBase64String(data, l, 16+32);
        }
    }
}
