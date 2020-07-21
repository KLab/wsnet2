using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using WSNet2.Core;

public class SampleClient : MonoBehaviour
{
    class EventReceiver : IEventReceiver
    {
        public void OnError(Exception e)
        {
            Debug.Log("OnError: "+e);
        }

        public void OnJoined(Player me)
        {
            Debug.Log("OnJoined: "+me.Id);
        }

        public void OnOtherPlayerJoined(Player player)
        {
            Debug.Log("OnOtherPlayerJoined: "+player.Id);
        }

        public void OnMessage(EvMessage ev)
        {
            var msg = ev.Body<StrMessage>();
            Debug.Log($"OnMessage[{ev.SenderID}]: {msg}");
        }

        public void OnLeave(Player player)
        {
            Debug.Log("OnLeave: "+player.Id);
        }

        public void OnClosed(string description)
        {
            Debug.Log("OnClose: "+description);
        }
    }

    public class StrMessage : IWSNetSerializable
    {
        string str;

        public StrMessage(){}
        public StrMessage(string str)
        {
            this.str = str;
        }

        public void Serialize(SerialWriter writer)
        {
            writer.Write(str);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            str = reader.ReadString();
        }

        public override string ToString()
        {
            return str;
        }
    }

    WSNet2Client cli;



    // Start is called before the first frame update
    void Start()
    {
        Serialization.Register<StrMessage>(1);

        cli = new WSNet2Client(
            "http://localhost:8080",
            "testapp",
            "id0001",
            null);

        var pubProps = new Dictionary<string, object>(){
            {"aaa", "public"},
            {"bbb", (int)13},
        };
        var privProps = new Dictionary<string, object>(){
            {"aaa", "private"},
            {"ccc", false},
        };
        var cliProps = new Dictionary<string, object>(){
            {"name", "FooBar"},
        };
        var roomOpt = new RoomOption(10, 100, pubProps, privProps);

        var receiver = new EventReceiver();

        cli.Create(
            roomOpt,
            cliProps,
            receiver,
            (room) => {
                Debug.Log("created: room="+room.Id);
                StartCoroutine(HandleRoom(room));
                return true;
            },
            (e) => Debug.Log("create failed: "+ e));
    }

    // Update is called once per frame
    void Update()
    {
        cli.ProcessCallback();
    }

    IEnumerator HandleRoom(Room room)
    {
        for(var i = 0; i < 100; i++)
        {
            yield return new WaitForSeconds(1);
            room.Broadcast(new StrMessage("abcdefg" + i));
        }
    }
}
