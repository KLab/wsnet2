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
            var macKey = "testMacKey1";

            var before = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds();
            var base64 = authgen.Generate(key, cliId, macKey);

            Console.WriteLine(base64);

            var after = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds();
            var data = Convert.FromBase64String(base64);

            var nonce = new Span<byte>(data, 0, 8).ToArray();
            var tdata = new Span<byte>(data, 8, 8).ToArray();
            var hash = new Span<byte>(data, 16, 32).ToArray();
            var iv = new Span<byte>(data, 48, 16).ToArray();
            var ckey = new Span<byte>(data, 64, data.Length-64).ToArray();

            var bkey = Encoding.ASCII.GetBytes(key);

            // check timestamp
            var timestamp = tdata[0] << 56 | tdata[1] << 48 | tdata[2] << 40 |
                tdata[3] << 32 |tdata[4] << 24 | tdata[5] << 16 | tdata[6] << 8 | tdata[7];
            Assert.GreaterOrEqual(timestamp, before);
            Assert.LessOrEqual(timestamp, after);

            // check hash
            var ms = new MemoryStream();
            ms.Write(Encoding.UTF8.GetBytes(cliId));
            ms.Write(nonce);
            ms.Write(tdata);
            var hmac = new HMACSHA256(bkey);
            Assert.AreEqual(hash, hmac.ComputeHash(ms.ToArray()));

            // check macKey
            var aes = Aes.Create();
            var rdr = new StreamReader(
                new CryptoStream(
                    new MemoryStream(ckey),
                    Aes.Create().CreateDecryptor(SHA256.Create().ComputeHash(bkey), iv),
                    CryptoStreamMode.Read));
            Assert.AreEqual(rdr.ReadLine(), macKey);
        }
    }
}
