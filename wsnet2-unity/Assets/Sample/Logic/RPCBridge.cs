using System;
using WSNet2;

namespace Sample.Logic
{
    public class RPCBridge
    {
        Room room;

        IMasterClient master;
        IClient client;

        RPCBridge(Room room)
        {
            this.room = room;
            room.RegisterRPC<PlayerEvent>(RPCPlayerEvent);
            room.RegisterRPC<GameState>(RPCSyncGameState);
            room.RegisterRPC((Action<string, long>)RPCSyncServerTick);
        }

        /// <summary>
        /// Constractor for MasterClient
        /// </summary>
        public RPCBridge(Room room, IMasterClient master) : this(room)
        {
            this.master = master;
        }

        /// <summary>
        /// Constractor for Client
        /// </summary>
        public RPCBridge(Room room, IClient client) : this(room)
        {
            this.client = client;
        }

        /// <summary>
        /// Send PlayerEvent (Client -> MasterClient)
        /// </summary>
        public void PlayerEvent(PlayerEvent ev)
        {
            room.RPC(RPCPlayerEvent, ev, Room.RPCToMaster);
        }

        void RPCPlayerEvent(string sender, PlayerEvent ev)
        {
            master?.OnPlayerEvent(sender, ev);
        }

        /// <summary>
        /// Synchronize GameState (MasterClient -> Client)
        /// </summary>
        public void SyncGameState(GameState state)
        {
            room.RPC(RPCSyncGameState, state);
        }

        void RPCSyncGameState(string sender, GameState state)
        {
            if (sender == room.Master.Id)
            {
                client?.OnSyncGameState(state);
            }
        }

        /// <summary>
        /// Synchronize server tick (MasterClient -> Client)
        /// </summary>
        public void SyncServerTick(long tick)
        {
            room.RPC(RPCSyncServerTick, tick);
        }

        void RPCSyncServerTick(string sender, long tick)
        {
            if (sender == room.Master.Id)
            {
                client?.OnSyncServerTick(tick);
            }
        }
    }

    public interface IMasterClient
    {
        void OnPlayerEvent(string sender, PlayerEvent ev);
    }

    public interface IClient
    {
        void OnSyncGameState(GameState state);
        void OnSyncServerTick(long tick);
    }
}
