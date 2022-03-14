using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   サーバ側の部屋数上限に達した例外
    /// </summary>
    public class RoomLimitException : Exception
    {
        public RoomLimitException(string message) : base(message) { }
    }

    /// <summary>
    ///   入室可能な部屋が見つからなかった例外
    /// </summary>
    public class RoomNotFoundException : Exception
    {
        public RoomNotFoundException(string message) : base(message) { }
    }

    /// <summary>
    ///   満室で入室できなかった例外
    /// </summary>
    public class RoomFullException : Exception
    {
        public RoomFullException(string message) : base(message) { }
    }
}
