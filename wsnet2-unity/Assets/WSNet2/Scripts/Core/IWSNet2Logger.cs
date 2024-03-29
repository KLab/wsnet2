﻿using System;

namespace WSNet2
{
    /// <summary>
    /// WSNet2のためのLoggerインターフェイス
    /// </summary>
    /// <typeparam name="TPayload">構造化ログのためのペイロード型</typeparam>
    /// <para>
    /// ペーロードにはWSNet2LogPayloadに独自のフィールドを加える拡張をした型を指定できます。
    /// </para>
    /// <para>
    /// このインターフェイスを実装するには、最低限次のLogメソッドを実装する必要があります。
    /// </para>
    /// <list type="bullet">
    ///    <item><term>Log(WSNet2LogLevel level, Exception exception, string format, params object[] args);</term></item>
    /// </list>
    /// <para>
    /// この他のオーバーロードメソッドはパフォーマンス対策などのために適宜実装してください。
    /// </para>
    public interface IWSNet2Logger<out TPayload> where TPayload : WSNet2LogPayload
    {
        /// <summary>構造化ログのためのペイロード</summary>
        TPayload Payload { get; }

        /// <summary>ログ出力メソッド (必須)</summary>
        void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args);

        // Unityの場合C#8.0がサポートされるまで無効
#if !UNITY_2 && !UNITY_3 && !UNITY_4 && !UNITY_5 && !UNITY_5_3_OR_NEWER || CSHARP_8_0_OR_NEWER

        void Log(WSNet2LogLevel logLevel, Exception exception, string message)
            => Log(logLevel, exception, message, empty);
        void Log<T1>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1)
            => Log(logLevel, exception, format, (object)a1);
        void Log<T1, T2>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2)
            => Log(logLevel, exception, format, (object)a1, (object)a2);
        void Log<T1, T2, T3>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3)
            => Log(logLevel, exception, format, (object)a1, (object)a2, (object)a3);
        void Log<T1, T2, T3, T4>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4)
            => Log(logLevel, exception, format, (object)a1, (object)a2, (object)a3, (object)a4);
        void Log<T1, T2, T3, T4, T5>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5)
            => Log(logLevel, exception, format, (object)a1, (object)a2, (object)a3, (object)a4, (object)a5);

        private static object[] empty = new object[] { };
#endif
    }

    /// <summary>LogLevel</summary>
    /// <remarks>Same of Microsoft.Extensions.Logging.LogLevel</remarks>
    public enum WSNet2LogLevel
    {
        Trace = 0,
        Debug = 1,
        Information = 2,
        Warning = 3,
        Error = 4,
        Critical = 5,
        None = 6,
    }

    /// <summary>
    /// 構造化ログのためのペイロード
    /// </summary>
    /// <para>
    /// フィールドを追加したい場合は継承する
    /// </para>
    public class WSNet2LogPayload
    {
        /// <summary>WSNet2のAppId</summary>
        public string AppId { get; set; }

        /// <summary>WSNet2のユーザID</summary>
        public string UserId { get; set; }

        /// <summary>部屋のID</summary>
        public string RoomId { get; set; }

        /// <summary>部屋番号</summary>
        public int RoomNum { get; set; }
    }
}
