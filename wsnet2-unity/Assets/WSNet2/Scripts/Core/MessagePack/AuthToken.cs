using MessagePack;

namespace WSNet2.Core
{
    [MessagePackObject]
    public class AuthToken
    {
        [Key("nonce")]
        public string nonce;

        [Key("hash")]
        public string hash;
    }
}
