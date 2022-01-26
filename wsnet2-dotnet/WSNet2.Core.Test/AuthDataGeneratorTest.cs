using NUnit.Framework;
using System;
using System.IO;
using System.Security.Cryptography;
using System.Text;

namespace WSNet2.Core.Test
{
    public class AuthDataGeneratorTest
    {
        [Test]
        public void TestGenerate()
        {
            var authgen = new AuthDataGenerator();

            var key = "testAppKey1";
            var cliId = "testClient1";

            var before = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds();
            var authdata = authgen.Generate(key, cliId);

            var after = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds();

            // check bearer
            var data = Convert.FromBase64String(authdata.Bearer.Substring("Bearar ".Length));
            var nonce = new Span<byte>(data, 0, 8).ToArray();
            var tdata = new Span<byte>(data, 8, 8).ToArray();
            var hash = new Span<byte>(data, 16, 32).ToArray();

            // check timestamp
            var timestamp = tdata[0] << 56 | tdata[1] << 48 | tdata[2] << 40 |
                tdata[3] << 32 | tdata[4] << 24 | tdata[5] << 16 | tdata[6] << 8 | tdata[7];
            Assert.GreaterOrEqual(timestamp, before);
            Assert.LessOrEqual(timestamp, after);

            // check hash
            var ms = new MemoryStream();
            ms.Write(Encoding.UTF8.GetBytes(cliId));
            ms.Write(nonce);
            ms.Write(tdata);
            var hmac = new HMACSHA256(Encoding.ASCII.GetBytes(key));
            Assert.AreEqual(hash, hmac.ComputeHash(ms.ToArray()));

            // check mackey
            var encdata = Convert.FromBase64String(authdata.EncryptedMACKey);
            var encKey = new Span<byte>(encdata, 16, encdata.Length - 16).ToArray();
            using var aes = Aes.Create();
            aes.Key = SHA256.Create().ComputeHash(Encoding.ASCII.GetBytes(key));
            aes.IV = new Span<byte>(encdata, 0, 16).ToArray();
            aes.Padding = PaddingMode.Zeros;
            var rdr = new StreamReader(
                new CryptoStream(
                    new MemoryStream(encKey), aes.CreateDecryptor(), CryptoStreamMode.Read));
            Assert.AreEqual(rdr.ReadLine(), authdata.MACKey);
        }
    }
}
