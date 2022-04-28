using System.Collections.Generic;

namespace WSNet2
{
    /// <summary>
    ///   プレイヤーが再入室しました
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     自分自身の入室イベントも送られる（一番最初になるはず）
    ///   </para>
    /// </remarks>
    public class EvRejoined : Event
    {
        /// <summary>プレイヤーのID</summary>
        public string ClientID { get; private set; }

        /// <summary>プロパティ（内部保持用）</summary>
        Dictionary<string, object> props;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     Propsはこの時点ではまだデシリアライズせず、
        ///     必要になったタイミングでGetPropsを呼ぶことでデシリアライズする。
        ///   </para>
        /// </remarks>
        public EvRejoined(SerialReader reader) : base(EvType.Joined, reader)
        {
            ClientID = reader.ReadString();
            props = null;
        }

        /// <summary>
        ///   Propを取得
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     可能ならrecycleを再利用したいため、Unityのメインスレッドから呼ぶ必要がある。
        ///   </para>
        /// </remarks>
        public Dictionary<string, object> GetProps(IDictionary<string, object> recycle = null)
        {
            if (props == null)
            {
                props = reader.ReadDict(recycle);
            }

            return props;
        }
    }
}
