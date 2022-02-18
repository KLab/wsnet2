using System;
using System.Collections;
using System.Collections.Generic;
using System.Security.Cryptography;
using System.Text;
using UnityEngine;
using WSNet2;
using WSNet2.Core;

public class SampleClient : MonoBehaviour
{
    public class StrMessage : IWSNet2Serializable
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
    AuthDataGenerator authgen = new AuthDataGenerator();

    void OnStrMsgRPC(string sender, StrMessage msg)
    {
        Debug.Log("OnStrMsgRPC["+sender+"]: "+msg);
    }

    // Start is called before the first frame update
    void Start()
    {
        Serialization.Register<StrMessage>(1);

        var userid = "id0001";
        cli = WSNet2Service.Instance.GetClient(
            "http://localhost:8080",
            "testapp",
            userid,
            authgen.Generate("testapppkey", userid));

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

        cli.Create(
            roomOpt,
            cliProps,
            (room) => {
                room.RegisterRPC<StrMessage>(OnStrMsgRPC);
                Debug.Log("created: room="+room.Id);
                StartCoroutine(HandleRoom(room));
            },
            (e) => Debug.Log("create failed: "+ e),
            new DefaultUnityLogger());
    }

    IEnumerator HandleRoom(Room room)
    {
        for(var i = 0; i < 100; i++)
        {
            yield return new WaitForSeconds(1);
            var msg = new StrMessage("strmsg "+i);
            room.RPC(OnStrMsgRPC, msg);
        }
    }
}
