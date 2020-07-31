
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using WSNet2.Core;

public class WSNet2Runner : MonoBehaviour
{
    public static WSNet2Runner Instance
    {
        get; private set;
    }

    public WSNet2Client Client
    {
        get; set;
    }

    // Active game room.
    public Room GameRoom
    {
        get; set;
    }

    // Active game receiver.
    public DelegatedEventReceiver GameEventReceiver
    {
        get; set;
    }

    public static void CreateInstance()
    {
        if (WSNet2Runner.Instance == null)
        {
            WSNet2Helper.RegisterTypes();
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

    void Update()
    {
        if (Client != null)
        {
            Client.ProcessCallback();
        }
    }
}