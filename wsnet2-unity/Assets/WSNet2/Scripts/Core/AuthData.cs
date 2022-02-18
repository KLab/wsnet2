namespace WSNet2.Core
{
    public class AuthData
    {
        public string MACKey { get; }
        public string Bearer { get; }
        public string EncryptedMACKey { get; }

        public AuthData(string macKey, string bearer, string encMKey)
        {
            MACKey = macKey;
            Bearer = bearer;
            EncryptedMACKey = encMKey;
        }
    }
}
