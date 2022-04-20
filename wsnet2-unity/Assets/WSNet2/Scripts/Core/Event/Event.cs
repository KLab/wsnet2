using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   イベント種別
    /// </summary>
    public enum EvType
    {
        PeerReady = 1,
        Pong,

        Joined = EvTypeExt.regularEvType,
        Left,
        RoomProp,
        ClientProp,
        MasterSwitched,
        Message,
        Rejoined,

        Succeeded = EvTypeExt.responseEvType,
        PermissionDenied,
        TargetNotFound,

        Closed = EvTypeExt.localEvType,
    }

    static class EvTypeExt
    {
        public const int regularEvType = 30;
        public const int responseEvType = 128;
        public const int localEvType = 0x10000;

        public static bool IsRegular(this EvType type)
        {
            // responseもregularに含まれる(SequenceNumを持つ)
            return (int)type >= regularEvType && (int)type < localEvType;
        }
    }

    /// <summary>
    ///   Gameサーバから送られてくるイベント
    /// </summary>
    public class Event
    {
        /// <summary>
        ///   受信に使ったArraySegmentの中身（使い終わったらバッファプールに返却する用）
        /// </summary>
        public byte[] BufferArray { get; private set; }

        /// <summary>イベント種別</summary>
        public EvType Type { get; private set; }

        /// <summary>通常メッセージか</summary>
        public bool IsRegular { get => Type.IsRegular(); }

        /// <summary>通し番号</summary>
        public uint SequenceNum { get; private set; }

        protected SerialReader reader;

        /// <summary>
        ///   受信バイト列からEventを構築
        /// </summary>
        public static Event Parse(ArraySegment<byte> buf)
        {
            var reader = WSNet2Serializer.NewReader(buf);
            var type = (EvType)reader.Get8();

            Event ev;
            switch (type)
            {
                case EvType.PeerReady:
                    ev = new EvPeerReady(reader);
                    break;
                case EvType.Pong:
                    ev = new EvPong(reader);
                    break;

                case EvType.Joined:
                    ev = new EvJoined(reader);
                    break;
                case EvType.Left:
                    ev = new EvLeft(reader);
                    break;
                case EvType.RoomProp:
                    ev = new EvRoomProp(reader);
                    break;
                case EvType.ClientProp:
                    ev = new EvClientProp(reader);
                    break;
                case EvType.MasterSwitched:
                    ev = new EvMasterSwitched(reader);
                    break;
                case EvType.Message:
                    ev = new EvRPC(reader);
                    break;
                case EvType.Rejoined:
                    ev = new EvRejoined(reader);
                    break;

                case EvType.Succeeded:
                case EvType.PermissionDenied:
                case EvType.TargetNotFound:
                    ev = new EvResponse(type, reader);
                    break;

                default:
                    throw new Exception($"unknown event type: {type}");
            }

            ev.BufferArray = buf.Array;
            return ev;
        }

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        protected Event(EvType type, SerialReader reader)
        {
            this.Type = type;
            this.reader = reader;

            if (IsRegular)
            {
                SequenceNum = reader.Get32();
            }
        }
    }
}
