using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   イベント種別
    /// </summary>
    public enum EvType
    {
        regularEvType = 30,
        localEvType = 0x10000,

        PeerReady = 1,
        Pong,

        Joined = regularEvType,
        Left,
        RomProp,
        ClientProp,
        Message,

        Closed = localEvType,
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
        public bool IsRegular { get{ return Type >= EvType.regularEvType && Type < EvType.localEvType; } }

        /// <summary>通し番号</summary>
        public uint SequenceNum { get; private set; }

        protected SerialReader reader;

        /// <summary>
        ///   受信バイト列からEventを構築
        /// </summary>
        public static Event Parse(ArraySegment<byte> buf)
        {
            var reader = Serialization.NewReader(buf);
            var type = (EvType)reader.Get8();

            Event ev;
            switch (type)
            {
                case EvType.PeerReady:
                    ev = new EvPeerReady(reader);
                    break;
                case EvType.Joined:
                    ev = new EvJoined(reader);
                    break;
                case EvType.Left:
                    ev = new EvLeft(reader);
                    break;
                case EvType.Message:
                    ev = new EvRPC(reader);
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
