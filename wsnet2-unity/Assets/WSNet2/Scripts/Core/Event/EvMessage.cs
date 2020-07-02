namespace WSNet2.Core
{
    /// <summary>
    ///   通常メッセージイベント
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     MsgTargetやMsgBroadcastによって送られたイベント
    ///   </para>
    /// </remarks>
    public class EvMessage : Event
    {
        /// <summary>送信者</summary>
        public string SenderID { get; private set; }

        IWSNetSerializable body;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     メッセージの中身はまだデシリアライズしない
        ///   </para>
        /// </remarks>
        public EvMessage(SerialReader reader) : base(EvType.Message, reader)
        {
            body = null;
            SenderID = reader.ReadString();
        }

        /// <summary>
        ///   メッセージの中身をデシリアライズして取得
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     recycleを再利用するにはUnityのメインスレッドから呼ぶ必要がある
        ///   </para>
        /// </remarks>
        public T Body<T>(T recycle = null) where T : class, IWSNetSerializable, new()
        {
            if (body == null)
            {
                body = reader.ReadObject(recycle);
            }

            return body as T;
        }
    }
}
