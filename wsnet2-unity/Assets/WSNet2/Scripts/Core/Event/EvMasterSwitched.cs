namespace WSNet2.Core
{
    /// <summary>
    ///   マスタープレイヤーが交代しました
    /// </summary>
    public class EvMasterSwitched : Event
    {
        public string NewMasterId { get; private set; }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public EvMasterSwitched(SerialReader reader) : base(EvType.MasterSwitched, reader)
        {
            NewMasterId = reader.ReadString();
        }
    }
}
