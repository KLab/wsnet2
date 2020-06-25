using System;

namespace WSNet2.Core
{
    public class MsgPool
    {
        const int regularMsgType = 30;

        public enum MsgType
        {
            Ping = 1,

            Leave = regularMsgType,
            RoomProp,
            ClientProp,
            Target,
            Broadcast,
            Kick,
        }

        int sequenceNum;
        int tookSeqNum;
        SerialWriter[] pool;

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

        public ArraySegment<byte> Take(int seqNum)
        {
            lock(this)
            {
                if (sequenceNum - pool.Length >= seqNum)
                {
                    throw new Exception($"Msg tool old: {seqNum}, {sequenceNum-pool.Length}");
                }
                if (seqNum > sequenceNum)
                {
                    return null;
                }

                tookSeqNum = seqNum;
                return pool[seqNum % pool.Length].ArraySegment();
            }
        }

        public void AddBroadcast(IWSNetSerializable data)
        {
            lock(this)
            {
                incrementSeqNum();
                var writer = pool[sequenceNum % pool.Length];
                writer.Reset();
                writer.Put8((int)MsgType.Broadcast);
                writer.Put24(sequenceNum);
                writer.Write(data);
            }
        }


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
