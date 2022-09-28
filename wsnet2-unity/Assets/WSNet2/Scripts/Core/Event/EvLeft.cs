namespace WSNet2
{
    /// <summary>
    ///   プレイヤーが退室しました
    /// </summary>
    public class EvLeft : Event
    {
        /// <summary>退室したPlayer</summary>
        public string ClientID { get; private set; }
        public string MasterID { get; private set; }
        public string Message { get; private set; }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public EvLeft(SerialReader reader) : base(EvType.Left, reader)
        {
            ClientID = reader.ReadString();
            MasterID = reader.ReadString();
            Message = reader.ReadString();
        }
    }
}
