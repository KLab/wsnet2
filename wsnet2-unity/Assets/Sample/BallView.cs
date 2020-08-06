using UnityEngine;

namespace Sample
{
    public class BallView : MonoBehaviour
    {
        Vector2 direction;
        float speed;

        public void UpdatePosition(Logic.Ball ball)
        {
            direction = ball.Direction;
            transform.position = ball.Position;
            speed = ball.Speed;
            transform.localScale = new Vector3(ball.Radius / 2, ball.Radius / 2, 1);
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

}
