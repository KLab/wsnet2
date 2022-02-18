namespace WSNet2.Core
{
    /// <summary>
    ///   RPCイベント
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     EvType.MessageのメッセージをRPCとして利用する。
    ///   </para>
    /// </remarks>
    public class EvRPC : Event
    {
        /// <summary>送信者</summary>
        public string SenderID { get; private set; }

        public byte RpcID { get; private set; }

        public SerialReader Reader { get { return reader; } }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     メッセージの中身はまだデシリアライズしない。
        ///     RPC呼び出し時にメインスレッドでデシリアライズする。
        ///   </para>
        /// </remarks>
        public EvRPC(SerialReader reader) : base(EvType.Message, reader)
        {
            SenderID = reader.ReadString();
            RpcID = reader.ReadByte();
        }
    }
}
