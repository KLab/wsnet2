using System;

namespace WSNet2
{
    /// <summary>
    ///   ロビーで起こりうる正常系の例外
    /// </summary>
    public class LobbyNormalException : Exception
    {
        public LobbyNormalException(string message) : base(message) { }
    }

    /// <summary>
    ///   サーバ側の部屋数上限に達した例外
    /// </summary>
    public class RoomLimitException : LobbyNormalException
    {
        public RoomLimitException(string message) : base(message) { }
    }

    /// <summary>
    ///   入室可能な部屋が見つからなかった例外
    /// </summary>
    public class RoomNotFoundException : LobbyNormalException
    {
        public RoomNotFoundException(string message) : base(message) { }
    }

    /// <summary>
    ///   満室で入室できなかった例外
    /// </summary>
    public class RoomFullException : LobbyNormalException
    {
        public RoomFullException(string message) : base(message) { }
    }
}
