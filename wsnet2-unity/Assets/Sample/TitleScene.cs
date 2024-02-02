using System;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.SceneManagement;
using MessagePack;
using MessagePack.Resolvers;
using Sample.Logic;
using WSNet2;

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
        /// </summary>
        public static uint MaxPlayers = 2;

        /// <summary>
        /// タイムアウト(秒)
        /// </summary>
        public static uint Deadline = 3;

        static TitleScene()
        {
            StaticCompositeResolver.Instance.Register(
                GeneratedResolver.Instance, StandardResolver.Instance);

            var option = MessagePackSerializerOptions.Standard.WithResolver(StaticCompositeResolver.Instance);

            MessagePackSerializer.DefaultOptions = option;
        }

        /// <summary>
        /// CPU戦ボタンコールバック
        /// </summary>
        public void OnClickVsCpu()
        {
            Debug.Log("OnClickVsCpu");
            SceneManager.LoadScene("Game");
        }

        /// <summary>
        /// 部屋作成ボタンコールバック
        /// </summary>
        public void OnClickCreate()
        {
            Debug.Log("OnClickCreate");
            var pubProps = new Dictionary<string, object>(){
                {WSNet2Helper.PubKey.Game, WSNet2Helper.GameName},
                {WSNet2Helper.PubKey.State, GameStateCode.WaitingPlayer.ToString()},
                {WSNet2Helper.PubKey.PlayerNum, (byte)1},
                {WSNet2Helper.PubKey.Updated, new DateTimeOffset(DateTime.Now).ToUnixTimeSeconds()},
            };
            var cliProps = new Dictionary<string, object>(){
                {"userId", userIdInput.text},
            };
            var roomOpt = new RoomOption(MaxPlayers, SearchGroup, pubProps, null);
            roomOpt.WithClientDeadline(Deadline);

            prepareWSNet2Client();
            G.Client.Create(
                roomOpt,
                cliProps,
                (room) =>
                {
                    room.Pause();
                    Debug.Log("created: room=" + room.Id);
                    G.GameRoom = room;
                    SceneManager.LoadScene("Game");
                },
                (e) => Debug.Log("create failed: " + e),
                new DefaultUnityLogger()
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

            var query = new Query();
            query.Equal("game", "pong");

            prepareWSNet2Client();
            G.Client.RandomJoin(
                SearchGroup,
                query,
                cliProps,
                (room) =>
                {
                    room.Pause();
                    Debug.Log("join: room=" + room.Id);
                    G.GameRoom = room;
                    SceneManager.LoadScene("Game");
                },
                (e) => Debug.Log("join failed: " + e),
                new DefaultUnityLogger()
            );
        }

        /// <summary>
        /// ランダム観戦ボタンコールバック
        /// </summary>
        public void OnClickRandomWatch()
        {
            Debug.Log("OnClickRandomWatch");
            var query = new Query();
            query.Equal("game", "pong");

            prepareWSNet2Client();
            G.Client.Search(SearchGroup, query, 1, false, true,
            (rooms) => {
                if (rooms.Length == 0)
                {
                    Debug.Log("search failed: no room found");
                    return;
                }

                G.Client.Watch(rooms[0].Id, null,
                (room) => {
                    room.Pause();
                    Debug.Log("watch: room=" + room.Id);
                    G.GameRoom = room;
                    SceneManager.LoadScene("Game");
                },
                (e) => Debug.Log("watch failed: " + e),
                new DefaultUnityLogger());
            },
            (e) => Debug.Log("search failed: " + e));
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
            var authData = new AuthDataGenerator().Generate(appKeyInput.text, userIdInput.text);
            Debug.Log($"lobby {lobbyInput.text}");
            Debug.Log($"appId {appIdInput.text}");
            Debug.Log($"appKey {appKeyInput.text}");
            Debug.Log($"userId {userIdInput.text}");

            Logic.WSNet2Helper.RegisterTypes();
            G.Client = WSNet2Service.Instance.GetClient(
                lobbyInput.text,
                appIdInput.text,
                userIdInput.text,
                authData);
        }
    }
}
