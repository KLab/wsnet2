#if !UNITY_5_3_OR_NEWER

namespace Sample.Logic
{

    /// <summary>
    /// UnityEngine.Vector2 を Unity非依存の環境で処理するための型
    /// </summary>
    public struct Vector2
    {
        public float x;
        public float y;

        public Vector2(float x, float y)
        {
            this.x = x;
            this.y = y;
        }
    }

    /// <summary>
    /// UnityEngine.Vector3 を Unity非依存の環境で処理するための型
    /// </summary>
    public struct Vector3
    {
        public float x;
        public float y;
        public float z;

        public Vector3(float x, float y, float z)
        {
            this.x = x;
            this.y = y;
            this.z = z;
        }
    }

}

#endif