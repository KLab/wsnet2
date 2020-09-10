using System.Collections.Generic;
using MessagePack;

namespace WSNet2.Core
{
    /// <summary>
    ///   部屋プロパティのマッチ条件クエリ
    /// </summary>
    public class Query
    {
        public enum Op : byte
        {
            Equal = 0,
            Not,
            LessThan,
            LessEqual,
            GreaterThan,
            GreaterEqual,
        }

        [MessagePackObject]
        public class Condition
        {
            [Key(0)]
            public string key;

            [Key(1)]
            public Op op;

            [Key(2)]
            public byte[] val;

            public Condition(){}

            public Condition(string key, Op op, byte[] val)
            {
                this.key = key;
                this.op = op;
                this.val = val;
            }
        }

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

                foreach (var cs in  cslist)
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

        private void and(Condition cond)
        {
            foreach (var cs in condsList)
            {
                cs.Add(cond);
            }
        }

        private byte[] serialize(int val)
        {
            var writer = Serialization.GetWriter();
            lock(writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }

        private byte[] serialize(string val)
        {
            var writer = Serialization.GetWriter();
            lock(writer)
            {
                writer.Reset();
                writer.Write(val);
                return writer.ToArray();
            }
        }
    }
}
