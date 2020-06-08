using NUnit.Framework;
using System;

namespace WSNet2.Core.Test
{
    public class SerializeTests
    {
        SerialWriter writer;

        [SetUp]
        public void Setup()
        {
            writer = new SerialWriter();
        }

        [Test]
        public void TestNull()
        {
            writer.Reset();
            var expect = new byte[]{(byte)Type.Null};
            writer.Write();
            Assert.AreEqual(writer.ArraySegment(), expect);
        }

        [TestCase(true, new byte[]{(byte)Type.True})]
        [TestCase(false, new byte[]{(byte)Type.False})]
        public void TestBool(bool b, byte[] expect)
        {
            writer.Reset();
            writer.Write(b);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadBool();
            Assert.AreEqual(r, b);
        }

        [TestCase(sbyte.MinValue, new byte[]{(byte)Type.SByte, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.SByte, 0x80})]
        [TestCase(sbyte.MaxValue, new byte[]{(byte)Type.SByte, 0xff})]
        public void TestSByte(sbyte v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadSByte();
            Assert.AreEqual(r, v);
        }

        [TestCase(byte.MinValue, new byte[]{(byte)Type.Byte, 0x00})]
        [TestCase(byte.MaxValue, new byte[]{(byte)Type.Byte, 0xff})]
        public void TestByte(byte v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadByte();
            Assert.AreEqual(r, v);
        }

        [TestCase(short.MinValue, new byte[]{(byte)Type.Short, 0x00, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.Short, 0x80, 0x00})]
        [TestCase(short.MaxValue, new byte[]{(byte)Type.Short, 0xff, 0xff})]
        public void TestShort(short v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadShort();
            Assert.AreEqual(r, v);
        }

        [TestCase(ushort.MinValue, new byte[]{(byte)Type.UShort, 0x00, 0x00})]
        [TestCase(ushort.MaxValue, new byte[]{(byte)Type.UShort, 0xff, 0xff})]
        public void TestUShort(ushort v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadUShort();
            Assert.AreEqual(r, v);
        }

        [TestCase(int.MinValue, new byte[]{(byte)Type.Int, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.Int, 0x80, 0x00, 0x00, 0x00})]
        [TestCase(int.MaxValue, new byte[]{(byte)Type.Int, 0xff, 0xff, 0xff, 0xff})]
        public void TestInt(int v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadInt();
            Assert.AreEqual(r, v);
        }

        [TestCase(uint.MinValue, new byte[]{(byte)Type.UInt, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(uint.MaxValue, new byte[]{(byte)Type.UInt, 0xff, 0xff, 0xff, 0xff})]
        public void TestUInt(uint v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadUInt();
            Assert.AreEqual(r, v);
        }

        [TestCase(long.MinValue, new byte[]{(byte)Type.Long, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.Long, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(long.MaxValue, new byte[]{(byte)Type.Long, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestLong(long v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadLong();
            Assert.AreEqual(r, v);
        }

        [TestCase(ulong.MinValue, new byte[]{(byte)Type.ULong, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(ulong.MaxValue, new byte[]{(byte)Type.ULong, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestULong(ulong v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(writer.ArraySegment(), expect);

            var reader = new SerialReader(writer.ArraySegment());
            var r = reader.ReadULong();
            Assert.AreEqual(r, v);
        }

    }
}
