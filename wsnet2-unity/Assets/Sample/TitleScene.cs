using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;
using WSNet2.Core;

namespace Sample
{
    /// <summary>
    /// タイトルシーンのコントローラ
    /// </summary>
    public class TitleScene : MonoBehaviour
    {
        /// <summary>
        /// ロビーのURL入力フォーム
        /// </summary>
        public InputField lobbyInput;

        /// <summary>
        /// appIdの入力フォーム
        /// </summary>
        public InputField appIdInput;

        /// <summary>
        /// appKeyの入力フォーム
        /// </summary>
        public InputField appKeyInput;

        /// <summary>
        /// ユーザIDの入力フォーム
        /// </summary>
        public InputField userIdInput;

        /// <summary>
        /// Pongゲームのサーチグループ
        /// </summary>
        public static uint SearchGroup = 1000;

        /// <summary>
        /// Pongゲームの最大プレイヤー数
        /// 2PlayerとMasterClientの3人
        /// </summary>
        public static uint MaxPlayers = 3;

        /// <summary>
        /// タイムアウト(秒)
        /// </summary>
        public static uint Deadline = 3;


        /// <summary>
        /// 部屋作成ボタンコールバック
        /// </summary>
        public void OnClickCreate()
        {
            Debug.Log("OnClickCreate");

            var pubProps = new Dictionary<string, object>(){
                {"game", "pong"},
                {"masterclient", "waiting"},
            };
            var privProps = new Dictionary<string, object>(){
                {"aaa", "private"},
                {"ccc", false},
            };
            var cliProps = new Dictionary<string, object>(){
                {"userId", userIdInput.text},
            };
            var roomOpt = new RoomOption(MaxPlayers, SearchGroup, pubProps, privProps);
            roomOpt.WithClientDeadline(Deadline);

            prepareWSNet2Client();
            WSNet2Runner.Instance.Client.Create(
                roomOpt,
                cliProps,
                (room) =>
                {
                    room.Pause();
                    Debug.Log("created: room=" + room.Id);
                    WSNet2Runner.Instance.GameRoom = room;
                    SceneManager.LoadScene("Game");
                    return true;
                },
                (e) => Debug.Log("create failed: " + e)
            );
        }

        /// <summary>
        /// ランダム入室ボタンコールバック
        /// </summary>
        public void OnClickRandomJoin()
        {
            Debug.Log("OnClickRandomJoin");

            var cliProps = new Dictionary<string, object>(){
                {"userId", userIdInput.text},
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

            prepareWSNet2Client();
            WSNet2Runner.Instance.Client.RandomJoin(
                SearchGroup,
                queries,
                cliProps,
                (room) =>
                {
                    room.Pause();
                    Debug.Log("join: room=" + room.Id);
                    WSNet2Runner.Instance.GameRoom = room;
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

        /// <summary>
        /// シングルトンのWSNet2Clientのインスタンスを作成し、ProcessCallbackのループを開始する
        /// サーバやユーザIDが決まったあと1度呼び出すこと
        /// </summary>
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