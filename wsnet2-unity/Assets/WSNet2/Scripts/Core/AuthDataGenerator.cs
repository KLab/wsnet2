using System;
using System.IO;
using System.Security.Cryptography;
using System.Text;

namespace WSNet2.Core
{
    public class AuthDataGenerator
    {
        Random rand = new Random();

        /// <summary>
        /// WSNet2Clientに渡すAuthDataを生成する
        /// </summary>
        /// <param name="key">wsnet2に登録したapp key</param>
        /// <param name="clientId">wsnet2上でのclientId</param>
        /// <returns>AuthData</returns>
        /// <remarks>
        /// 本番運用ではkeyはクライアントに含めないでください。
        /// つまり、この生成処理はAPIサーバで行います。
        /// このAPI通信でmacKeyを暗号化することで中間者攻撃を防げます。
        /// </remarks>
        public AuthData Generate(string key, string clientId)
        {
            var bearer = GenerateBearer(key, clientId);
            var macKey = RandomString(16);
            var encMKey = EncryptMACKey(key, macKey);

            return new AuthData(bearer: bearer, macKey: macKey, encMKey: encMKey);
        }

        public string GenerateBearer(string key, string clientId)
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

            return $"Bearer {Convert.ToBase64String(data, l, 16+32)}";
        }

        string RandomString(int n)
        {
            const string pool = "0123456789ABDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz";
            var s = new StringBuilder(n);
            for (var i = 0; i < n; i++)
            {
                s.Append(pool[rand.Next(pool.Length)]);
            }

            return s.ToString();
        }

        string EncryptMACKey(string key, string macKey)
        {
            using var aes = Aes.Create();
            aes.Padding = PaddingMode.Zeros;

            var bmkey = Encoding.ASCII.GetBytes(macKey);
            var ms = new MemoryStream(macKey.Length + aes.BlockSize/4);

            // iv
            var iv = new byte[aes.BlockSize/8];
            for (var i = 0; i < iv.Length; i++) iv[i] = (byte)rand.Next(256);
            ms.Write(iv, 0, iv.Length);

            // encrypt macKey
            aes.Key = SHA256.Create().ComputeHash(Encoding.ASCII.GetBytes(key));
            aes.IV = iv;
            using var cs = new CryptoStream(ms, aes.CreateEncryptor(), CryptoStreamMode.Write);
            cs.Write(bmkey, 0, bmkey.Length);
            cs.FlushFinalBlock();

            return Convert.ToBase64String(ms.ToArray());
        }
    }
}
