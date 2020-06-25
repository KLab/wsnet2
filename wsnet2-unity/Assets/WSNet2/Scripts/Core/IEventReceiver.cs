using System;

namespace WSNet2.Core
{
    public interface IEventReceiver
    {
        void OnError(Exception e);

        void OnJoined(Player me);

        void OnOtherPlayerJoined(Player player);

        void OnMessage(EvMessage ev);
    }
}
