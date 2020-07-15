using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   送信メッセージを一時的に貯めるPool.
    /// </summary>
    /// <remarks>
    ///   <para>
    ///     再接続時に再送するため、一定数を保持する役目も持つ。
    ///   </para>
    /// </remarks>
    public class MsgPool
    {
        const int regularMsgType = 30;

        /// <summary>
        ///   Msg種別
        /// </summary>
        public enum MsgType
        {
            Ping = 1,

            Leave = regularMsgType,
            RoomProp,
            ClientProp,
            Target,
            ToMaster,
            Broadcast,
            Kick,
        }

        int sequenceNum;
        int tookSeqNum;
        SerialWriter[] pool;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="poolSize">保持できるMsg数</param>
        /// <param name="initialBufSize">各Msg(SerialWriter)の初期バッファサイズ</param>
        public MsgPool(int poolSize, int initialBufSize)
        {
            sequenceNum = 0;
            tookSeqNum = 0;
            pool = new SerialWriter[poolSize];
            for (var i=0; i<pool.Length; i++)
            {
                pool[i] = Serialization.NewWriter(initialBufSize);
            }
        }

        /// <summary>
        ///   送信するバイト列を取得
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     Takeで取得できたバッファは次にTakeを(より新しい番号で)呼ぶまで上書きされることはない。
        ///     基本的には番号順に呼ばれるが、再接続時に巻き戻る可能性がある。
        ///   </para>
        /// </remarks>
        /// <param name="seqNum">取得するMsgの通し番号</param>
        /// <return>
        ///   seqNum番目のMsgのバイト列。
        ///   まだ生成されていない番号のときはnull。
        ///   もうバッファから落ちた古い番号のときは例外を投げる。
        /// </return>
        public ArraySegment<byte>? Take(int seqNum)
        {
            lock(this)
            {
                if (sequenceNum - pool.Length >= seqNum)
                {
                    throw new Exception($"Msg too old: {seqNum}, {sequenceNum-pool.Length}");
                }
                if (seqNum > sequenceNum)
                {
                    return null;
                }

                tookSeqNum = seqNum;
                return pool[seqNum % pool.Length].ArraySegment();
            }
        }

        public void PostRPC(byte id, string param, string[] targets)
        {
            lock(this)
            {
                incrementSeqNum();
                var writer = pool[sequenceNum % pool.Length];
                writeRPCType(writer, id, targets);
                writer.Write(param);
            }
        }

        public void PostRPC<T>(byte id, T param, string[] targets) where T : class, IWSNetSerializable
        {
            lock(this)
            {
                incrementSeqNum();
                var writer = pool[sequenceNum % pool.Length];
                writeRPCType(writer, id, targets);
                writer.Write(param);
            }
        }

        private void writeRPCType(SerialWriter writer, byte id, string[] targets)
        {
            writer.Reset();

            if (targets == Room.RPCToMaster)
            {
                writer.Put8((int)MsgType.ToMaster);
                writer.Put24(sequenceNum);
            }
            else if (targets.Length == 0)
            {
                writer.Put8((int)MsgType.Broadcast);
                writer.Put24(sequenceNum);
            }
            else
            {
                writer.Put8((int)MsgType.Target);
                writer.Put24(sequenceNum);
                writer.Write(targets);
            }

            writer.Write(id);
        }

        /// <summary>
        ///   通し番号を進める.
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     直前にTakeされた場所を上書きいてしまう場合は例外を送出
        ///   </para>
        /// </remarks>
        void incrementSeqNum()
        {
            if (sequenceNum + 1 >= tookSeqNum + pool.Length)
            {
                throw new Exception($"MsgPool full: {tookSeqNum}..{sequenceNum}");
            }

            sequenceNum++;
        }
    }
}
