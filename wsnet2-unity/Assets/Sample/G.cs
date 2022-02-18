using UnityEngine;
using WSNet2.Core;

namespace Sample
{
    /// <summary>
    /// グローバル変数
    /// </summary>
    public static class G
    {
        /// <summary>
        /// WSNet2クライアント
        /// </summary>
        public static WSNet2Client Client;

        /// <summary>
        /// 現在アクティブなゲームルーム
        /// </summary>
        public static Room GameRoom;
    }
}
