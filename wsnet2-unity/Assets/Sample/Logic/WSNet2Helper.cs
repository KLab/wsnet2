using WSNet2.Core;

namespace Sample.Logic
{
    public static class WSNet2Helper
    {
        static AuthDataGenerator authgen = new AuthDataGenerator();

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
                WSNet2Serializer.Register<Sample.Logic.GameState>(10);
                WSNet2Serializer.Register<Sample.Logic.Bar>(11);
                WSNet2Serializer.Register<Sample.Logic.Ball>(12);
                WSNet2Serializer.Register<Sample.Logic.PlayerEvent>(20);
            }
        }
    }
}
