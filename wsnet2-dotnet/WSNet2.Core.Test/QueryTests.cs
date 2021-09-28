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
        public void TestAnd()
        {
            var query = new Query();

            // ()*() = ()
            query.And();
            var expect = new object[]{new object[]{}};
            Assert.AreEqual(
                MessagePackSerializer.SerializeToJson(expect),
                MessagePackSerializer.SerializeToJson(query.condsList));

            // ()* ((A+B)) = (A+B)
            query.And(
                new Query().Or(
                    new Query().Equal("A", 1),
                    new Query().Equal("B", 2)));
            expect = new object[]
            {
                new object[]{
                    new object[]{"A", (byte)Query.Op.Equal, serialize(1)}
                },
                new object[]{
                    new object[]{"B", (byte)Query.Op.Equal, serialize(2)}
                },
            };
            Assert.AreEqual(
                MessagePackSerializer.SerializeToJson(expect),
                MessagePackSerializer.SerializeToJson(query.condsList));

            // (A+B)*((C+D)*(E+F)) = (ACE+ACF+ADE+ADF+BCE+BCF+BDE+BDF)
            query.And(
                new Query().Or(new Query().Equal("C", 3), new Query().Equal("D", 4)),
                new Query().Or(new Query().Equal("E", 5), new Query().Equal("F", 6)));
            expect = new object[]
            {
                new object[]{
                    new object[]{"A", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"C", (byte)Query.Op.Equal, serialize(3)},
                    new object[]{"E", (byte)Query.Op.Equal, serialize(5)},
                },
                new object[]{
                    new object[]{"A", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"C", (byte)Query.Op.Equal, serialize(3)},
                    new object[]{"F", (byte)Query.Op.Equal, serialize(6)},
                },
                new object[]{
                    new object[]{"A", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"D", (byte)Query.Op.Equal, serialize(4)},
                    new object[]{"E", (byte)Query.Op.Equal, serialize(5)},
                },
                new object[]{
                    new object[]{"A", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"D", (byte)Query.Op.Equal, serialize(4)},
                    new object[]{"F", (byte)Query.Op.Equal, serialize(6)},
                },
                new object[]{
                    new object[]{"B", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"C", (byte)Query.Op.Equal, serialize(3)},
                    new object[]{"E", (byte)Query.Op.Equal, serialize(5)},
                },
                new object[]{
                    new object[]{"B", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"C", (byte)Query.Op.Equal, serialize(3)},
                    new object[]{"F", (byte)Query.Op.Equal, serialize(6)},
                },
                new object[]{
                    new object[]{"B", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"D", (byte)Query.Op.Equal, serialize(4)},
                    new object[]{"E", (byte)Query.Op.Equal, serialize(5)},
                },
                new object[]{
                    new object[]{"B", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"D", (byte)Query.Op.Equal, serialize(4)},
                    new object[]{"F", (byte)Query.Op.Equal, serialize(6)},
                },
            };
            Assert.AreEqual(
                MessagePackSerializer.SerializeToJson(expect),
                MessagePackSerializer.SerializeToJson(query.condsList));
        }

        [Test]
        public void TestOr()
        {
            var query = new Query();

            // ()*() = ()
            query.Or();
            var expect = new object[]{new object[]{}};
            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect),
                MessagePackSerializer.Serialize(query.condsList));

            // ()*(A+B) = (A+B)
            query.Or(
                new Query().Equal("k1", 1),
                new Query().Equal("k1", 2));
            expect = new object[]
            {
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)}
                },
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(2)}
                },
            };
            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect),
                MessagePackSerializer.Serialize(query.condsList));

            // (A+B)*() = (A+B)
            query.Or(new Query());
            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect),
                MessagePackSerializer.Serialize(query.condsList));

            // (A+B)*((C+D)+(E)) = (AC+AD+AE+BC+BD+BE)
            var q2 = new Query().Or(
                new Query().Equal("k2", 1),
                new Query().Equal("k2", 2));
            query.Or(q2, new Query().Equal("k3", 3));
            expect = new object[]
            {
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize(1)},
                },
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize(2)},
                },
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(3)},
                },
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize(1)},
                },
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"k2", (byte)Query.Op.Equal, serialize(2)},
                },
                new object[]{
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(2)},
                    new object[]{"k3", (byte)Query.Op.Equal, serialize(3)},
                },
            };
            query.Or(new Query());
            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect),
                MessagePackSerializer.Serialize(query.condsList));
        }

        [Test]
        public void TestInt()
        {
            var query = new Query();
            query.Equal("k1", 1);
            query.Not("k2", 2);
            query.LessThan("k3", 3);
            query.LessEqual("k4", 4);
            query.GreaterThan("k5", 5);
            query.GreaterEqual("k6", 6);
            query.Between("k7", 17, 27);

            var expect = new object[]
            {
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Equal, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.Not, serialize(2)},
                    new object[]{"k3", (byte)Query.Op.LessThan, serialize(3)},
                    new object[]{"k4", (byte)Query.Op.LessEqual, serialize(4)},
                    new object[]{"k5", (byte)Query.Op.GreaterThan, serialize(5)},
                    new object[]{"k6", (byte)Query.Op.GreaterEqual, serialize(6)},
                    new object[]{"k7", (byte)Query.Op.GreaterEqual, serialize(17)},
                    new object[]{"k7", (byte)Query.Op.LessEqual, serialize(27)},
                }
            };

            Assert.AreEqual(
                MessagePackSerializer.Serialize(expect),
                MessagePackSerializer.Serialize(query.condsList));
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
        public void TestContain()
        {
            var query = new Query();
            query.Contain("k1", 1);
            query.NotContain("k2", "test");

            var expect = new object[]
            {
                new object[]
                {
                    new object[]{"k1", (byte)Query.Op.Contain, serialize(1)},
                    new object[]{"k2", (byte)Query.Op.NotContain, serialize("test")},
                },
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
