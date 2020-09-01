using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   Msgのレスポンスとして送られるイベントのインターフェイス
    /// </summary>
    /// 派生型によってエラー種別を判定
    public interface IEvResponse
    {
        /// <summary>
        ///   元となるMsgのMsgType
        /// </summary>
        MsgType MsgType { get; }

        /// <summary>
        ///   元となるMsgのシーケンス番号
        /// </summary>
        int MsgSeqNum { get; }

        /// <summary>
        ///   元となるMsgの内容（成功時は空）
        /// </summary>
        ArraySegment<byte> Payload { get; }
    }
}
