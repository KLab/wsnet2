using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;
using WSNet2.Core;

namespace Sample
{
    public class TitleScene : MonoBehaviour
    {
        public InputField lobbyInput;
        public InputField appIdInput;
        public InputField appKeyInput;
        public InputField userIdInput;

        public static uint SearchGroup = 1000;

        public void OnClickCreate()
        {
            Debug.Log("OnClickCreate");

            var pubProps = new Dictionary<string, object>(){
            {"game", "pong"},
        };
            var privProps = new Dictionary<string, object>(){
            {"aaa", "private"},
            {"ccc", false},
        };
            var cliProps = new Dictionary<string, object>(){
            {"userId", userIdInput.text},
        };
            var roomOpt = new RoomOption(2, SearchGroup, pubProps, privProps);
            var receiver = new DelegatedEventReceiver();

            prepareWSNet2Client();
            WSNet2Runner.Instance.Client.Create(
                roomOpt,
                cliProps,
                receiver,
                (room) =>
                {
                    room.Running = false;
                    Debug.Log("created: room=" + room.Id);
                    WSNet2Runner.Instance.GameRoom = room;
                    WSNet2Runner.Instance.GameEventReceiver = receiver;
                    SceneManager.LoadScene("Game");
                    return true;
                },
                (e) => Debug.Log("create failed: " + e)
            );
        }

        public void OnClickRandomJoin()
        {
            Debug.Log("OnClickRandomJoin");

            var cliProps = new Dictionary<string, object>(){
            {"userId", userIdInput.text},
        };
            var query = new Dictionary<string, object>(){
            {"bbb", (int)13},
        };

            var queries = new PropQuery[][]{
            new PropQuery[] {
                new PropQuery{
                    key = "game",
                    op = OpType.Equal,
                    val = Logic.WSNet2Helper.Serialize("pong"),
                },
            },
        };

            var receiver = new DelegatedEventReceiver();

            prepareWSNet2Client();
            WSNet2Runner.Instance.Client.RandomJoin(
                SearchGroup,
                queries,
                cliProps,
                receiver,
                (room) =>
                {
                    room.Running = false;
                    Debug.Log("join: room=" + room.Id);
                    WSNet2Runner.Instance.GameRoom = room;
                    WSNet2Runner.Instance.GameEventReceiver = receiver;
                    SceneManager.LoadScene("Game");
                    return true;
                },
                (e) => Debug.Log("join failed: " + e)
            );
        }

        // Start is called before the first frame update
        void Start()
        {
        }

        // Update is called once per frame
        void Update()
        {
        }

        void prepareWSNet2Client()
        {
            Debug.Log($"lobby {lobbyInput.text}");
            Debug.Log($"appId {appIdInput.text}");
            Debug.Log($"appKey {appKeyInput.text}");
            Debug.Log($"userId {userIdInput.text}");

            WSNet2Runner.CreateInstance();
            WSNet2Runner.Instance.Client = new WSNet2Client(
                lobbyInput.text,
                appIdInput.text,
                userIdInput.text,
                Logic.WSNet2Helper.GenAuthData(appKeyInput.text, userIdInput.text));
        }
    }
}