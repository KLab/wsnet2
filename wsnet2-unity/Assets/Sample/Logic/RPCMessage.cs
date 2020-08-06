using WSNet2.Core;

#if UNITY_5_3_OR_NEWER
using Vector2 = UnityEngine.Vector2;
using Vector3 = UnityEngine.Vector3;
#endif

namespace Sample.Logic
{
    // 空メッセージ
    public class EmptyMessage : IWSNetSerializable
    {
        public EmptyMessage() { }
        public void Serialize(SerialWriter writer) { }
        public void Deserialize(SerialReader reader, int len) { }
    }
}