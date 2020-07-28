using System;
using System.Collections;
using System.Collections.Generic;
using UnityEngine;

public class Bar : MonoBehaviour
{
    public float width;
    public float height;
    public float speed;

    bool isLeft;

    public void Init(bool leftSide)
    {
        speed = 5f;
        width = gameObject.GetComponent<Renderer>().bounds.size.x;
        height = gameObject.GetComponent<Renderer>().bounds.size.y;

        Vector2 pos = Vector2.zero;

        if (leftSide)
        {
            isLeft = true;
            pos = new Vector2(GameScript.bottomLeft.x, 0);
            pos += Vector2.right * 50;
        }
        else
        {
            isLeft = false;
            pos = new Vector2(GameScript.topRight.x, 0);
            pos -= Vector2.right * 50;
        }

        transform.position = pos;
    }

    // Start is called before the first frame update
    void Start()
    {

    }

    // Update is called once per frame
    void Update()
    {

    }
}
