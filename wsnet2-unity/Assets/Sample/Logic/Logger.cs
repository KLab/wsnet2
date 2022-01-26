namespace Sample.Logic
{
    static class Logger
    {
        public static void Debug(string format, params object[] args)
        {
#if UNITY_5_3_OR_NEWER
            UnityEngine.Debug.Log(string.Format(format, args));
#else
            System.Console.WriteLine(string.Format(format, args));
#endif
        }
    }
}
