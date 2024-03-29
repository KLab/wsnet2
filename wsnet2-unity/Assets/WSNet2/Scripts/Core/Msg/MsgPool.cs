﻿using System;
using System.Collections;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Security.Cryptography;
using System.Threading;

namespace WSNet2
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

        HMAC hmac;

        /// <summary>
        ///   コンストラクタ
        /// </summary>
        /// <param name="poolSize">保持できるMsg数</param>
        /// <param name="initialBufSize">各Msg(SerialWriter)の初期バッファサイズ</param>
        public MsgPool(int poolSize, int initialBufSize, HMAC hmac)
        {
            sequenceNum = 0;
            tookSeqNum = 0;
            pool = new SerialWriter[poolSize];
            for (var i = 0; i < pool.Length; i++)
            {
                pool[i] = WSNet2Serializer.NewWriter(initialBufSize);
            }

            hasMsg = new BlockingCollection<bool>(1);

            this.hmac = hmac;
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
            lock (this)
            {
                if (sequenceNum - pool.Length >= seqNum)
                {
                    throw new Exception($"Msg too old: {seqNum}, {sequenceNum - pool.Length}");
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
        public int PostLeave(string message)
        {
            lock (this)
            {
                var writer = writeMsgType(MsgType.Leave);
                writer.Write(message);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }

        /// <summary>
        ///   Master移譲メッセージを投下
        /// </summary>
        public int PostSwitchMaster(string newMasterId)
        {
            lock (this)
            {
                var writer = writeMsgType(MsgType.SwitchMaster);
                writer.Write(newMasterId);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }

        /// <summary>
        ///   RoomPorps変更メッセージを投下
        /// </summary>
        public int PostRoomProp(
            bool visible, bool joinable, bool watchable,
            uint searchGroup,
            ushort maxPlayers,
            ushort clientDeadline,
            IDictionary<string, object> publicProps,
            IDictionary<string, object> privateProps)
        {
            lock (this)
            {
                var flags = (byte)((visible ? 1 : 0) + (joinable ? 2 : 0) + (watchable ? 4 : 0));

                var writer = writeMsgType(MsgType.RoomProp);
                writer.Write(flags);
                writer.Write(searchGroup);
                writer.Write(maxPlayers);
                writer.Write(clientDeadline);
                writer.Write(publicProps);
                writer.Write(privateProps);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }

        /// <summary>
        ///   自分自身のプロパティ変更メッセージを投下
        /// </summary>
        public int PostClientProp(IDictionary<string, object> props)
        {
            lock (this)
            {
                var writer = writeMsgType(MsgType.ClientProp);
                writer.Write(props);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }

        /// <summary>
        ///   強制退室メッセージを投下
        /// </summary>
        public int PostKick(string targetId, string message)
        {
            lock (this)
            {
                var writer = writeMsgType(MsgType.Kick);
                writer.Write(targetId);
                writer.Write(message);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }

        /// <summary>
        ///   RPCメッセージを投下
        /// </summary>
        public int PostRPC(byte id, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, bool param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, sbyte param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, byte param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, char param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, short param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, ushort param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, int param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, uint param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, long param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, ulong param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, float param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, double param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, string param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC<T>(byte id, T param, string[] targets) where T : class, IWSNet2Serializable
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, IEnumerable param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, IDictionary<string, object> param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, bool[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, sbyte[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, byte[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, char[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, short[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, ushort[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, int[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, uint[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, long[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, ulong[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, float[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
            }
        }
        public int PostRPC(byte id, double[] param, string[] targets)
        {
            lock (this)
            {
                var writer = writeRPCHeader(id, targets);
                writer.Write(param);
                writer.AppendHMAC(hmac);
                return sequenceNum;
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

#if DEBUG
        /// <summary>
        ///   PayloadからMsg情報を取り出す（NetworkInformer用）
        /// </summary>
        public static NetworkInformer.RoomSendInfo ParsePayload(Room room, ArraySegment<byte> payload)
        {
            var bodysize = payload.Count;
            var reader = WSNet2Serializer.NewReader(payload);
            var msgType = (MsgType)reader.Get8();
            switch (msgType)
            {
                case MsgType.Ping:
                    return new NetworkInformer.RoomSendPingInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = 0,
                        TimestampMilliSec = reader.Get64(),
                    };
                case MsgType.Leave:
                    return new NetworkInformer.RoomSendLeaveInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = reader.Get24(),
                    };
                case MsgType.RoomProp:
                    var seqnum = reader.Get24();
                    var flags = reader.ReadByte();
                    return new NetworkInformer.RoomSendRoomPropInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = seqnum,
                        Visible = (flags & 1) != 0,
                        Joinable = (flags & 2) != 0,
                        Watchable = (flags & 4) != 0,
                        SearchGroup = reader.ReadUInt(),
                        MaxPlayers = reader.ReadUShort(),
                        ClientDeadline = reader.ReadUShort(),
                        PublicProps = NetworkInformer.CutOutOne(reader),
                        PrivateProps = NetworkInformer.CutOutOne(reader),
                    };
                case MsgType.ClientProp:
                    return new NetworkInformer.RoomSendPlayerPropInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = reader.Get24(),
                        Props = NetworkInformer.CutOutOne(reader),
                    };
                case MsgType.SwitchMaster:
                    return new NetworkInformer.RoomSendSwitchMasterInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = reader.Get24(),
                        NewMaster = reader.ReadString(),
                    };
                case MsgType.Target:
                    seqnum = reader.Get24();
                    var targets = reader.ReadStrings();
                    var rpcId = reader.ReadByte();
                    return new NetworkInformer.RoomSendRPCInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = seqnum,
                        RpcID = rpcId,
                        MethodName = room.MethodName(rpcId),
                        Targets = targets,
                        Param = NetworkInformer.CutOutOne(reader),
                    };
                case MsgType.ToMaster:
                    seqnum = reader.Get24();
                    rpcId = reader.ReadByte();
                    return new NetworkInformer.RoomSendRPCInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = seqnum,
                        RpcID = rpcId,
                        MethodName = room.MethodName(rpcId),
                        Targets = Room.RPCToMaster,
                        Param = NetworkInformer.CutOutOne(reader),
                    };
                case MsgType.Broadcast:
                    seqnum = reader.Get24();
                    rpcId = reader.ReadByte();
                    return new NetworkInformer.RoomSendRPCInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = seqnum,
                        RpcID = rpcId,
                        MethodName = room.MethodName(rpcId),
                        Targets = new string[0],
                        Param = NetworkInformer.CutOutOne(reader),
                    };
                case MsgType.Kick:
                    return new NetworkInformer.RoomSendKickInfo()
                    {
                        BodySize = bodysize,
                        RoomID = room.Id,
                        MsgType = msgType,
                        SequenceNum = reader.Get24(),
                        Target = reader.ReadString(),
                    };
                default:
                    throw new Exception($"Unknown MsgType {msgType}");
            }
        }
#endif
    }
}
