using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using UnityEngine.InputSystem;
using WSNet2.Core;


public class GameScript : MonoBehaviour
{
    public Text roomText;

    public InputAction moveInput;
    public Ball ballAsset;
    public Bar barAsset;

    public static Vector2 bottomLeft;
    public static Vector2 topRight;

    Bar bar1;
    Bar bar2;
    Ball ball;

    Bar playerBar;

    Bar cpuBar;

    bool isOnlineMode;
    bool isMasterClient;

    float nextSyncTime;

    void RoomLog(string s)
    {
        roomText.text += s + "\n";
    }

    public class SyncPositionMessage : IWSNetSerializable
    {
        // TODO Vector を.NETCore実装と共存させる方法を考える

        public Vector2 bar1Pos;
        public Vector2 bar2Pos;
        public Vector2 ballPos;

        public SyncPositionMessage() { }

        public void Serialize(SerialWriter writer)
        {
            writer.Write(bar1Pos.x);
            writer.Write(bar1Pos.y);
            writer.Write(bar2Pos.x);
            writer.Write(bar2Pos.y);
            writer.Write(ballPos.x);
            writer.Write(ballPos.y);
        }

        public void Deserialize(SerialReader reader, int len)
        {
            bar1Pos.x = reader.ReadFloat();
            bar1Pos.y = reader.ReadFloat();
            bar2Pos.x = reader.ReadFloat();
            bar2Pos.y = reader.ReadFloat();
            ballPos.x = reader.ReadFloat();
            ballPos.y = reader.ReadFloat();
        }
    }

    void RPCSyncPosition(string sender, SyncPositionMessage msg)
    {
        bar1.transform.position = msg.bar1Pos;
        bar2.transform.position = msg.bar2Pos;
        ball.transform.position = msg.ballPos;
    }

    // Start is called before the first frame update
    void Start()
    {
        moveInput.Enable();

        bottomLeft = Camera.main.ScreenToWorldPoint(new Vector2(0, 0));
        topRight = Camera.main.ScreenToWorldPoint(new Vector2(Screen.width, Screen.height));

        bar1 = Instantiate(barAsset);
        bar2 = Instantiate(barAsset);
        ball = Instantiate(ballAsset);
        bar1.Init(true);
        bar2.Init(false);

        playerBar = bar1;

        isOnlineMode = WSNet2Runner.Instance != null && WSNet2Runner.Instance.GameRoom != null;

        if (isOnlineMode)
        {
            roomText.text = "Room:" + WSNet2Runner.Instance.GameRoom.Id + "\n";

            // Roomの処理を開始する前に EventReceiver と RPC の登録を行う必要がある
            WSNet2Runner.Instance.GameEventReceiver.OnClosedDelegate += reason =>
            {
                RoomLog("OnClosed:" + reason);
            };

            WSNet2Runner.Instance.GameEventReceiver.OnErrorDelegate += e =>
            {
                RoomLog("OnError:" + e);
            };

            WSNet2Runner.Instance.GameEventReceiver.OnJoinedDelegate += p =>
            {
                RoomLog("OnJoined:" + p.Id);
            };

            WSNet2Runner.Instance.GameEventReceiver.OnMasterPlayerSwitchedDelegate += (prev, cur) =>
            {
                RoomLog("OnMasterPlayerSwitched:" + prev.Id + " -> " + cur.Id);
            };

            WSNet2Runner.Instance.GameEventReceiver.OnOtherPlayerJoinedDelegate += (p) =>
            {
                RoomLog("OnOtherPlayerJoined:" + p.Id);
            };

            WSNet2Runner.Instance.GameEventReceiver.OnOtherPlayerLeftDelegate += (p) =>
            {
                RoomLog("OnOtherPlayerLeft:" + p.Id);
            };

            WSNet2Runner.Instance.GameEventReceiver.RegisterRPC<SyncPositionMessage>(RPCSyncPosition);

            WSNet2Runner.Instance.GameRoom.Running = true;

            if (WSNet2Runner.Instance.GameRoom.Master.Id == WSNet2Runner.Instance.GameRoom.Me.Id)
            {
                // TODO 仮 .NETCore実装が MasterClientになる予定
                isMasterClient = true;
            }
        }
        else
        {
            roomText.text = "";
            cpuBar = bar2;
            RestartGame();
        }
    }

    void RestartGame()
    {
        ball.setRandomDirection();
        ball.speed = 3f;
    }

    void Update()
    {
        if (playerBar != null)
        {
            var pos = new Vector2(playerBar.transform.position.x, playerBar.transform.position.y);
            var value = moveInput.ReadValue<float>();
            var maxY = topRight.y - playerBar.height / 2f;
            var minY = bottomLeft.y + playerBar.height / 2f;

            pos += Vector2.up * value * playerBar.speed;
            pos.y = Math.Min(pos.y, maxY);
            pos.y = Math.Max(pos.y, minY);
            playerBar.transform.position = pos;
        }

        if (cpuBar != null)
        {
            var pos = new Vector2(cpuBar.transform.position.x, cpuBar.transform.position.y);
            var maxY = topRight.y - cpuBar.height / 2f;
            var minY = bottomLeft.y + cpuBar.height / 2f;

            pos.y = ball.transform.position.y;
            pos.y = Math.Min(pos.y, maxY);
            pos.y = Math.Max(pos.y, minY);
            cpuBar.transform.position = pos;
        }

        if (ball != null)
        {
            ball.transform.Translate(ball.direction * ball.speed);

            if (ball.transform.position.y < bottomLeft.y + ball.radius ||
                ball.transform.position.y > topRight.y - ball.radius)
            {
                ball.direction.y *= -1;
            }
        }

        if (bar1 != null)
        {
            var bx = ball.transform.position.x;
            var by = ball.transform.position.y;
            var br = ball.radius;

            var px = bar1.transform.position.x;
            var py = bar1.transform.position.y;
            var pw = bar1.width;
            var ph = bar1.height;

            if (bx - br <= px + pw / 2f && bx + br >= px + pw / 2f)
            {
                if (py - ph / 2f <= by && py + ph / 2f >= by)
                {
                    if (ball.direction.x < 0)
                    {
                        ball.direction.x *= -1;
                    }
                }
            }
        }

        if (bar2 != null)
        {
            var bx = ball.transform.position.x;
            var by = ball.transform.position.y;
            var br = ball.radius;

            var px = bar2.transform.position.x;
            var py = bar2.transform.position.y;
            var pw = bar2.width;
            var ph = bar2.height;

            if (bx - br <= px - pw / 2f && bx + br >= px - pw / 2f)
            {
                if (py - ph / 2f <= by && py + ph / 2f >= by)
                {
                    if (ball.direction.x > 0)
                    {
                        ball.direction.x *= -1;
                    }
                }
            }
        }


        if (isMasterClient)
        {
            nextSyncTime -= Time.deltaTime;
            if (nextSyncTime < 0)
            {
                nextSyncTime = 0.1f;
                WSNet2Runner.Instance.GameRoom.RPC(RPCSyncPosition, new SyncPositionMessage
                {
                    bar1Pos = bar1.transform.position,
                    bar2Pos = bar2.transform.position,
                    ballPos = ball.transform.position,
                });

            }
        }
    }
}
