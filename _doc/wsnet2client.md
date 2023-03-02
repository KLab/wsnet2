# WSNet2Client クラス

## 目次

- [概要](#概要)
- [コンストラクタ](#コンストラクタ)
- [部屋作成](#部屋作成)
  - [Create](#Create)
- [入室](#入室)
  - [Join(roomId)](#joinroomid)
  - [Join(number)](#joinnumber)
  - [RandomJoin](#randomjoin)
- [観戦](#観戦)
  - [Watch(roomId)](#watchroomid)
  - [Watch(number)](#watchnumber)
- [部屋検索](#部屋検索)
  - [Search](#search)

## 概要
WSNet2へのアクセスは、[`WSNet2.WSNet2Client`](../wsnet2-unity/Assets/WSNet2/Scripts/Core/WSNet2Client.cs)クラスを利用します。
このクラスを使ってRoomの作成、入室、観戦、検索を行います。

各メソッドで指定する`logger`、`roomLogger`については[`WSNet2のLogger`](logger.md)を参照してください。

## コンストラクタ

```C#
WSNet2Client(string baseUri, string appId, string userId, string authData, IWSNet2Logger<WSNet2LogPayload> logger);
```

- `baseUri`: WSNet2のLobbyのURLのベース
- `appId`: Wsnetに登録してあるApplication ID
- `userId`: プレイヤーIDとなるID
- `authData`: 認証情報（ゲームAPIサーバから入手）
- `logger`: クライアント用のLogger。部屋には引き継がれない。

## 部屋作成
### Create
```C#
void Create(
    RoomOption roomOption,
    IDictionary<string, object> clientProps,
    Action<Room> onSuccess,
    Action<Exception> onFailed,
    IWSNet2Logger<WSNet2LogPayload> roomLogger);
```

Roomを作成して入室します。

- `roomOption`: Roomのプロパティなど。[RoomOption](roomoption.md)を参照。
- `clientProps`: プレイヤー（自分自身）のカスタムプロパティ
- `onSuccess`: 成功時コールバック。引数は作成したRoom
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト。
- `roomLogger`: 部屋用のLogger。他の部屋で使ったものを使い回すとPayloadが上書きされてしまうため注意

RoomOptionの例：
最大人数2人、検索不可、観戦不可、部屋番号付きで指定の公開カスタムプロパティを持つ部屋を作成する

```C#
var publicProps = new Dictionary<string, object>(){
    {"type", RoomType.Duel},
    {"state", RoomState.Waiting},
};
var roomOpt = new RoomOption(2, 1, publicProps, nil).Visible(false).Watchable(false).WithNumber(true);

wscli.CreateRoom(roomOpt, new Dictionary<string, object>(){{"name", "player1"}}, onSuccess, onFailed, logger);
```

## 入室

### Join(roomId)

```C#
void Join(
    string roomId,
    Query query,
    IDictionary<string, object> clientProps,
    Action<Room> onSuccess,
    Action<Exception> onFailed,
    IWSNet2Logger<WSNet2LogPayload> roomLogger);
```

RoomIDを指定して、プレイヤーとして入室します。

- `roomId`: Room ID
- `query`: 部屋条件クエリ
- `clientProps`:  プレイヤー（自分自身）のカスタムプロパティ
- `onSuccess`: 成功時コールバック。引数は入室したRoom。
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト
- `roomLogger`: 部屋用のLogger。他の部屋で使ったものを使い回すとPayloadが上書きされてしまうため注意

### Join(number)

```C#
void Join(
    int number,
    Query query,
    IDictionary<string, object> clientProps,
    Action<Room> onSuccess,
    Action<Exception> onFailed,
    IWSNet2Logger<WSNet2LogPayload> roomLogger);
```

部屋番号を指定して、プレイヤーとして入室します。

- `number`: 部屋番号
- `query`: 部屋条件クエリ
- `clientProps`:  プレイヤー（自分自身）のカスタムプロパティ
- `onSuccess`: 成功時コールバック。引数は入室したRoom。
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト
- `roomLogger`: 部屋用のLogger。他の部屋で使ったものを使い回すとPayloadが上書きされてしまうため注意

### RandomJoin

条件に合うRoomの中からランダムに、プレイヤーとして入室します。

```
void RandomJoin(
    uint group,
    Query query,
    IDictionary<string, object> clientProps,
    Action<Room> onSuccess,
    Action<Exception> onFailed,
    IWSNet2Logger<WSNet2LogPayload> roomLogger);
```

- `group`: 検索グループ
- `query`: 部屋条件クエリ
- `clientProps`:  プレイヤー（自分自身）のカスタムプロパティ
- `onSuccess`: 成功時コールバック。引数は入室したRoom。
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト
- `roomLogger`: 部屋用のLogger。他の部屋で使ったものを使い回すとPayloadが上書きされてしまうため注意

## 観戦

### Watch(roomId)

```C#
void Watch(
    string roomId,
    Query query,
    Action<Room> onSuccess,
    Action<Exception> onFailed,
    IWSNet2Logger<WSNet2LogPayload> roomLogger);
```

RoomIDを指定して、観戦入室します。

- `roomId`: Room ID
- `query`: 部屋条件クエリ
- `onSuccess`: 成功時コールバック。引数は入室したRoom。
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト。
- `roomLogger`: 部屋用のLogger。他の部屋で使ったものを使い回すとPayloadが上書きされてしまうため注意

### Watch(number)

部屋番号を指定して、観戦入室します。

```C#
void Watch(
    int number,
    Query query,
    Action<Room> onSuccess,
    Action<Exception> onFailed,
    IWSNet2Logger<WSNet2LogPayload> roomLogger);
```

- `number`: 部屋番号
- `query`: 部屋条件クエリ
- `onSuccess`: 成功時コールバック。引数は入室したRoom。
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト。
- `roomLogger`: 部屋用のLogger。他の部屋で使ったものを使い回すとPayloadが上書きされてしまうため注意

## 部屋検索

### Search(group/roomIds/roomNumbers)

```C#
void Search(
    uint group,
    Query query,
    int limit,
    Action<PublicRoom[]> onSuccess,
    Action<Exception> onFailed);

void Search(
    uint group,
    Query query,
    int limit,
    bool checkJoinable,
    bool checkWatchable,
    Action<PublicRoom[]> onSuccess,
    Action<Exception> onFailed);

void Search(
    string[] roomIds,
    Query query,
    Action<PublicRoom[]> onSuccess,
    Action<Exception> onFailed);

void Search(
    int[] roomNumbers,
    Query query,
    Action<PublicRoom[]> onSuccess,
    Action<Exception> onFailed);
```

条件に合うRoom一覧を取得します。

- `group`: 検索グループ
- `roomIds`: 部屋IDリスト
- `roomNumbers`: 部屋番号リスト
- `query`: 部屋条件クエリ
- `limit`: 件数上限
- `checkJoinable`: true: 入室可能な部屋のみ含める
- `checkWatchable`: true: 観戦可能な部屋のみ含める
- `onSuccess`: 成功時コールバック。引数はRoom一覧
- `onFailed`: 失敗時コールバック。引数は例外オブジェクト
