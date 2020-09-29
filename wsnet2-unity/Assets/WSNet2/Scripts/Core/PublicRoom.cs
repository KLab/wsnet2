using System;
using System.Collections.Generic;

namespace WSNet2.Core
{
    /// <summary>
    ///   検索結果で得られる公開部屋情報
    /// </summary>
    public class PublicRoom
    {
        RoomInfo roomInfo;

        public string Id => roomInfo.id;

        public bool Visible => roomInfo.visible;

        public bool Joinable => roomInfo.joinable;

        public bool Watchable => roomInfo.watchable;

        public int Number => roomInfo.number;

        public uint SearchGroup => roomInfo.searchGroup;

        public uint MaxPlayers => roomInfo.maxPlayers;

        public uint Players => roomInfo.players;

        public uint Watchers => roomInfo.watchers;

        public IReadOnlyDictionary<string, object> PublicProps { get; private set; }

        public DateTime Created { get; private set; }

        public PublicRoom(RoomInfo roomInfo)
        {
            this.roomInfo = roomInfo;

            var reader = Serialization.NewReader(roomInfo.publicProps);
            this.PublicProps = reader.ReadDict();

            this.Created = DateTimeOffset.FromUnixTimeSeconds(roomInfo.created).DateTime;
        }
    }
}
