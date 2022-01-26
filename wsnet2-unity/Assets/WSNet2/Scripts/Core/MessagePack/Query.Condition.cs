using MessagePack;

namespace WSNet2.Core
{
    public partial class Query
    {
        public enum Op : byte
        {
            Equal = 0,
            Not,
            LessThan,
            LessEqual,
            GreaterThan,
            GreaterEqual,
            Contain,
            NotContain,
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

            public Condition() { }

            public Condition(string key, Op op, byte[] val)
            {
                this.key = key;
                this.op = op;
                this.val = val;
            }
        }
    }
}
