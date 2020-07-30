using System;
using UnityEngine;
using WSNet2.Core;

public class DelegatedEventReceiver : WSNet2.Core.EventReceiver
{
    Action<Exception> OnErrorDelegate;
    Action<Player> OnJoinedDelegate;
    Action<Player> OnOtherPlayerJoinedDelegate;
    Action<Player> OnLeaveDelegate;
    Action<string> OnClosedDelegate;


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

    public override void OnLeave(Player player)
    {
        Debug.Log("OnLeave: " + player.Id);
        if (OnLeaveDelegate != null)
        {
            OnLeaveDelegate(player);
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
