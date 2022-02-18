using System;

namespace WSNet2.Core
{
    /// <summary>
    ///   入室可能な部屋が見つからなかった例外
    /// </summary>
    public class RoomNotFoundException : Exception
    {
        public RoomNotFoundException(string message) : base(message) { }
    }
}
