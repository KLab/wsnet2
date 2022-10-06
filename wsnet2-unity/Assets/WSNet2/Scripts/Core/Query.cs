using System.Collections.Generic;
using System.Linq;

namespace WSNet2
{
    /// <summary>
    ///   部屋プロパティのマッチ条件クエリ
    /// </summary>
    public partial class Query
    {
        /// <summary>
        ///   マッチ条件リスト
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     二重配列：外側=OR結合、内側＝AND結合
        ///   </para>
        /// </remarks>
        internal List<List<Condition>> condsList;

        public Query()
        {
            condsList = new List<List<Condition>>()
            {
                new List<Condition>(),
            };
        }

        /// <summary>
        ///   queriesをAND結合して追加する
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     (A+B).And((C+D),(E+F)) = (A+B)*((C+D)*(E+F)) = (ACE+ACF+ADE+ADF+BCE+BCF+BDE+BDF)
        ///   </para>
        /// </remarks>
        public Query And(params Query[] queries)
        {
            if (queries == null || queries.Length == 0)
            {
                return this;
            }

            foreach (var q in queries)
            {
                var cslist = condsList;
                condsList = new List<List<Condition>>();

                foreach (var cs in cslist)
                {
                    foreach (var qcs in q.condsList)
                    {
                        var ncs = new List<Condition>(cs);
                        ncs.AddRange(qcs);
                        condsList.Add(ncs);
                    }
                }
            }

            return this;
        }

        /// <summary>
        ///   queriesをOR結合して追加する
        /// </summary>
        /// <remarks>
        ///   <para>
        ///     (A+B).Or((C+D),(E+F)) = (A+B)*((C+D)+(E+F)) = (AC+AD+AE+AF+BC+BD+BE+BF)
        ///   </para>
        /// </remarks>
        public Query Or(params Query[] queries)
        {
            if (queries == null || queries.Length == 0)
            {
                return this;
            }

            var cslist = condsList;
            condsList = new List<List<Condition>>();

            foreach (var cs in cslist)
            {
                foreach (var q in queries)
                {
                    foreach (var qcs in q.condsList)
                    {
                        var ncs = new List<Condition>(cs);
                        ncs.AddRange(qcs);
                        condsList.Add(ncs);
                    }
                }
            }

            return this;
        }

        /// <summary>
        ///   nullかどうかの条件を追加
        /// </summary>
        public Query IsNull(string key)
        {
            and(new Condition(key, Op.Equal, serialize()));
            return this;
        }
        public Query IsNotNull(string key)
        {
            and(new Condition(key, Op.Not, serialize()));
            return this;
        }

        /// <summary>
        ///   boolの条件を追加
        /// </summary>
        public Query Equal(string key, bool val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, bool val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }

        /// <summary>
        ///   sbyteの条件を追加
        /// </summary>
        public Query Equal(string key, sbyte val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, sbyte val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, sbyte val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, sbyte val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, sbyte val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, sbyte val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, sbyte min, sbyte max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   byteの条件を追加
        /// </summary>
        public Query Equal(string key, byte val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, byte val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, byte val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, byte val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, byte val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, byte val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, byte min, byte max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   Charの条件を追加
        /// </summary>
        public Query Equal(string key, char val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, char val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }

        /// <summary>
        ///   shortの条件を追加
        /// </summary>
        public Query Equal(string key, short val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, short val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, short val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, short val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, short val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, short val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, short min, short max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   ushortの条件を追加
        /// </summary>
        public Query Equal(string key, ushort val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, ushort val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, ushort val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, ushort val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, ushort val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, ushort val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, ushort min, ushort max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   Intの条件を追加
        /// </summary>
        public Query Equal(string key, int val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, int val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, int val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, int val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, int val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, int val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, int min, int max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   uintの条件を追加
        /// </summary>
        public Query Equal(string key, uint val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, uint val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, uint val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, uint val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, uint val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, uint val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, uint min, uint max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   longの条件を追加
        /// </summary>
        public Query Equal(string key, long val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, long val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, long val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, long val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, long val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, long val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, long min, long max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   ulongの条件を追加
        /// </summary>
        public Query Equal(string key, ulong val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, ulong val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, ulong val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, ulong val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, ulong val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, ulong val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, ulong min, ulong max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   floatの条件を追加
        /// </summary>
        public Query Equal(string key, float val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, float val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, float val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, float val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, float val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, float val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, float min, float max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   doubleの条件を追加
        /// </summary>
        public Query Equal(string key, double val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, double val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }
        public Query LessThan(string key, double val)
        {
            and(new Condition(key, Op.LessThan, serialize(val)));
            return this;
        }
        public Query LessEqual(string key, double val)
        {
            and(new Condition(key, Op.LessEqual, serialize(val)));
            return this;
        }
        public Query GreaterThan(string key, double val)
        {
            and(new Condition(key, Op.GreaterThan, serialize(val)));
            return this;
        }
        public Query GreaterEqual(string key, double val)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(val)));
            return this;
        }
        public Query Between(string key, double min, double max)
        {
            and(new Condition(key, Op.GreaterEqual, serialize(min)));
            and(new Condition(key, Op.LessEqual, serialize(max)));
            return this;
        }

        /// <summary>
        ///   stringの条件を追加
        /// </summary>
        public Query Equal(string key, string val)
        {
            and(new Condition(key, Op.Equal, serialize(val)));
            return this;
        }
        public Query Not(string key, string val)
        {
            and(new Condition(key, Op.Not, serialize(val)));
            return this;
        }

        /// <summary>
        ///   Listに含まれる条件を追加
        /// </summary>
        public Query Contain(string key, bool val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, sbyte val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, byte val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, char val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, short val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, ushort val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, int val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, uint val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, long val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, ulong val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, float val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, double val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }
        public Query Contain(string key, string val)
        {
            and(new Condition(key, Op.Contain, serialize(val)));
            return this;
        }

        /// <summary>
        ///   Listに含まれない条件を追加
        /// </summary>
        public Query NotContain(string key, bool val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, sbyte val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, byte val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, char val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, short val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, ushort val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, int val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, uint val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, long val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, ulong val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, float val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, double val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }
        public Query NotContain(string key, string val)
        {
            and(new Condition(key, Op.NotContain, serialize(val)));
            return this;
        }

        public override string ToString()
        {
            var condsListStr = condsList.Select(
                conds =>
                {
                    var ss = conds.Select(
                        cond =>
                        {
                            var op = cond.op switch
                            {
                                Op.Equal => "==",
                                Op.Not => "!=",
                                Op.LessThan => "<",
                                Op.LessEqual => "<=",
                                Op.GreaterThan => ">",
                                Op.GreaterEqual => ">=",
                                Op.Contain => "∋",
                                Op.NotContain => "∌",
                                _ => $"Op({(byte)cond.op})",
                            };
                            var val = WSNet2Serializer.NewReader(cond.val).Read();
                            return $"{{{cond.key} {op} {val}}}";
                        });
                    return $"[{string.Join(",", ss)}]";
                });
            return $"[{string.Join(",", condsListStr)}]";
        }

        private void and(Condition cond)
        {
            foreach (var cs in condsList)
            {
                cs.Add(cond);
            }
        }

        private byte[] serialize()
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write();
                return writer.ToArray();
            }
        }

        private byte[] serialize(bool val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(sbyte val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(byte val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(char val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(short val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(ushort val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(int val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(uint val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(long val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(ulong val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(float val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(double val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(string val)
        {
            var writer = WSNet2Serializer.GetWriter();
            lock (writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }
    }
}
