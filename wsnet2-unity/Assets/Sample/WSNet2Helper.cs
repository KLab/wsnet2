using System;
using System.Text;
using System.Security.Cryptography;
using WSNet2.Core;

public static class WSNetHelper
{
    public static AuthData GenAuthData(string key, string userid)
    {
        var auth = new AuthData();

        auth.Timestamp = ((DateTimeOffset)DateTime.UtcNow).ToUnixTimeSeconds().ToString();

        var rng = new RNGCryptoServiceProvider();
        var nbuf = new byte[8];
        rng.GetBytes(nbuf);
        auth.Nonce = BitConverter.ToString(nbuf).Replace("-", "").ToLower();

        var str = userid + auth.Timestamp + auth.Nonce;
        var hmac = new HMACSHA256(Encoding.ASCII.GetBytes(key));
        var hash = hmac.ComputeHash(Encoding.ASCII.GetBytes(str));
        auth.Hash = BitConverter.ToString(hash).Replace("-", "").ToLower();

        return auth;
    }

    public static void RegisterTypes()
    {
        Serialization.Register<SampleClient.StrMessage>(1);
    }
}