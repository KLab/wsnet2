using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.InputSystem;

public class GameScript : MonoBehaviour
{
    public InputAction moveInput;
    public Ball ballAsset;
    public Bar barAsset;

    public static Vector2 bottomLeft;
    public static Vector2 topRight;

    Bar bar1;
    Bar bar2;
    Ball ball;

    Bar playerBar;

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
        ball.setRandomDirection();
        ball.speed = 3f;

        playerBar = bar1;
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
                    ball.direction.x *= -1;
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

            if (bx - br <= px + pw / 2f && bx + br >= px + pw / 2f)
            {
                if (py - ph / 2f <= by && py + ph / 2f >= by)
                {
                    ball.direction.x *= -1;
                }
            }
        }
    }
}
