using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   Msgがエラーになったときに送られるイベント
    /// </summary>
    /// 派生型によってエラー種別を判定
    public interface EvMsgError
    {
        /// <summary>
        ///   エラーとなったMsgのMsgType
        /// </summary>
        MsgType MsgType { get; }

        /// <summary>
        ///   エラーとなったMsgのシーケンス番号
        /// </summary>
        int MsgSeqNum { get; }

        /// <summary>
        ///   エラーとなったMsgの内容
        /// </summary>
        ArraySegment<byte> Payload { get; }
    }
}
