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

    void RoomLog(string s)
    {
        roomText.text += s + "\n";
    }

    class GameEventReceiver : WSNet2.Core.EventReceiver
    {
        public override void OnError(Exception e)
        {
            Debug.Log("OnError: " + e);
            RoomLog("OnError:" + e);
        }

        public override void OnJoined(Player me)
        {
            Debug.Log("OnJoined: " + me.Id);
        }

        public override void OnOtherPlayerJoined(Player player)
        {
            Debug.Log("OnOtherPlayerJoined: " + player.Id);
        }

        public override void OnLeave(Player player)
        {
            Debug.Log("OnLeave: " + player.Id);
        }

        public override void OnClosed(string description)
        {
            Debug.Log("OnClose: " + description);
        }
    }

    public static EventReceiver CreateEventReceiver()
    {
        var r = new GameEventReceiver();

        // r.RegisterRPC<StrMessage>(OnStrMsgRPC);

        return r;
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

    // Update is called once per frame
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

    }
}
