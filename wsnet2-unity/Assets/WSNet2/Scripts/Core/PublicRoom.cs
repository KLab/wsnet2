using System;
using System.Collections.Generic;

namespace WSNet2.Core
{
    /// <summary>
    ///   検索結果で得られる公開部屋情報
    /// </summary>
    public class PublicRoom
    {
        /// <summary>RoomID</summary>
        public string Id => info.id;

        /// <summary>検索可能</summary>
        public bool Visible => info.visible;

        /// <summary>入室可能</summary>
        public bool Joinable => info.joinable;

        /// <summary>観戦可能</summary>
        public bool Watchable => info.watchable;

        /// <summary>部屋番号</summary>
        public int Number => info.number;

        /// <summary>検索グループ</summary>
        public uint SearchGroup => info.searchGroup;

        /// <summary>最大人数</summary>
        public uint MaxPlayers => info.maxPlayers;

        /// <summary>プレイヤー人数</summary>
        public uint PlayerCount => info.players;

        /// <summary>観戦人数</summary>
        public uint WatcherCount => info.watchers;

        /// <summary>ルームの公開プロパティ</summary>
        public IReadOnlyDictionary<string, object> PublicProps => publicProps;

        public DateTime Created { get; private set; }

        protected RoomInfo info;
        protected Dictionary<string, object> publicProps;

        public PublicRoom(RoomInfo roomInfo)
        {
            info = roomInfo;

            var reader = Serialization.NewReader(roomInfo.publicProps);
            publicProps = reader.ReadDict();

            Created = DateTimeOffset.FromUnixTimeSeconds(roomInfo.created).DateTime;
        }
    }
}
