using System;
using System.Collections.Generic;

namespace WSNet2
{
    /// <summary>
    ///   Msgのレスポンス
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     base.Typeで種別を判定
    ///     - Succeeded
    ///     - PermissionDenied
    ///     - TargetNotFound
    ///   </para>
    ///   <para>
    ///     レスポンスがあるのは次のMsg
    ///     - RoomProp
    ///     - ClientProp
    ///     - SwitchMaster
    ///     - Kick
    ///   </para>
    /// </remarks>
    public class EvResponse : Event
    {
        public struct RoomPropPayload
        {
            public bool Visible;
            public bool Joinable;
            public bool Watchable;
            public uint SearchGroup;
            public ushort MaxPlayers;
            public ushort ClientDeadline;
            public Dictionary<string, object> PublicProps;
            public Dictionary<string, object> PrivateProps;

            public RoomPropPayload(SerialReader reader)
            {
                var flags = reader.ReadByte();
                Visible = (flags & 1) != 0;
                Joinable = (flags & 2) != 0;
                Watchable = (flags & 4) != 0;
                SearchGroup = reader.ReadUInt();
                MaxPlayers = reader.ReadUShort();
                ClientDeadline = reader.ReadUShort();
                PublicProps = reader.ReadDict();
                PrivateProps = reader.ReadDict();
            }
        }

        /// <summary>
        ///   元となるMsgのシーケンス番号
        /// </summary>
        public int MsgSeqNum { get; private set; }

        /// <summary>
        ///   TargetNotFoundのとき不在だったTarget
        /// </summary>
        public string[] Targets { get; private set; }

        /// <summary>
        ///   元となるMsgの内容（Succeededでは空）
        /// </summary>
        public ArraySegment<byte> Payload { get; private set; }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        public EvResponse(EvType type, SerialReader reader) : base(type, reader)
        {
            MsgSeqNum = reader.Get24();

            if (type == EvType.TargetNotFound)
            {
                Targets = reader.ReadStrings();
            }

            if (type != EvType.Succeeded)
            {
                Payload = reader.GetRest();
            }
        }

        public RoomPropPayload GetRoomPropPayload()
        {
            var reader = WSNet2Serializer.NewReader(Payload);
            return new RoomPropPayload(reader);
        }

        public Dictionary<string, object> GetClientPropPayload()
        {
            var reader = WSNet2Serializer.NewReader(Payload);
            return reader.ReadDict();
        }

        public string GetSwitchMasterPayload()
        {
            var reader = WSNet2Serializer.NewReader(Payload);
            return reader.ReadString();
        }

        public string GetKickPayload()
        {
            var reader = WSNet2Serializer.NewReader(Payload);
            return reader.ReadString();
        }
    }
}
