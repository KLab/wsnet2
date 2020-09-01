using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   Msgのレスポンス
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     base.Typeで種別を判定
    ///     - Succeeded
    ///     - PermissionDenied
    ///     - TargetNotFound
    ///   </para>
    /// </remarks>
    public class EvResponse : Event
    {
        /// <summary>
        ///   元となるMsgのMsgType
        /// </summary>
        public MsgType MsgType { get; private set; }

        /// <summary>
        ///   元となるMsgのシーケンス番号
        /// </summary>
        public int MsgSeqNum { get; private set; }

        /// <summary>
        ///   元となるMsgの内容（成功時は空）
        /// </summary>
        public ArraySegment<byte> Payload { get; private set; }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public EvResponse(EvType type, SerialReader reader) : base(type, reader)
        {
            MsgType = (MsgType)reader.Get8();
            MsgSeqNum = reader.Get24();
            Payload = reader.GetRest();
        }
    }
}
