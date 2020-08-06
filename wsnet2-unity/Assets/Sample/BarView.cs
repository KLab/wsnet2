using UnityEngine;

namespace Sample
{
    public class BarView : MonoBehaviour
    {
        Vector2 direction;
        float speed;

        public void UpdatePosition(Logic.Bar bar)
        {
            direction = bar.Direction;
            transform.position = bar.Position;
            speed = bar.Speed;
            transform.localScale = new Vector3(bar.Width / 4, bar.Height / 4, 1);
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