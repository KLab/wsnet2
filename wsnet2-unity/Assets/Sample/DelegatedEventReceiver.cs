using System;
using UnityEngine;
using WSNet2.Core;

namespace Sample
{
    public class DelegatedEventReceiver : WSNet2.Core.EventReceiver
    {
        public Action<Exception> OnErrorDelegate;
        public Action<Player> OnJoinedDelegate;
        public Action<Player> OnOtherPlayerJoinedDelegate;
        public Action<Player> OnOtherPlayerLeftDelegate;
        public Action<Player, Player> OnMasterPlayerSwitchedDelegate;
        public Action<string> OnClosedDelegate;


        public override void OnError(Exception e)
        {
            Debug.Log("OnError: " + e);
            if (OnErrorDelegate != null)
            {
                OnErrorDelegate(e);
            }
        }

        public override void OnJoined(Player me)
        {
            Debug.Log("OnJoined: " + me.Id);
            if (OnJoinedDelegate != null)
            {
                OnJoinedDelegate(me);
            }
        }

        public override void OnOtherPlayerJoined(Player player)
        {
            Debug.Log("OnOtherPlayerJoined: " + player.Id);
            if (OnOtherPlayerJoinedDelegate != null)
            {
                OnOtherPlayerJoinedDelegate(player);
            }
        }

        public override void OnOtherPlayerLeft(Player player)
        {
            Debug.Log("OnLeave: " + player.Id);
            if (OnOtherPlayerLeftDelegate != null)
            {
                OnOtherPlayerLeftDelegate(player);
            }
        }

        public override void OnMasterPlayerSwitched(Player pred, Player newly)
        {
            Debug.Log("OnMasterPlayerSwitched: " + pred.Id + " " + newly.Id);
            if (OnMasterPlayerSwitchedDelegate != null)
            {
                OnMasterPlayerSwitchedDelegate(pred, newly);
            }
        }

        public override void OnClosed(string description)
        {
            Debug.Log("OnClose: " + description);
            if (OnClosedDelegate != null)
            {
                OnClosedDelegate(description);
            }
        }
    }
}