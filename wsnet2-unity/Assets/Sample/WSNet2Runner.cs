using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using WSNet2.Core;

namespace Sample
{
    public class WSNet2Runner : MonoBehaviour
    {
        /// <summary>
        /// シングルトンインスタンス
        /// </summary>
        public static WSNet2Runner Instance
        {
            get; private set;
        }

        /// <summary>
        /// WSNet2クライアント
        /// </summary>
        public WSNet2Client Client
        {
            get; set;
        }

        /// <summary>
        /// 現在アクティブなゲームルーム
        /// </summary>
        public Room GameRoom
        {
            get; set;
        }

        public static void CreateInstance()
        {
            if (WSNet2Runner.Instance == null)
            {
                Logic.WSNet2Helper.RegisterTypes();
                new GameObject("WSNet2Runner").AddComponent<WSNet2Runner>();
            }
        }

        void Awake()
        {
            if (Instance == null)
            {
                Instance = this;
                DontDestroyOnLoad(gameObject);
            }
            else
            {
                Destroy(gameObject);
            }
        }

        void Start()
        {
            Debug.Log("WSNet2Runner.Start");
        }
    }
}
