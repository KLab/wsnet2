using System;
using WSNet2;

namespace Sample.Logic
{
    public static class WSNet2Helper
    {
        /// <summary>部屋のPublicPropertyのキー</summary>
        public class PubKey
        {
            /// <summary>ゲーム名 ("pong")</summary>
            public const string Game = "game";

            /// <summary>ゲームの状態 (string; GameStateCode)</summary>
            public const string State = "state";

            /// <summary>プレイヤー待ち状況の変更時刻 (long; unixtime)</summary>
            public const string Updated = "updated";

            /// <summary>ランダム入室絞り込み用 プレイヤー数 (byte)</summary>
            public const string PlayerNum = "playernum";
        }

        // ゲーム名
        public const string GameName = "pong";

        // 検索グループ
        public const uint SearchGroup = 1000;

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
    public static class RoomExtension
    {
        public static GameStateCode GameState(this Room room)
        {
            if (room != null && room.PublicProps.TryGetValue(WSNet2Helper.PubKey.State, out var s)) {
                return (GameStateCode)Enum.Parse(typeof(GameStateCode), (string)s);
            }
            return GameStateCode.None;
        }
    }
}
