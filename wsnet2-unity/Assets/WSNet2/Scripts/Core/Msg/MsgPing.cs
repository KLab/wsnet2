using System;

namespace WSNet2.Core
{
    public class MsgPing
    {
        public readonly ArraySegment<byte> Value;

        public MsgPing()
        {
            Value = new ArraySegment<byte>(new byte[]{(byte)MsgType.Ping});
        }
    }
}
