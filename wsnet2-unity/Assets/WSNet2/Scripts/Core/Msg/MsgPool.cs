using System;
using System.Collections;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Threading;

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
    class MsgPool
    {
        int sequenceNum;
        int tookSeqNum;
        SerialWriter[] pool;

        ///<summary>PoolにMsgが追加されたフラグ</summary>
        /// <remarks>
        ///   <para>
        ///     msgPoolにAdd*したあとTryAdd(true)する。
        ///     送信ループがTake()で待機しているので、Addされたら動き始める。
        ///     サイズ=1にしておくことで、送信前に複数回Addされても1度のループで送信される。
        ///   </para>
        /// </remarks>
        BlockingCollection<bool> hasMsg;

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

            hasMsg = new BlockingCollection<bool>(1);
        }

        /// <summary>
        ///   Msgが来るまで待つ
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     スレッドをブロックする
        ///   </para>
        /// </remarks>
        public void Wait(CancellationToken ct)
        {
            _ = hasMsg.Take(ct);
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

        /// <summary>
        ///   Leaveメッセージを投下
        /// </summary>
        public void PostLeave()
        {
            lock(this)
            {
                writeMsgType(MsgType.Leave);
            }
        }

        /// <summary>
        ///   RPCメッセージを投下
        /// </summary>
        public void PostRPC(byte id, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
            }
        }
        public void PostRPC(byte id, bool param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, sbyte param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, byte param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, short param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, ushort param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, int param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, uint param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, long param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, ulong param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, float param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, double param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, string param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC<T>(byte id, T param, string[] targets) where T : class, IWSNetSerializable
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, IEnumerable param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, IDictionary<string, object> param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, bool[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, sbyte[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, byte[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, short[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, ushort[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, int[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, uint[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, long[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, ulong[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, float[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }
        public void PostRPC(byte id, double[] param, string[] targets)
        {
            lock(this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
            }
        }

        private SerialWriter writeRPCHeader(byte id, string[] targets)
        {
            SerialWriter writer;

            if (targets == Room.RPCToMaster)
            {
                writer = writeMsgType(MsgType.ToMaster);
            }
            else if (targets.Length == 0)
            {
                writer = writeMsgType(MsgType.Broadcast);
            }
            else
            {
                writer = writeMsgType(MsgType.Target);
                writer.Write(targets);
            }

            writer.Write(id);
            return writer;
        }

        private SerialWriter writeMsgType(MsgType msgType)
        {
            incrementSeqNum();
            hasMsg.TryAdd(true);
            var writer = pool[sequenceNum % pool.Length];
            writer.Reset();
            writer.Put8((int)msgType);
            writer.Put24(sequenceNum);
            return writer;
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
