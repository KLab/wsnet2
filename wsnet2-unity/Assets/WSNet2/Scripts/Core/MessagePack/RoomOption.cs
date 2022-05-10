using System;
using System.Collections.Generic;
using MessagePack;

namespace WSNet2
{
    [MessagePackObject]
    public class RoomOption
    {
        /// <summary>
        ///   部屋のログレベル
        /// </summary>
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
        public uint searchGroup;

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

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="maxPlayers">最大プレイヤー数</param>
        /// <param name="searchGroup">検索グループ</param>
        /// <param name="publicProps">公開プロパティ</param>
        /// <param name="privateProps">非公開プロパティ</param>
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

            var writer = WSNet2Serializer.GetWriter();
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

        /// <summary>
        ///   検索可能フラグを設定する
        /// </summary>
        /// <remarks>
        ///   デフォルトtrue
        /// </remarks>
        public RoomOption Visible(bool val)
        {
            this.visible = val;
            return this;
        }

        /// <summary>
        ///   入室可能フラグを設定する
        /// </summary>
        /// <remarks>
        ///   デフォルトtrue
        /// </remarks>
        public RoomOption Joinable(bool val)
        {
            this.joinable = val;
            return this;
        }

        /// <summary>
        ///   観戦可能フラグを設定する
        /// </summary>
        /// <remarks>
        ///   デフォルトtrue
        /// </remarks>
        public RoomOption Watchable(bool val)
        {
            this.watchable = val;
            return this;
        }

        /// <summary>
        ///   部屋番号の割り当て設定
        /// </summary>
        /// <remarks>
        ///   デフォルトfalse
        /// </remarks>
        public RoomOption WithNumber(bool val)
        {
            this.withNumber = val;
            return this;
        }

        /// <summary>
        ///   ClientDeadlineを設定する
        /// </summary>
        /// <param name="sec">設定値（秒）</param>
        /// <remarks>
        ///   デフォルト値はサーバ側の設定による
        /// </remarks>
        public RoomOption WithClientDeadline(uint sec)
        {
            this.clientDeadline = sec;
            return this;
        }

        /// <summary>
        ///   部屋のLogLevelを設定する
        /// </summary>
        /// <param name="l">ログレベル</param>
        /// <remarks>
        ///   デフォルト値はサーバ側の設定による
        /// </remarks>
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
