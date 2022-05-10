using MessagePack;

namespace WSNet2
{
    [MessagePackObject]
    public class RoomInfo
    {
        [Key("id")]
        public string id;

        [Key("app_id")]
        public string appId;

        [Key("host_id")]
        public uint hostId;

        [Key("visible")]
        public bool visible;

        [Key("joinable")]
        public bool joinable;

        [Key("watchable")]
        public bool watchable;

        [Key("number")]
        public int number;

        [Key("search_group")]
        public uint searchGroup;

        [Key("max_players")]
        public uint maxPlayers;

        [Key("players")]
        public uint players;

        [Key("watchers")]
        public uint watchers;

        [Key("public_props")]
        public byte[] publicProps;

        [Key("private_props")]
        public byte[] privateProps;

        [Key("created")]
        public long created;
    }
}
