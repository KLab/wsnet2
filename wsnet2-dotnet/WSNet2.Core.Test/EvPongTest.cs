using NUnit.Framework;
using System.Collections.Generic;

namespace WSNet2.Core.Test
{
    public class EvPongTest
    {
        [Test]
        public void TestEvPongPayload()
        {
            var payload = new byte[]{
                (byte)Type.ULong, 1,2,3,4,5,6,7,8,
                (byte)Type.UInt, 0,0,0,9,
                (byte)Type.Dict, 2,
                1, (byte)'a', 0, 9, (byte)Type.ULong, 2,3,4,5,6,7,8,9,
                2, (byte)'b', (byte)'b', 0, 9, (byte)Type.ULong, 3,4,5,6,7,8,9,10,
            };
            var explmts = new Dictionary<string, ulong>(){{"a", 0x0203040506070809}, {"bb", 0x030405060708090a}};

            var reader = WSNet2Serializer.NewReader(payload);
            var ev = new EvPong(reader);
            var lmts = new Dictionary<string, ulong>(){{"a", 1}, {"bb", 2}};
            ev.GetLastMsgTimestamps(lmts);

            Assert.AreEqual(0x0102030405060708, ev.PingTimestamp);
            Assert.AreEqual(9, ev.WatcherCount);
            Assert.AreEqual(explmts, lmts);
        }
    }
}
