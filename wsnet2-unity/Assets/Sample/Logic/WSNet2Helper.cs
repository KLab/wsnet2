using System;
using System.Text;
using System.Collections.Generic;
using System.Security.Cryptography;
using WSNet2.Core;

namespace Sample.Logic
{
    public static class WSNet2Helper
    {
        public static byte[] Serialize<T>(T v) where T : class, IWSNetSerializable
        {
            // FIXME: もう少し便利な方法が提供してほしい
            var w = WSNet2.Core.Serialization.GetWriter();
            lock (w)
            {
                w.Reset();
                w.Write<T>(v);
                var seg = w.ArraySegment();
                var ret = new byte[seg.Count];
                System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
                return ret;
            }
        }

        public static byte[] Serialize(int v)
        {
            // FIXME: もう少し便利な方法が提供してほしい
            var w = WSNet2.Core.Serialization.GetWriter();
            lock (w)
            {
                w.Reset();
                w.Write(v);
                var seg = w.ArraySegment();
                var ret = new byte[seg.Count];
                System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
                return ret;
            }
        }

        public static byte[] Serialize(string v)
        {
            // FIXME: もう少し便利な方法が提供してほしい
            var w = WSNet2.Core.Serialization.GetWriter();
            lock (w)
            {
                w.Reset();
                w.Write(v);
                var seg = w.ArraySegment();
                var ret = new byte[seg.Count];
                System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
                return ret;
            }
        }

        public static byte[] Serialize(Dictionary<string, object> v)
        {
            // FIXME: もう少し便利な方法が提供してほしい
            var w = WSNet2.Core.Serialization.GetWriter();
            lock (w)
            {
                w.Reset();
                w.Write(v);
                var seg = w.ArraySegment();
                var ret = new byte[seg.Count];
                System.Array.Copy(seg.Array, seg.Offset, ret, 0, seg.Count);
                return ret;
            }
        }

        /// <summary>
        /// WSNet2にログインするための認証用データを生成する
        /// </summary>
        /// <param name="key">アプリケーション固有の鍵</param>
        /// <param name="userid">ユーザID</param>
        /// <returns></returns>
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

        /// <summary>
        /// WSNet2のシリアライザでシリアライズする独自型の登録を行う
        /// プロセス開始後1度だけ呼び出すこと
        /// </summary>
        public static void RegisterTypes()
        {
            Serialization.Register<Sample.Logic.GameState>(10);
            Serialization.Register<Sample.Logic.Bar>(11);
            Serialization.Register<Sample.Logic.Ball>(12);
            Serialization.Register<Sample.Logic.PlayerEvent>(20);
        }
    }
}