using NUnit.Framework;
using System;

using System.Collections.Generic;
using System.Collections;

namespace WSNet2.Core.Test
{
    class Obj1 : IWSNetSerializable, IEquatable<Obj1>
    {
        public int Num;
        public string Str;

        public Obj1()
        {
            Console.WriteLine("new Obj1!!");
        }

        public Obj1(int n, string s)
        {
            Num = n;
            Str = s;
        }

        public void Serialize(SerialWriter writer)
        {
            writer.Write(Num);
            writer.Write(Str);
        }

        public void Deserialize(SerialReader reader, int size)
        {
            Num = reader.ReadInt();
            Str = reader.ReadString();
        }

        public bool Equals(Obj1 o)
        {
            return Num == o.Num && Str == o.Str;
        }

        public override string ToString()
        {
            return string.Format("Obj1:{0},\"{1}\"", Num, Str);
        }
    }

    class Obj2 : IWSNetSerializable, IEquatable<Obj2>
    {
        public short S;
        public Obj1 Obj;

        public Obj2()
        {
        }

        public Obj2(short s, Obj1 obj)
        {
            S = s;
            Obj = obj;
        }

        public void Serialize(SerialWriter writer)
        {
            writer.Write(S);
            writer.Write(Obj);
        }

        public void Deserialize(SerialReader reader, int size)
        {
            S = reader.ReadShort();
            Obj = reader.ReadObject(Obj);
        }

        public bool Equals(Obj2 o)
        {
            return S==o.S && Obj.Equals(o.Obj);
        }

        public override string ToString()
        {
            return string.Format("Obj2:{0},<{1}>", S, Obj);
        }
    }


    public class SerializeTests
    {
        SerialWriter writer;

        [OneTimeSetUp]
        public void OnTimeSetUp()
        {
            Serialization.Register<Obj1>((byte)'A');
            Serialization.Register<Obj2>((byte)'B');
        }

        [SetUp]
        public void Setup()
        {
            writer = Serialization.NewWriter();
        }

        [Test]
        public void TestNull()
        {
            writer.Reset();
            var expect = new byte[]{(byte)Type.Null};
            writer.Write();
            Assert.AreEqual(expect, writer.ArraySegment());
        }

        [TestCase(true, new byte[]{(byte)Type.True})]
        [TestCase(false, new byte[]{(byte)Type.False})]
        public void TestBool(bool b, byte[] expect)
        {
            writer.Write(b);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadBool();
            Assert.AreEqual(b, r);
        }

        [TestCase(sbyte.MinValue, new byte[]{(byte)Type.SByte, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.SByte, 0x80})]
        [TestCase(sbyte.MaxValue, new byte[]{(byte)Type.SByte, 0xff})]
        public void TestSByte(sbyte v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadSByte();
            Assert.AreEqual(v, r);
        }

        [TestCase(byte.MinValue, new byte[]{(byte)Type.Byte, 0x00})]
        [TestCase(byte.MaxValue, new byte[]{(byte)Type.Byte, 0xff})]
        public void TestByte(byte v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadByte();
            Assert.AreEqual(v, r);
        }

        [TestCase(short.MinValue, new byte[]{(byte)Type.Short, 0x00, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.Short, 0x80, 0x00})]
        [TestCase(short.MaxValue, new byte[]{(byte)Type.Short, 0xff, 0xff})]
        public void TestShort(short v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadShort();
            Assert.AreEqual(v, r);
        }

        [TestCase(ushort.MinValue, new byte[]{(byte)Type.UShort, 0x00, 0x00})]
        [TestCase(ushort.MaxValue, new byte[]{(byte)Type.UShort, 0xff, 0xff})]
        public void TestUShort(ushort v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadUShort();
            Assert.AreEqual(v, r);
        }

        [TestCase(int.MinValue, new byte[]{(byte)Type.Int, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.Int, 0x80, 0x00, 0x00, 0x00})]
        [TestCase(int.MaxValue, new byte[]{(byte)Type.Int, 0xff, 0xff, 0xff, 0xff})]
        public void TestInt(int v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadInt();
            Assert.AreEqual(v, r);
        }

        [TestCase(uint.MinValue, new byte[]{(byte)Type.UInt, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(uint.MaxValue, new byte[]{(byte)Type.UInt, 0xff, 0xff, 0xff, 0xff})]
        public void TestUInt(uint v, byte[] expect)
        {
            writer.Reset();
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadUInt();
            Assert.AreEqual(v, r);
        }

        [TestCase(long.MinValue, new byte[]{(byte)Type.Long, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(0, new byte[]{(byte)Type.Long, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(long.MaxValue, new byte[]{(byte)Type.Long, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestLong(long v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadLong();
            Assert.AreEqual(v, r);
        }

        [TestCase(ulong.MinValue, new byte[]{(byte)Type.ULong, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(ulong.MaxValue, new byte[]{(byte)Type.ULong, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestULong(ulong v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadULong();
            Assert.AreEqual(v, r);
        }

        [TestCase("", new byte[]{(byte)Type.Str8, 0})]
        [TestCase("abc", new byte[]{(byte)Type.Str8, 3, 0x61, 0x62, 0x63})]
        [TestCase("あ", new byte[]{(byte)Type.Str8, 3, 0xe3, 0x81, 0x82})]
        [TestCase("🍣🍺", new byte[]{(byte)Type.Str8, 8, 0xF0, 0x9F, 0x8D, 0xA3, 0xF0,0x9F, 0x8D, 0xBA})]
        public void TestStr8(string v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadString();
            Assert.AreEqual(v, r);
        }

        [TestCase(256)]
        [TestCase(0xffff)]
        public void Test16(int len)
        {
            var v = "";
            var expect = new byte[3+len];
            expect[0] = (byte)Type.Str16;
            expect[1] = (byte)((len & 0xff00)>>8);
            expect[2] = (byte)(len & 0xff);
            for (int i=0; i<len; i++)
            {
                v += "A";
                expect[3+i] = (byte)'A';

            }

            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadString();
            Assert.AreEqual(v, r);
        }

        [Test]
        public void TestObject()
        {
            var v = new Obj1(1, "abc");
            var expect = new byte[]{
                (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80, 0x00, 0x00, 0x01,
                (byte)Type.Str8, 3, 0x61, 0x62, 0x63,
            };

            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadObject<Obj1>();

            Assert.AreEqual(v, r);

            var v2 = new Obj2(2, v);
            expect = new byte[]{
                (byte)Type.Obj, (byte)'B', 0, 17,
                (byte)Type.Short, 0x80, 0x02,
                (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80, 0x00, 0x00, 0x01,
                (byte)Type.Str8, 3, 0x61, 0x62, 0x63,
            };
            writer.Reset();
            writer.Write(v2);
            Assert.AreEqual(expect, writer.ArraySegment());

            reader = Serialization.NewReader(writer.ArraySegment());
            var r2 = new Obj2();
            r2 = reader.ReadObject(r2);

            Assert.AreEqual(v2, r2);
        }

        [Test]
        public void TestList()
        {
            var v = new object[]{
                (byte) 10,
                new Obj1(11,"abc"),
                new List<object>(){
                    (byte)20,
                    new Obj2(21, new Obj1(21, "def")),
                },
                new Obj1(12,"ghi"),
            };
            var expect = new byte[]{
                (byte)Type.List,
                (byte)v.Length,

                // byte(10)
                0, 2, (byte)Type.Byte, 10,

                // Obj1(11,"abc")
                0, 14, (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80,0,0,11, (byte)Type.Str8, 3, 0x61, 0x62,0x63,

                // List<object>
                0, 29, (byte)Type.List, 2,
                // byte(20)
                0, 2, (byte)Type.Byte, 20,
                // Obj2(21, Obj1)
                0, 21, (byte)Type.Obj, (byte)'B', 0, 17,
                // short(21)
                (byte)Type.Short, 0x80, 21,
                // Obj1(21, "def")
                (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80,0,0,21, (byte)Type.Str8, 3, 0x64,0x65,0x66,

                // Obj1(11,"abc")
                0, 14, (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80,0,0,12, (byte)Type.Str8, 3, 0x67, 0x68,0x69,
            };

            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var recycle = new List<object>(){
                null,
                new Obj1(0,""),
            };
            var r = reader.ReadList(recycle);

            Assert.AreEqual(v, r);
        }

    }
}
