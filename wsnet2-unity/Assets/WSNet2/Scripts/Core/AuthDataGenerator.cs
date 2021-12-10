using System;
using System.IO;
using System.Security.Cryptography;
using System.Text;

namespace WSNet2.Core
{
    public class AuthDataGenerator
    {
        Random rand = new Random();

        public string Generate(string key, string clientId, string macKey)
        {
            // result: nonce 64bits, timestamp 64bits, hmac 256bits, iv 64bits, crypted macKey *bits

            byte[] bkey = Encoding.ASCII.GetBytes(key);
            byte[] bcid = Encoding.UTF8.GetBytes(clientId);
            byte[] bmkey = Encoding.ASCII.GetBytes(macKey);
            using var ms = new MemoryStream(bcid.Length + 8 + 8 + 32 + 8 + macKey.Length + 16);
            var offset = GenAuthData(ms, bkey, bcid);

            // iv
            var iv = new byte[16];
            for (var i=0; i<16; i++) iv[i] = (byte)rand.Next(256);
            ms.Write(iv, 0, iv.Length);

            // crypted macKey
            using var aes = Aes.Create();
            aes.Key = SHA256.Create().ComputeHash(bkey);
            aes.IV = iv;
            using var cs = new CryptoStream(ms, aes.CreateEncryptor(), CryptoStreamMode.Write);
            cs.Write(bmkey, 0, bmkey.Length);
            cs.FlushFinalBlock();

            return Convert.ToBase64String(ms.ToArray(), offset, (int)ms.Length - offset);
        }

        public string GenerateForConnect(string key, string clientId)
        {
            byte[] bkey = Encoding.ASCII.GetBytes(key);
            byte[] bcid = Encoding.UTF8.GetBytes(clientId);
            var ms = new MemoryStream(bcid.Length+8+8);
            var offset = GenAuthData(ms, bkey, bcid);
            return Convert.ToBase64String(ms.ToArray(), offset, (int)ms.Length - offset);
        }

        int GenAuthData(MemoryStream ms, byte[] bkey, byte[] bcid)
        {
            // clientId
            ms.Write(bcid, 0, bcid.Length);
            var offset = (int)ms.Length;

            // nonce
            for (var i=0; i<8; i++)
            {
                ms.WriteByte((byte)rand.Next(256));
            }

            // timestamp
            var now = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds();
            ms.WriteByte((byte)((now >> 56) & 0xff));
            ms.WriteByte((byte)((now >> 48) & 0xff));
            ms.WriteByte((byte)((now >> 40) & 0xff));
            ms.WriteByte((byte)((now >> 32) & 0xff));
            ms.WriteByte((byte)((now >> 24) & 0xff));
            ms.WriteByte((byte)((now >> 16) & 0xff));
            ms.WriteByte((byte)((now >> 8) & 0xff));
            ms.WriteByte((byte)(now & 0xff));

            var hmac = new HMACSHA256(bkey);
            var hash = hmac.ComputeHash(ms.ToArray());

            ms.Write(hash, 0, hash.Length);

            return offset;
        }
    }
}
