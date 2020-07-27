namespace WSNet2.Core
{
    /// <summary>
    ///   プレイヤーが退室しました
    /// </summary>
    public class EvLeaved : Event
    {
        /// <summary>退室したPlayer</summary>
        public string ClientID { get; private set; }
        public string MasterID { get; private set; }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public EvLeaved(SerialReader reader) : base(EvType.Leaved, reader)
        {
            ClientID = reader.ReadString();
            MasterID = reader.ReadString();
        }
    }
}
