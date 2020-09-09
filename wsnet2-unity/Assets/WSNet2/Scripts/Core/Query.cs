using System.Collections.Generic;
using MessagePack;

namespace WSNet2.Core
{
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

        static SerialWriter writer = Serialization.NewWriter(64);

        internal List<List<Condition>> condsList;

        public Query()
        {
            condsList = new List<List<Condition>>()
            {
                new List<Condition>(),
            };
        }

        public Query Or(params Query[] queries)
        {
            var cslist = condsList;
            condsList = new List<List<Condition>>();

            foreach (var cs in cslist)
            {
                foreach (var q in queries)
                {
                    foreach (var qcs in q.condsList)
                    {
                        var ncs = new List<Condition>();
                        ncs.AddRange(cs);
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
