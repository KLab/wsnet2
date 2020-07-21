using NUnit.Framework;
using System;
using System.Collections.Generic;

namespace WSNet2.Core.Test
{
    /// <summary>
    ///   シリアライズ可能なオブジェクトの例
    /// </summary>
    class Obj1 : IWSNetSerializable, IEquatable<Obj1>
    {
        public static int NewCount = 0;

        public int Num;
        public string Str;

        public Obj1()
        {
            NewCount++;
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

    /// <summary>
    ///   ネストしたシリアライズ可能なオブジェクト
    /// </summary>
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
            // this.Objを使い回すことでnewされるのを抑制可能
            // nullでも大丈夫
            Obj = reader.ReadObject(Obj);
        }

        public bool Equals(Obj2 o)
        {
            if (Obj==null)
            {
                return S==o.S && o.Obj == null;
            }

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
            writer = Serialization.NewWriter();
        }

        [SetUp]
        public void Setup()
        {
            writer.Reset();
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

        [TestCase(float.NegativeInfinity, new byte[]{(byte)Type.Float, 0x00, 0x7f, 0xff, 0xff})]
        [TestCase(1.25f, new byte[]{(byte)Type.Float, 0xbf, 0xa0, 0x00, 0x00})]
        [TestCase(-1.25f, new byte[]{(byte)Type.Float, 0x40, 0x5f, 0xff, 0xff})]
        public void TestFloat(float v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadFloat();
            Assert.AreEqual(v, r);
        }

        [TestCase(double.MaxValue, new byte[]{(byte)Type.Double, 0xff, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        [TestCase(1.25f, new byte[]{(byte)Type.Double, 0xbf, 0xf4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        [TestCase(-1.25f, new byte[]{(byte)Type.Double, 0x40, 0x0b, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestDouble(double v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadDouble();
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
            Obj1.NewCount = 0;
            var r = reader.ReadObject<Obj1>();

            Assert.AreEqual(v, r);
            Assert.AreEqual(1, Obj1.NewCount);

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
            var r2 = new Obj2(0, new Obj1());
            Obj1.NewCount = 0;
            r2 = reader.ReadObject(r2);
            Assert.AreEqual(v2, r2);
            Assert.AreEqual(0, Obj1.NewCount);

            Obj1 v3 = null;
            expect = new byte[]{
                (byte)Type.Null,
            };

            writer.Reset();
            writer.Write(v3);
            Assert.AreEqual(expect, writer.ArraySegment());
            reader = Serialization.NewReader(writer.ArraySegment());
            var r3 = reader.ReadObject<Obj1>();
            Assert.AreEqual(v3, r3);

            var v4 = new Obj2(4, null);
            expect = new byte[]{
                (byte)Type.Obj, (byte)'B', 0, 4,
                (byte)Type.Short, 0x80, 4,
                (byte)Type.Null,
            };
            writer.Reset();
            writer.Write(v4);
            Assert.AreEqual(expect, writer.ArraySegment());
            reader = Serialization.NewReader(writer.ArraySegment());
            var r4 = reader.ReadObject<Obj2>();
            Console.WriteLine(r4);
            Assert.AreEqual(v4, r4);
        }

        [Test]
        public void TestList()
        {
            var v = new object[]{
                (byte) 10,
                new Obj1(11,"abc"),
                new List<object>(){ // ネストも可能
                    (byte)20,
                    new Obj2(21, new Obj1(21, "def")),
                },
                new Obj1(12,"ghi"),
                null,
                new bool[]{ true, false, true },
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

                // null
                0, 1, (byte)Type.Null,

                // bool(t, f, t)
                0, 4, (byte)Type.Bools, 0, 3, 0b10100000,
            };

            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var recycle = new List<object>(){
                null,
                new Obj1(0,""), // 同じindex位置に同じ型のObjectがあると使い回す
            };
            Obj1.NewCount = 0;
            var r = reader.ReadList(recycle);
            Assert.AreEqual(v, r);
            Assert.AreEqual(2, Obj1.NewCount);

            reader = Serialization.NewReader(writer.ArraySegment());
            var r2 = reader.ReadArray(recycle);
            Assert.AreEqual(typeof(object[]), r2.GetType());
            Assert.AreEqual(v, r2);
        }

        [Test]
        public void TestReadListT()
        {
            var bin = new byte[]{
                (byte)Type.List, 3,

                0, 14, (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80,0,0,1, (byte)Type.Str8, 3, 0x61, 0x62, 0x63,

                0, 14, (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80,0,0,2, (byte)Type.Str8, 3, 0x64, 0x65, 0x66,

                0, 14, (byte)Type.Obj, (byte)'A', 0, 9,
                (byte)Type.Int, 0x80,0,0,3, (byte)Type.Str8, 2, 0x67, 0x68,
            };
            var expect = new List<Obj1>(){
                new Obj1(1, "abc"),
                new Obj1(2, "def"),
                new Obj1(3, "gh"),
            };

            var reader = Serialization.NewReader(new ArraySegment<byte>(bin));
            var recycle = new List<Obj1>(){
                new Obj1(),
                new Obj1(),
            };
            Obj1.NewCount = 0;
            var r = reader.ReadList<Obj1>(recycle);
            Assert.AreEqual(typeof(List<Obj1>), r.GetType());
            Assert.AreEqual(expect, r);
            Assert.AreEqual(1, Obj1.NewCount);

            reader = Serialization.NewReader(new ArraySegment<byte>(bin));
            var r2 = reader.ReadArray<Obj1>(recycle);
            Assert.AreEqual(typeof(Obj1[]), r2.GetType());
            Assert.AreEqual(expect, r2);
        }

        [Test]
        public void TestDict()
        {
            var v = new Dictionary<string, object>(){
                {"abc", 123},
                {"def", "ghi"},
                {"jkl", new Obj1(10, "mno")},
            };
            var expect = new byte[]{
                (byte)Type.Dict,
                (byte)v.Count,

                // "abc": 123
                3, 0x61, 0x62, 0x63,
                0, 5, (byte)Type.Int, 0x80, 0, 0, 123,

                // "def": "ghi"
                3, 0x64, 0x65, 0x66,
                0, 5, (byte)Type.Str8, 3, 0x67, 0x68, 0x69,

                // "jkl": Obj1(10, "mno")
                3, 0x6a, 0x6b, 0x6c,
                0, 14, (byte)Type.Obj, (byte)'A', 0, 10,
                (byte)Type.Int, 0x80,0,0,10, (byte)Type.Str8, 3, 0x6d, 0x6e, 0x6f,
            };

            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var recycle = new Dictionary<string, object>(){
                {"jkl", new Obj1()},
            };
            Obj1.NewCount = 0;
            var r = reader.ReadDict(recycle);

            Assert.AreEqual(v, r);
            Assert.AreEqual(0, Obj1.NewCount);
        }

        [TestCase(new bool[]{}, new byte[]{(byte)Type.Bools, 0x00, 0x00})]
        [TestCase(new bool[]{true, false, true}, new byte[]{(byte)Type.Bools, 0, 3, 0b10100000})]
        [TestCase(new bool[]{false, false, true, false, true, true, false, true}, new byte[]{(byte)Type.Bools, 0, 8, 0b00101101})]
        [TestCase(new bool[]{true, true, false, true, false, false, true, false, true}, new byte[]{(byte)Type.Bools, 0, 9, 0b11010010, 0b10000000})]
        public void TestBools(bool[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadBools();
            Assert.AreEqual(v, r);
        }

        [TestCase(new sbyte[]{}, new byte[]{(byte)Type.SBytes, 0x00, 0x00})]
        [TestCase(new sbyte[]{0, 1, -128, 127}, new byte[]{(byte)Type.SBytes, 0, 4, 0x80, 0x81, 0x00, 0xff})]
        public void TestSBytes(sbyte[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadSBytes();
            Assert.AreEqual(v, r);
        }

        [TestCase(new byte[]{}, new byte[]{(byte)Type.Bytes, 0x00, 0x00})]
        [TestCase(new byte[]{0, 1, 127, 255}, new byte[]{(byte)Type.Bytes, 0, 4, 0, 1, 127, 255})]
        public void TestBytes(byte[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadBytes();
            Assert.AreEqual(v, r);
        }

        [TestCase(new short[]{}, new byte[]{(byte)Type.Shorts, 0x00, 0x00})]
        [TestCase(new short[]{0, 1, short.MinValue, short.MaxValue},
                  new byte[]{(byte)Type.Shorts, 0, 4, 0x80, 0x00, 0x80, 0x01, 0x00, 0x00, 0xff, 0xff})]
        public void TestShorts(short[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadShorts();
            Assert.AreEqual(v, r);
        }

        [TestCase(new ushort[]{}, new byte[]{(byte)Type.UShorts, 0x00, 0x00})]
        [TestCase(new ushort[]{0, 1, ushort.MaxValue},
                  new byte[]{(byte)Type.UShorts, 0, 3, 0x00, 0x00, 0x00, 0x01, 0xff, 0xff})]
        public void TestUShorts(ushort[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadUShorts();
            Assert.AreEqual(v, r);
        }

        [TestCase(new int[]{}, new byte[]{(byte)Type.Ints, 0x00, 0x00})]
        [TestCase(new int[]{0, 1, int.MinValue, int.MaxValue},
                  new byte[]{(byte)Type.Ints, 0, 4, 0x80, 0x00, 0x00, 0x00, 0x80, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0xff})]
        public void TestInts(int[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadInts();
            Assert.AreEqual(v, r);
        }

        [TestCase(new uint[]{}, new byte[]{(byte)Type.UInts, 0x00, 0x00})]
        [TestCase(new uint[]{0, 1, uint.MaxValue},
                  new byte[]{(byte)Type.UInts, 0, 3, 0x00,0x00,0x00,0x00, 0x00,0x00,0x00,0x01, 0xff,0xff,0xff,0xff})]
        public void TestUInts(uint[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadUInts();
            Assert.AreEqual(v, r);
        }

        [TestCase(new long[]{}, new byte[]{(byte)Type.Longs, 0x00, 0x00})]
        [TestCase(new long[]{0, 1, long.MinValue, long.MaxValue},
                  new byte[]{(byte)Type.Longs, 0, 4,
                             0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                             0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
                             0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                             0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestLongs(long[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadLongs();
            Assert.AreEqual(v, r);
        }

        [TestCase(new ulong[]{}, new byte[]{(byte)Type.ULongs, 0x00, 0x00})]
        [TestCase(new ulong[]{0, 1, ulong.MaxValue},
                  new byte[]{(byte)Type.ULongs, 0, 3,
                             0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                             0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
                             0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})]
        public void TestULongs(ulong[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadULongs();
            Assert.AreEqual(v, r);
        }

        [TestCase(new float[]{}, new byte[]{(byte)Type.Floats, 0, 0})]
        [TestCase(new float[]{0f, float.NegativeInfinity, float.MaxValue, 1.25f},
                  new byte[]{(byte)Type.Floats, 0, 4,
                             0x80, 0x00, 0x00, 0x00,
                             0x00, 0x7f, 0xff, 0xff,
                             0xff, 0x7f, 0xff, 0xff,
                             0xbf, 0xa0, 0x00, 0x00})]
        public void TestFloats(float[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadFloats();
            Assert.AreEqual(v, r);
        }

        [TestCase(new double[]{}, new byte[]{(byte)Type.Doubles, 0, 0})]
        [TestCase(new double[]{0d, double.NegativeInfinity, double.MaxValue, 1.25d},
                  new byte[]{(byte)Type.Doubles, 0, 4,
                             0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
                             0x00, 0x0f, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
                             0xff, 0xef, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
                             0xbf, 0xf4, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})]
        public void TestDoubles(double[] v, byte[] expect)
        {
            writer.Write(v);
            Assert.AreEqual(expect, writer.ArraySegment());

            var reader = Serialization.NewReader(writer.ArraySegment());
            var r = reader.ReadDoubles();
            Assert.AreEqual(v, r);
        }
    }
}
