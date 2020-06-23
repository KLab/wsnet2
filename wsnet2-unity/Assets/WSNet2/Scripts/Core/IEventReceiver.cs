using System;

namespace WSNet2.Core
{
    public interface IEventReceiver
    {
        public void OnError(Exception e);

        public void OnJoined(Player me);

        public void OnOtherPlayerJoined(Player player);
    }
}
