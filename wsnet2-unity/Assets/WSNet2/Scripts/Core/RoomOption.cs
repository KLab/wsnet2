using System;
using System.Collections.Generic;
using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class RoomOption
    {
        [Key("visible")]
        public bool visible;

        [Key("watchable")]
        public bool watchable;

        [Key("with_number")]
        public bool withNumber;

        [Key("search_group")]
        public uint  searchGroup;

        [Key("client_deadline")]
        public uint clientDeadline;

        [Key("max_players")]
        public uint maxPlayers;

        [Key("public_props")]
        public byte[] publicProps;

        [Key("private_props")]
        public byte[] privateProps;

        public RoomOption()
        {
        }

        public RoomOption(
            uint maxPlayers,
            uint searchGroup,
            IDictionary<string, object> publicProps,
            IDictionary<string, object> privateProps)
        {
            this.visible = true;
            this.watchable = true;
            this.withNumber = false;
            this.searchGroup = searchGroup;
            this.maxPlayers = maxPlayers;

            var writer = Serialization.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(publicProps);
                this.publicProps = writer.ArraySegment().ToArray();

                writer.Reset();
                writer.Write(privateProps);
                this.privateProps = writer.ArraySegment().ToArray();
            }
        }

        public RoomOption Visible(bool val)
        {
            this.visible = val;
            return this;
        }

        public RoomOption Watchable(bool val)
        {
            this.watchable = val;
            return this;
        }

        public RoomOption WithNumber(bool val)
        {
            this.withNumber = val;
            return this;
        }

        public RoomOption WithClientDeadline(uint sec)
        {
            this.clientDeadline = sec;
            return this;
        }

        public override string ToString()
        {
            return string.Format(
                "RoomOption{{\r\n\tv:{0},w:{1},n:{2},sg:{3},mp:{4},\r\n\tpub:{5},\r\n\tpriv{6}}}",
                visible, watchable, withNumber, searchGroup, maxPlayers,
                BitConverter.ToString(publicProps), BitConverter.ToString(privateProps));
        }
    }

}
