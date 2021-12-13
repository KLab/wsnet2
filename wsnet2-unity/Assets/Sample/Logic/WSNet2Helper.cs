using WSNet2.Core;

namespace Sample.Logic
{
    public static class WSNet2Helper
    {
        static AuthDataGenerator authgen = new AuthDataGenerator();

        public static byte[] Serialize<T>(T v) where T : class, IWSNet2Serializable
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
        /// <returns>AuthData</returns>
        public static AuthData GenAuthData(string key, string userid)
        {
            return authgen.Generate(key, userid);
        }

        static bool RegisterTypesOnce = false;

        /// <summary>
        /// WSNet2のシリアライザでシリアライズする独自型の登録を行う
        /// プロセス開始後1度だけ呼び出すこと
        /// </summary>
        public static void RegisterTypes()
        {
            if (!RegisterTypesOnce)
            {
                RegisterTypesOnce = true;
                Serialization.Register<Sample.Logic.GameState>(10);
                Serialization.Register<Sample.Logic.Bar>(11);
                Serialization.Register<Sample.Logic.Ball>(12);
                Serialization.Register<Sample.Logic.PlayerEvent>(20);
            }
        }
    }
}
