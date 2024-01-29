# C#クライアントの使い方

## 目次
- [C#プロジェクトへのインポート](#cプロジェクトへのインポート)
- [wsnet2clientの利用](#wsnet2clientの利用)
- [部屋の作成](#部屋の作成)
- [部屋の検索](#部屋の検索)
- [プレイヤーとして入室](#プレイヤーとして入室)
- [メッセージの送受信](#メッセージの送受信)
- [退室](#退室)
- [観戦](#観戦)
- [その他の操作](#その他の操作)
- [切断とエラー](#切断とエラー)

## C#プロジェクトへのインポート

WSNet2を利用するのに必要なC#コードは[`wsnet2-unity/Assets/WSNet2`](../wsnet2-unity/Assets/WSNet2)以下に置かれています。
Unityの場合はこのディレクトリを`Assets`以下にコピーしてください。

.NETアプリケーションで利用する場合は、[`wsnet2-unity/Assets/WSNet2/Scripts/Core`](../wsnet2-unity/Assets/WSNet2/Scripts/Core)以下のファイルをプロジェクト（*.csproj）に含めてください。

## `WSNet2Client`の利用

WSNet2の利用は[`WSNet2Client`クラス](wsnet2client.md)を利用したロビーへのリクエスト（部屋作成・検索・入室）から始めます。
これは、Unityでは`WSNet2Service.Instance.GetClient()`で取得できるほか、.NETアプリケーションでは直接`new WSNet2Client(...)`します。

WSNet2Clientの利用には認証情報が必要になります。
詳しくは[WSNet2のユーザ認証](user_auth.md)を参照して下さい。

`WSNet2Client`を利用して入室したあとは、`onSuccess`コールバックに渡される[`Room`](room.md)オブジェクトを通して通信します。

### スレッドプール最小値の設定

WSNet2Client利用時、スレッドプールの最小値を設定していない場合に次のメッセージの警告が表示されることがあります。

```
The ThreadPool may become congested. Please consider increasing the ThreadPool.MinThreads values.
c.f. https://gist.github.com/JonCole/e65411214030f0d823cb
WorkerThreads: 8, CompletionPortThreads: 8
```

このような場合、次のメソッドで適切な値を設定してください。

```C#
ThreadPool.SetMinThreads(200, 200);
```

大きな値を設定してもすぐにスレッドが作られるわけではないので、やや大きめにしても大丈夫です。
スマホクライアントの場合10以上を目安に、MasterClientのようにサーバサイドで動かす場合は100を目安に、プロファイラ等を見ながら決めてください。

## 部屋の作成

`WSNet2Client.Create()`メソッドで部屋を作成し入室します。

作成する部屋の属性は[`RoomOption`](roomoption.md)で指定します。
`RoomOption`には制限人数やフラグ、プロパティを設定します。
自分自身のプロパティは`clientProps`引数で指定します。

部屋の作成が成功すると、`onSuccess`コールバックが呼ばれます。
この時点で入室まで完了しています。
部屋を作成したプレイヤーは、部屋のMasterになっています。

入室した部屋情報は`onSuccess`の引数`room`として渡されます。
まずはこの`room`にイベントレシーバやRPCを登録します。

### イベントレシーバ/RPCの登録

`onSuccess`の処理中に受信したイベントはプールに溜められ、処理されません。
つまり、`onSuccess`の中で全てのレシーバ/RPCを設定できれば、イベントを取りこぼすことはありません。

Unityでシーン遷移してからレシーバを設定したい場合などのために、
`room.Pause()`でイベント処理を一時停止できます。
レシーバを設定した後で`room.Restart()`して再開します。
一時停止中も`onSuccess`と同様、受信したイベントはプールに溜められます。
プールのサイズは[WSNet2Settings.EvPoolSize](wsnet2settings.md#EvPoolSize)で設定できます。

### 例

```C#
// WSNet2Clientを取得
var client = WSNet2Service.Instance.GetClient(baseUri, appId, userId, authData);

// 部屋の属性とプロパティ
var maxPlayers = 4;
var searchGroup = 1;
var publicProps = new Dictionary<string, object>(){
    {"public_key1", "some string"},
    {"public_key2", (int)30},
};
var privateProps = new Dictionary<string, object>(){
    {"private_key1", 0.8f},
    {"private_key2", "privedata"},
};
var roomOption = new RoomOption(maxPlayers, searchGroup, publicProps, privateProps);
roomOption.WithNumber(true); // 部屋番号つきの部屋を作成

// 自分のプロパティ
var playerProps = new Dictionary<string, object>(){
    {"name", "player1"},
};

// 部屋を作成して入室
client.Create(
    roomOption,
    playerProps,
    (room) =>
    {
        Debug.Log("room created: " + room.Id);
        this.room = room;

        // RPCやレシーバを登録
        room.OnErrorClosed += (e) => Debug.LogError(e.ToString());
        room.OnClosed += (msg) => Debug.Log($"closed: {msg}");
        room.OnOtherPlayerJoined += (p) => Debug.Log($"new player: {p.Id}");

        // RPCは全てのクライアントで同じ順番で登録する
        room.RegisterRPC(RPCGameState);
        room.RegisterRPC<GameMessage>(RPCGameMessage);
    },
    (exception) =>
    {
        Debug.LogError($"create room failed: {exception}");
    });

```

## 部屋の検索

`WSNet2Client.Search()`メソッドで、現在存在する部屋を検索できます。
負荷軽減のため`searchGroup`単位で検索します。
また、[`Query`](query.md)により部屋の公開プロパティによるフィルタリングができます。

### 例

```C#
// 検索条件
var searchGroup = 1;
var limit = 10;

// publicPropsでのフィルタリング
var query = new Query();
query.Equal("public_key1", "some string");
query.Between("public_key2", 10, 50);

// 部屋の検索
client.Search(
    searchGroup,
    query,
    limit,
    (rooms) =>
    {
        // 部屋一覧をroomsとして受け取れる
        foreach(var room in rooms)
        {
            Debug.Log($"room: {room.Id}");
        }
    },
    (exception) =>
    {
        Debug.LogError($"search room failed: {exception}");
    });
```

## プレイヤーとして入室

既存の部屋へプレイヤーとして入室するには、
roomIdやroomNumberで部屋を指定して入室する`WSNet2Client.Join()`、
または検索条件に合致する部屋へランダム入室する`WSNet2Client.RandomJoin()`を利用します。

入室するには部屋が`joinable=true`である必要があります。
加えて、RandomJoinの対象になるのは`visible=true`の部屋のみです。
また、roomIdやroomNumberで部屋を特定した場合でも、Queryの条件にも合致しないと入室できません。

入室が成功すると`onSuccess`コールバックが呼ばれ、roomが渡されます。
この時点で、他のプレイヤーにも入室したことが通知されます。

部屋作成と同様、イベントハンドラやRPCを`onSuccess`で登録するか、
`Room.Pause()`して登録完了してから`Room.Restart()`します。

### 例

```C#
// 部屋番号による指定
int roomNumber = 123456;
var query = new Query().Between("public_key2", 10, 50);

// 自分のプロパティ
var playerProps = new Dictionary<string, object>(){
    {"name", "player2"},
};

client.Join(
    roomNumber,
    query,
    playerProps,
    (room) =>
    {
        Debug.Log("room created: " + room.Id);

        // イベント処理を一時停止
        room.Pause();

        // roomを保存
        GameScene.Room = room;

        // シーン遷移後にレシーバを設定してroom.Restart()する
        SceneManager.LoadScene(GameScene.Name);
    },
    (exception) =>
    {
        Debug.LogError($"create room failed: {exception}");
    });
```

## メッセージの送受信

メッセージの送受信は基本的にはRPC（Remote Procedure Call）の形で行います。
RPCで呼び出すメソッドは[`Room.RegisterRPC()`](room.md#rpcの登録)によって登録します。
この登録内容は順序も含め、部屋に参加する全てのクライアントで一致している必要があります。

登録できるメソッドは、第一引数に送信者のPlayerID文字列、第二引数にRPCのパラメータ（ない場合もある）を取るActionです。
パラメータの型は、プリミティブ型または`IWSNet2Serializable`を実装した型、それらの配列型、辞書型(IDictionary<string, object>)です。

RPCの呼び出しは`RPC`メソッドで、対象のPlayerIDを指定して送信します。
対象を指定しない場合は全てのクライアント（観戦者含む）に送信されます。
また、`Room.RPCToMaster`を指定することで、その時点のMasterを対象とすることができます。

メッセージ順序の一貫性を保つため、自分自身のみが対象であっても、必ずWSNet2サーバを経由して実行されます。

### 例


```C#
class GameMessage : IWSNet2Serializable
{
    ...省略...
}

void RPCGameState(string senderId, int state)
{
    ...省略...
}

void RPCGameMessage(string senderId, GameMessage message)
{
    ...省略...
}

...

    client.Create(
        roomOption,
        playerProps,
        (room) =>
        {
            this.room = room;

            ...省略...

            // RPCは全てのクライアントで同じ順番で登録する
            room.RegisterRPC(RPCGameState);
            room.RegisterRPC<GameMessage>(RPCGameMessage);
        },
        (exception) => Debug.LogError($"create room failed: {exception}"));

...

    // 全てのクライアントで RPCGameState("...", 10); が呼び出される
    room.RPC(RPCGameState, 10);

    // Masterでのみ RPCGamemessage("...", message); が呼び出される
    var message = new GameMessage(...);
    room.RPC(RPCGameMessage, message, Room.RPCToMaster);
```

## 退室

`room.Leave()`メソッドで退室を要求できます。
呼び出した時点ではまだ退室はしておらず、`room.OnClosed`が呼ばれた段階で退室が完了します。

他のクライアント（観戦者含む）には`room.OnOtherPlayerLeft`として通知されます。
退室するのが部屋のMasterの場合、`room.OnOtherPlayerLeft`の前に`room.OnMasterPlayerSwitched`で新たなMasterが通知されます。
したがって、Masterが不在になることはありません。

全てのプレイヤーが退室した部屋は即座に終了し、新たに入室や観戦をすることはできません。
このとき残っていた観戦者も全て退室させられます。

## 観戦

観戦は、プレイヤーとしての入室に近いですが、次の点が異なります。

- `client.Join()`のかわりに`client.Watch()`で入室する
  - ランダム入室は無い
- 観戦者には`Player`オブジェクトが無い
  - PlayerIDが無い
  - プロパティを持たない
- 他クライアントに入退室が通知されない
- 部屋のMasterにならない
- RPCの対象として指定されない
  - PlayerIDが無いため
  - 全員を対象とするRPCだけが届く

一方、RPCは一般プレイヤー同様呼び出せます。

## その他の操作

### 自身のプロパティ変更

プレイヤー自身のプロパティは、いつでも`room.ChangeMyProperty()`で変更できます。
この引数の辞書には変更したいキーのみ含めればよく、その他のキーの値は変更されません。
このような仕様のため、一度追加したキーを削除することはできません。
`null`にすることなどで対処してください。

観戦者はプロパティを持たないため、この操作はできません。

### Masterの権限

次の操作はMasterだけができます。

* Masterの交代 `room.SwitchMaster()`
* 部屋のプロパティ変更 `room.ChangeRoomProperty()`
* プレイヤーのKick `room.Kick()`

## 切断とエラー

ネットワーク的な切断や回線の切り替えなどが起こると、自動的なwebsocket再接続とメッセージ再送が行われます。
この切断と接続は`room.OnConnectionStateChanged`で通知を受け取れますが入室状態に変化はなく、
アプリケーション側では特別なにかする必要はありません。
サーバ側でも入室状態のままで、他クライアントには通知されません。

`Leave`や`Kick`による正常な切断では`room.OnClosed`が、
何らかのエラーによる切断では`room.OnErrorClosed`が呼び出されます。
これらが呼び出されたときにはサーバ側でも退室状態となり、他クライアントにも`room.OnOtherPlayerLeft`が通知されています。

サーバ側での切断判定として、部屋ごとに指定された時間(`Room.ClientDeadline`)以上メッセージが送られない場合、
クライアントを退室として扱います。

再度入室したい場合は`client.Join()`を呼び出し、全く新しいPlayerとして入室します。

その他、退室はしないエラーは`room.OnError`によって通知されます。
この場合はそのまま入室状態としてメッセージの送受信を続けられます。
