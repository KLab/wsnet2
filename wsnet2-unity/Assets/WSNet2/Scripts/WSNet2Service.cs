using UnityEngine;
using UnityEngine.Networking;
using System;
using System.Collections;
using System.Collections.Generic;
using System.Threading.Tasks;

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
        Action doOnUpdate;
        object doOnUpdateLock;

        IWSNet2Logger<WSNet2LogPayload> defaultLogger;

        void Initialize()
        {
            clients = new Dictionary<string, WSNet2Client>();
            newClients = new Dictionary<string, WSNet2Client>();
            DontDestroyOnLoad(this.gameObject);
            defaultLogger = new DefaultUnityLogger();
            doOnUpdateLock = new object();
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
        /// <param name="logger">Logger</param>
        public WSNet2Client GetClient(string baseUri, string appId, string userId, AuthData authData, IWSNet2Logger<WSNet2LogPayload> logger=null)
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

            cli = new WSNet2Client(baseUri, appId, userId, authData, logger ?? defaultLogger);
            cli.HttpPost = httpPost;

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

            lock (doOnUpdateLock)
            {
                doOnUpdate?.Invoke();
                doOnUpdate = null;
            }
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

        void httpPost(string url, IReadOnlyDictionary<string, string> headers, byte[] content, TaskCompletionSource<(int, byte[])> tcs)
        {
            lock (doOnUpdateLock)
            {
                doOnUpdate += () => StartCoroutine(doPost(url, headers, content, tcs));
            }
        }

        IEnumerator doPost(string url, IReadOnlyDictionary<string, string> headers, byte[] content, TaskCompletionSource<(int, byte[])> tcs)
        {
            using var uploadHandler = new UploadHandlerRaw(content);
            using var downloadHandler = new DownloadHandlerBuffer();
            using var req = new UnityWebRequest(url, "POST", downloadHandler, uploadHandler);

            foreach (var h in headers)
            {
                req.SetRequestHeader(h.Key, h.Value);
            }

            yield return req.SendWebRequest();

            // 接続できないなどレスポンスを受け取れないケースや中断
            if (req.responseCode == 0 || !downloadHandler.isDone)
            {
                try
                {
                    // stack traceを記録するため一回throw
                    throw new Exception($"http post failed: {req.error} {url}");
                }
                catch (Exception e)
                {
                    tcs.TrySetException(e);
                }

                yield break;
            }

            tcs.TrySetResult(((int)req.responseCode, downloadHandler.data));
        }
    }
}
