using UnityEngine;
using System.Collections.Generic;
using WSNet2.Core;

namespace WSNet2
{
    /// <summary>
    ///   UnityでWSNet2を扱うためのGameObject
    /// </summary>
    public class WSNet2Service : MonoBehaviour
    {
        const char keyDelimiter = '@';

        public static WSNet2Service Instance
        {
            get
            {
                if (instance == null)
                {
                    instance = new GameObject("WSNet2Service").AddComponent<WSNet2Service>();
                    instance.Initialize();
                }

                return instance;
            }
        }

        static WSNet2Service instance;

        Dictionary<string, WSNet2Client> clients;
        Dictionary<string, WSNet2Client> newClients;

        void Initialize()
        {
            clients = new Dictionary<string, WSNet2Client>();
            newClients = new Dictionary<string, WSNet2Client>();
            DontDestroyOnLoad(this.gameObject);

            if (WSNet2Logger.Logger is WSNet2Logger.DefaultConsoleLogger) {
                WSNet2Logger.Logger = new DefaultUnityLogger();
                WSNet2Logger.Info("DefaultUnityLogger Installed");
            }
        }

        /// <summary>
        ///   WSNet2Clientを取得
        /// </summary>
        /// <remarks>
        ///   すでにインスタンスがある場合は、baseUriとauthDataを更新して使い回す
        /// </remarks>
        /// <param name="baseUri">LobbyのURI</param>
        /// <param name="appId">WSNetに登録してあるApplicationID</param>
        /// <param name="userId">プレイヤーID</param>
        /// <param name="authData">認証データ</param>
        public WSNet2Client GetClient(string baseUri, string appId, string userId, string authData)
        {
            var key = $"{userId}{keyDelimiter}{appId}";

            WSNet2Client cli;
            if (clients.TryGetValue(key, out cli))
            {
                cli.SetBaseUri(baseUri);
                cli.UpdateAuthData(authData);
                return cli;
            }

            if (newClients.TryGetValue(key, out cli))
            {
                cli.SetBaseUri(baseUri);
                cli.UpdateAuthData(authData);
                return cli;
            }

            cli = new WSNet2Client(baseUri, appId, userId, authData);

            // Note: ProcessCallback 内で clients の追加を直接行うと InvalidOperationException が発生する
            newClients[key] = cli;
            return cli;
        }

        void Update()
        {
            mergeNewClients();
            foreach (var cli in clients.Values)
            {
                cli.ProcessCallback();
            }
            mergeNewClients();
        }

        void OnDestroy()
        {
            mergeNewClients();
            foreach (var cli in clients.Values)
            {
                cli.ForceDisconnect();
            }

            instance = null;
        }

        void mergeNewClients() {
            // Note: ProcessCallback 内で clients の追加を直接行うと InvalidOperationException が発生する
            foreach (var kv in newClients)
            {
                clients[kv.Key] = kv.Value;
            }
            newClients.Clear();
        }
    }
}
