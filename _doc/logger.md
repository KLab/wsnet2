# WSNet2のLogger

- [概要](#概要)
- [IWSNet2Loggerの実装](#IWSnet2Loggerの実装)
  - [Payloadプロパティ](#Payloadプロパティ)
  - [Logメソッド](#Logメソッド)
  - [オプショナルなLogメソッド](#オプショナルなLogメソッド)
- [Loggerの登録](#Loggerの登録)
- [例](#例)
  - [dotnetコンソールアプリケーションでZLoggerを使う](#dotnetコンソールアプリケーションでZLoggerを使う)

## 概要

WSNet2内でのログを受け取るには、WSNet2Clientを通してLoggerを設定します。
Loggerは`IWSNet2Logger`インターフェイスを実装する必要があります。

このLoggerインターフェイスはZLoggerなどの構造化ログにも対応できるようにPayloadを備えています。

## IWSNet2Loggerの実装

IWSNet2Loggerを実装するクラスでは、最低限`Payload`プロパティと可変長引数をもつ`Log`メソッドを実装する必要があります。

```C#
public interface IWSNet2Logger<out TPayload> where TPayload : WSNet2LogPayload
{
    /// <summary>構造化ログのためのペイロード</summary>
    TPayload Payload { get; }

    /// <summary>ログ出力メソッド (必須)</summary>
    void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args);

    ...
}
```

### Payloadプロパティ

```C#
TPayload Payload { get; }
```

構造化ログのためのPayloadです。
`WSNet2LogPayload`クラスで定義されているように、アプリID、ユーザID、部屋のIDと部屋番号が含まれます。
これらの値はLoggerの設定時や部屋への入室時に自動的に書き込まれます。
このため、同じLoggerインスタンスを異なる部屋のLoggerに設定してしまうと、上書きされてしまい期待通りのログが得られなくなるので注意してください。

```C#
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
```

このPayloadをどのように出力するかは`Log`メソッドの実装に任されます。

利用者側でPayloadにフィールドを追加するには、`WSNet2LogPayload`を継承したクラスを`IWSNet2Logger`の`TPayload`に指定します。
具体的には[例](#例)を見てください。

### Logメソッド

```C#
void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args);
```

ログ出力時に呼ばれるメソッドです。
`logLevel`引数は`Microsoft.Extensions.Logging.LogLevel`と同等の列挙型です。
ログレベルに応じて適した出力をしてください。

ログに付随する例外がある場合は`exception`引数に例外のインスタンスが渡されます。
例外と関連しないログではこの引数は`null`です。

### オプショナルなLogメソッド

ZLoggerのようにパラメータのBoxingを回避するためのオーバーロードメソッドを用意しています。
デフォルト実装があるため、明示的に実装しなくても構いません。
（このオーバーロードはUnityでは利用できません）

```C#
void Log(WSNet2LogLevel logLevel, Exception exception, string message);

void Log<T1>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1);

void Log<T1, T2>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2);

void Log<T1, T2, T3>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3);

void Log<T1, T2, T3, T4>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4);

void Log<T1, T2, T3, T4, T5>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5);
```

## Loggerの登録

Loggerの登録はWSNet2Clientのコンストラクタ、CreateやJoinメソッドの引数で行います。

コンストラクタで登録したLoggerは入室前のリクエストや部屋検索についてのログ出力に使われます。
一方、CreateやJoin等のメソッドの引数で登録したものは、その部屋についてのログ出力に使われます。

Unityで[`WSNet2Service`](../wsnet2-unity/Assets/WSNet2/Scripts/WSNet2Service.cs)を利用する場合、`GetClient`メソッドの引数で指定したものがWSNet2Clientコンストラクタの引数になります。
ここで指定を省略した場合、`UnityEngine.Debug`を利用する[`DefaultUnityLogger`](../wsnet2-unity/Assets/WSNet2/Scripts/DefaultUnityLogger.cs)が使われます。

`Create`、`Join`、`RandomJoin`、`Watch`で登録したLoggerは、それで入室した部屋についてのログ出力に使われます。
この時`Payload`の部屋情報などは上書きされるため、Loggerを使い回すのは避けてください。

## 例

### dotnetコンソールアプリケーションでZLoggerを使う

IWSNet2Loggerの実装としてZLoggerをラップし、独自に拡張したPayloadをもつLoggerクラスを定義します。

```C#
using Microsoft.Extensions.Logging;
using ZLogger;
using WSNet2;

// 独自のLoggerクラス.
// 拡張したPayloadをもつIWSNet2Loggerインターフェイスを実装している
public class Logger : IWSNet2Logger<Logger.LogPayload>
{
    // Payloadとなるクラス
    // WSNet2LogPayloadを継承し、独自のフィールド(AdditionalData)を追加している
    public class LogPayload : WSNet2LogPayload
    {
        public int AdditionalData { get; set; }
    }

    // Payloadプロパティの実装
    // WSNet2LogPayloadを継承した型であればインターフェイスを満たせる
    public LogPayload Payload { get; } = new LogPayload();

    // ZLoggerのインスタンスを保持する
    ILogger logger;

    public Logger(ILogger zlogger)
    {
        loggr = zlogger;
    }

    // 必須のLogメソッドの実装
    // シンプルにZLoggerに渡す
    public void Log(WSNet2LogLevel logLevel, Exception exception, string format, params object[] args)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, string.Format(format, args));
    }

    // 以下、Boxing回避のためのメソッドの実装
    // ZLoggerのジェネリックメソッドによるBoxing回避をそのまま利用できる

    public void Log(WSNet2LogLevel logLevel, Exception exception, string message)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, message);
    }
    public void Log<T1>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1)
    }
    public void Log<T1, T2>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2);
    }
    public void Log<T1, T2, T3>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2, a3);
    }
    public void Log<T1, T2, T3, T4>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2, a3, a4);
    }
    public void Log<T1, T2, T3, T4, T5>(WSNet2LogLevel logLevel, Exception exception, string format, T1 a1, T2 a2, T3 a3, T4 a4, T5 a5)
    {
        logger.ZLogWithPayload((LogLevel)logLevel, exception, Payload, format, a1, a2, a3, a4, a5);
    }
}
```

この`Logger`を次のように利用します。

```C#
    // ZLoggerのインスタンスを用意
    using var loggerFactory = LoggerFactory.Create(builder =>
    {
        builder.ClearProviders();
        builder.SetMinimumLevel(LogLevel.Debug);
        builder.AddZLoggerConsole(options => options.EnableStructuredLogging = true);
    });
    var zlogger = loggerFactory.CreateLogger("WSNet2");

    // LoggerをWSNet2Clientに登録
    var logger = new Logger(zlogger);
    var wsclient = new WSNet2Client(lobbyUrl, appId, myId, authData, logger);

    ...

    // 部屋用のLoggerを新しく作り、RandomJoin時に登録
    var roomLogger = new Logger(zlogger);
    wsclient.RandomJoin(
        0, null, null,
        room => {
            // Payloadの独自フィールドは自由に書き換える
            roomLogger.Payload.AdditionalData = 42;
            ...
        },
        exception => {
            ...
        },
        roomLogger);
```
