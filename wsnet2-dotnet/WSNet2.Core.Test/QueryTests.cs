using NUnit.Framework;
using MessagePack;

namespace WSNet2.Core.Test
{
    public class QueryTests
    {
        SerialWriter writer;

        [OneTimeSetUp]
        public void Setup()
        {
            writer = Serialization.NewWriter();
        }

        [Test]
        public void TestString()
        {
            var query = new Query();

            query.Equal("k1", "equal");
            query.Not("k2", "not");

            var expect = new object[]
            {
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize("equal")},
                    new object[]{"k2", (byte)Query.Op.Not, serialize("not")},
                }
            };

            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect), 
                MessagePackSerializer.Serialize(query.condsList));
        }

        [Test]
        public void TestSample()
        {
            var query = new Query();

            query.Equal("k1", 1);
            query.Equal("k2", "test");
            query.Or(
                new Query().Equal("o1", 10),
                new Query().Equal("o1", 20),
                new Query().Equal("o1", 30));

            var expect = new object[]
            {
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(10)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(20)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(30)},
                },
            };

            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect), 
                MessagePackSerializer.Serialize(query.condsList));

            query.Equal("k3", 100);
            query.Or(
                new Query().Equal("o2", 1).Equal("o3", 1),
                new Query().Equal("o2", 2).Equal("o3", 2));

            expect = new object[]
            {
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(10)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(100)},
                    new object[]{"o2", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"o3", (byte)Query.Op.Equal, serialize(1)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(10)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(100)},
                    new object[]{"o2", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"o3", (byte)Query.Op.Equal, serialize(2)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(20)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(100)},
                    new object[]{"o2", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"o3", (byte)Query.Op.Equal, serialize(1)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(20)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(100)},
                    new object[]{"o2", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"o3", (byte)Query.Op.Equal, serialize(2)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(30)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(100)},
                    new object[]{"o2", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"o3", (byte)Query.Op.Equal, serialize(1)},
                },
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize("test")},
                    new object[]{"o1", (byte)Query.Op.Equal, serialize(30)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(100)},
                    new object[]{"o2", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"o3", (byte)Query.Op.Equal, serialize(2)},
                },
            };

            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect), 
                MessagePackSerializer.Serialize(query.condsList));
        }

        private byte[] serialize(int val)
        {
            lock(writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ArraySegment().ToArray();
            }
        }

        private byte[] serialize(string val)
        {
            lock(writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ArraySegment().ToArray();
            }
        }
    }
}
