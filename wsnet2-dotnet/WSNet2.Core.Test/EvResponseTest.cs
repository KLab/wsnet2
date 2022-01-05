using NUnit.Framework;
using System;
using System.Collections.Generic;
using System.Security.Cryptography;

namespace WSNet2.Core.Test
{
    public class EvResponsePayloadTests
    {
        MsgPool msgpool;

        [OneTimeSetUp]
        public void OneTimeSetup()
        {
            msgpool = new MsgPool(2, 128, new HMACSHA1(new byte[]{0}));
        }

        [Test]
        public void TestRoomPropPayload()
        {
            testRoomPropPayload(true, false, false, 10, 20, 30, null, null);
            testRoomPropPayload(
                false, true, false, 11, 21, 31,
                new Dictionary<string, object>(){{"k1", 100}}, null);
            testRoomPropPayload(
                false, false, true, 12, 22, 32,
                null, new Dictionary<string, object>(){{"k2", new int[]{1,2,3}}});
        }

        void testRoomPropPayload(
            bool visible, bool joinable, bool watchable,
            uint searchGroup, ushort maxPlayers, ushort clientDeadline,
            Dictionary<string, object> publicProps,
            Dictionary<string, object> privateProps)
        {
            var seqnum = msgpool.PostRoomProp(
                visible, joinable, watchable,
                searchGroup, maxPlayers, clientDeadline,
                publicProps, privateProps);

            var msg = msgpool.Take(seqnum).Value;
            var buf = new byte[3 + msg.Count];
            msg.CopyTo(buf, 3);
            var reader = Serialization.NewReader(new ArraySegment<byte>(buf));
            var ev = new EvResponse(EvType.PermissionDenied, reader);
            var payload = ev.GetRoomPropPayload();

            Assert.AreEqual(visible, payload.Visible);
            Assert.AreEqual(joinable, payload.Joinable);
            Assert.AreEqual(watchable, payload.Watchable);
            Assert.AreEqual(searchGroup, payload.SearchGroup);
            Assert.AreEqual(maxPlayers, payload.MaxPlayers);
            Assert.AreEqual(clientDeadline, payload.ClientDeadline);
            Assert.AreEqual(publicProps, payload.PublicProps);
            Assert.AreEqual(privateProps, payload.PrivateProps);
        }
    }
}
