using System;
using System.Collections.Generic;

namespace WSNet2.Core
{
    public abstract class EventReceiver
    {
        public Dictionary<Delegate, byte> RPCMap { get; private set; }
        public List<Action<string, SerialReader>> RPCActions { get; private set; }

        public EventReceiver()
        {
            RPCMap = new Dictionary<Delegate, byte>();
            RPCActions = new List<Action<string, SerialReader>>();
        }

        public abstract void OnError(Exception e);
        public abstract void OnJoined(Player me);
        public abstract void OnOtherPlayerJoined(Player player);
        public abstract void OnOtherPlayerLeft(Player player);
        public abstract void OnMasterPlayerSwitched(Player pred, Player newly);
        public abstract void OnClosed(string description);

        public int RegisterRPC(Action<string, string> rpc)
        {
            return registerRPC(
                rpc,
                (senderId, reader) => rpc(senderId, reader.ReadString()));
        }

        public int RegisterRPC<T>(Action<string, T> rpc, bool cacheObject = false) where T : class, IWSNetSerializable, new()
        {
            if (!cacheObject)
            {
                return registerRPC(
                    rpc,
                    (senderId, reader) => rpc(senderId, reader.ReadObject<T>()));
            }

            T obj = new T();

            return registerRPC(
                rpc,
                (senderId, reader) => {
                    obj = reader.ReadObject(obj);
                    rpc(senderId, obj);
                });
        }

        private int registerRPC(Delegate rpc, Action<string, SerialReader> action)
        {
            var id = RPCActions.Count;

            if (id > byte.MaxValue)
            {
                throw new Exception("RPC map full");
            }
            
            if (RPCMap.ContainsKey(rpc))
            {
                throw new Exception("RPC target already registered");
            }
            
            RPCMap[rpc] = (byte)id;
            RPCActions.Add(action);

            return id;
        }
    }
}
