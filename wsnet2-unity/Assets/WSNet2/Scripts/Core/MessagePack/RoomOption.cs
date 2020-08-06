using System;
using System.Collections.Generic;
using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class RoomOption
    {
        public enum LogLevel
        {
            DEFAULT = 0,
            NOLOG,
            ERROR,
            INFO,
            DEBUG,
            ALL,
        }

        [Key("visible")]
        public bool visible;

        [Key("joinable")]
        public bool joinable;

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

        [Key("log_level")]
        public LogLevel logLevel;

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
            this.joinable = true;
            this.watchable = true;
            this.withNumber = false;
            this.searchGroup = searchGroup;
            this.maxPlayers = maxPlayers;

            var writer = Serialization.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(publicProps);
                var arr = (IList<byte>)writer.ArraySegment();
                this.publicProps = new byte[arr.Count];
                arr.CopyTo(this.publicProps, 0);

                writer.Reset();
                writer.Write(privateProps);
                arr = (IList<byte>)writer.ArraySegment();
                this.privateProps = new byte[arr.Count];
                arr.CopyTo(this.privateProps, 0);
            }
        }

        public RoomOption Visible(bool val)
        {
            this.visible = val;
            return this;
        }

        public RoomOption Joinable(bool val)
        {
            this.joinable = val;
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

        public RoomOption SetLogLevel(LogLevel l)
        {
            this.logLevel = l;
            return this;
        }

        public override string ToString()
        {
            return string.Format(
                "RoomOption{{\r\n\tv:{0},w:{1},n:{2},sg:{3},mp:{4},\r\n\tpub:{5},\r\n\tpriv{6}}}",
                visible, joinable, watchable, withNumber, searchGroup, maxPlayers,
                BitConverter.ToString(publicProps), BitConverter.ToString(privateProps));
        }
    }

}
