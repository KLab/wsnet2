using System;
using System.Collections.Generic;

namespace WSNet2
{
    /// <summary>
    ///   Room内にいるPlayer
    /// </summary>
    public class Player
    {
        /// <summary>ID</summary>
        public string Id { get; private set; }

        /// <summary>カスタムプロパティ</summary>
        /// <remarks>
        ///   <para>
        ///     値はシリアライズ可能なものに限る
        ///   </para>
        /// </remarks>
        public Dictionary<string, object> Props;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public Player(ClientInfo info)
        {
            Id = info.Id;
            var reader = WSNet2Serializer.NewReader(info.Props);
            Props = reader.ReadDict();
        }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public Player(string id, Dictionary<string, object> props)
        {
            Id = id;
            Props = props;
        }
    }
}
