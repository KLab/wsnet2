using System;

namespace WSNet2.Core
{
    public interface IEventReceiver
    {
        public void OnJoin(ClientInfo cinfo);
    }
}
