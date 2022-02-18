namespace WSNet2.Core
{
    public enum MsgType
    {
        Ping = 1,

        Leave = MsgTypeExt.regularMsgType,
        RoomProp,
        ClientProp,
        SwitchMaster,
        Target,
        ToMaster,
        Broadcast,
        Kick,
    }

    static class MsgTypeExt
    {
        public const int regularMsgType = 30;

        public static bool IsRegular(this MsgType type) => (int)type >= regularMsgType;
    }
}
