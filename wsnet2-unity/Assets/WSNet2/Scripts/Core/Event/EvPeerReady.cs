namespace WSNet2.Core
{
    /// <summary>
    ///   Gameサーバ側のPeerの準備完了通知
    /// </summary>
    public class EvPeerReady : Event
    {
        /// <summary>
        ///   Game側が最後に受け取ったMsgの通し番号
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     クライアントはwebsocket接続した後、このEventを受け取るまでMsgを送信しない。
        ///     このEvent受信後、LasMsgSeqNumの次のMsgから送りはじめる（再送含む）。
        ///   </para>
        /// </remarks>
        public int LastMsgSeqNum { get; private set; }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public EvPeerReady(SerialReader reader) : base(EvType.PeerReady, reader)
        {
            LastMsgSeqNum = reader.Get24();
        }
    }
}
