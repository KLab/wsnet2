## 目次

- [概要](#概要)
- [部屋のプロパティ](#部屋のプロパティ)
  - [IdとNumber](#IdとNumber)
  - [検索や入室に関わるフラグ](#検索や入室に関わるフラグ)
  - [SearchGroup](#SearchGroup)
  - [公開プロパティと非公開プロパティ](#公開プロパティと非公開プロパティ)
  - [クライアントの情報](#クライアントの情報)
  - [その他の情報](#その他の情報)
- [イベントレシーバ](#イベントレシーバ)
  - [OnJoined](#onjoined)
  - [OnClosed](#onclosed)
  - [OnOtherPlayerJoined, OnOtherPlayerLeft](#onotherplayerjoined-onotherplayerleft)
  - [OnMasterPlayerSwitched](#onmasterplayerswitched)
  - [OnRoomPropertyChanged](#onroompropertychanged)
  - [OnPlayerPropertyChanged](#onplayerpropertychanged)
  - [OnError, OnErrorClosed](#onerror-onerrorclosed)
- [RPC](#rpc)
  - [RPC対象メソッドのシグネチャ](#rpc対象メソッドのシグネチャ)
  - [RPCの登録](#rpcの登録)
  - [RPCの呼び出し](#rpcの呼び出し)
- [その他の操作](#その他の操作)
  - [受信イベント処理の一時停止](#受信イベント処理の一時停止)
  - [プロパティの変更](#プロパティの変更)
  - [マスターの交代](#マスターの交代)
  - [退室](#退室)
  - [Kick](#kick)

## 概要

ここでは、[`Room`](/WSNet/wsnet2/blob/master/wsnet2-unity/Assets/WSNet2/Scripts/Core/WSNet2Client.cs)オブジェクトの使い方を説明します。

プレイヤーまたは観戦者がWSNet2の部屋へ入室したとき、`onSuccess`コールバックで`Room`オブジェクトが渡されます。
WSNet2での通信を受け取るには、この`Room`にイベントレシーバやRPCを登録します。

イベントの処理は`onSuccess`が完了してからはじまります。
イベントの取りこぼしを防ぐためには、この中でイベントレシーバとRPCの登録を済ませるか、
`onSuccess`の中で`room.Pause()`を呼んでイベント処理を一時停止しておき、
登録が済んでから`room.Restart()`を呼ぶようにしてください。

## 部屋のプロパティ

### IdとNumber

どちらも部屋を特定するための情報です。
Roomの`Id`は全ての部屋に割り当てられる、ある程度長い一意な文字列です。

一方`Number`は6桁程度の数値で、部屋の作成時に`RoomOption.WithNumber(true)`にすると割り当てられます。
この番号は同時に存在する部屋と重複することはありません。
桁数はWSNet2のインフラ構築時に設定できます。

### 検索や入室に関わるフラグ

#### Visible

検索可能フラグ。
`false`の部屋は検索結果に含まれず、ランダム入室の対象にはなりません。

#### Joinable

プレイヤーとしての入室可能フラグ。
`false`の部屋には入室（Join, RandomJoin）できません。
一方`true`であれば、`Visible`の値に関わらず`Id`または`Number`で部屋を特定して入室（Join）することができます。

#### Watchable

観戦可能フラグ。
`false`の部屋は観戦（Watch）できません。

### SearchGroup

検索グループ。
サーバ側負荷軽減のため、部屋の検索（ランダム入室のための検索も含む）は検索グループ単位で行います。
検索グループの割り当てルールはプロジェクト側で設計する必要があります。
検索するときに同時に存在する同じグループの部屋数が1000件以下になるようにしてください。

### 公開プロパティと非公開プロパティ

`PublicProps`と`PrivateProps`はどちらも辞書（`Dictionary<string, object>`）型で、
キーは任意のUTF-8文字列、値は[シリアライズ可能な型](シリアライズ可能な型)です。
どちらも部屋にいる全てのクライアント（プレイヤーと観戦者）に共有されます。

加えて、`PublicProps`は部屋の検索結果にも含まれるので入室していないクライアントからも参照できるほか、
[`Query`](Queryの使い方)によるフィルタリングにも使われます。

### クライアントの情報

#### Me

自分自身を表すPlayerです。
観戦者では`null`です。

#### Master

現在のマスタープレイヤーです。

#### Players

部屋にいる全プレイヤーです。

#### Watchers

観戦人数です。
同期にはタイムラグがあります。

### その他の情報

#### ClientDeadline

最後のメッセージからこの時間経過しても次のメッセージがサーバに届かなかった場合、
クライアントは切断しているとサーバ側で判断し、退室させます。

WSNet2のC#実装側で自動的にこの値未満の間隔でPingメッセージを送信しています。

#### RttMillsec

直前のPing-Pong応答にかかった時間（ミリ秒）です。
自分自身の回線状況の参考にしてください。

#### LastMsgTimestamps

サーバが各プレイヤーから受取った最後のメッセージの受信時刻です。

他のプレイヤーの接続状況の参考にできますが、
自分自身の接続状況が悪いとそもそも更新されない点には注意してください。

## イベントレシーバ

`Room`には次のイベントレシーバのデリゲートが用意されています。

このうち`OnClosed`と`OnErrorClosed`は、部屋から退室して接続が終了したという通知になります。
もう一度同じ部屋に入りたい場合は、新しいクライアントとして、`Join`または`Watch`から始める必要があります。

### OnJoined
```C#
void OnJoined(Player me);
```

自分自身の入室イベントです。入室して最初に届きます。
引数は自分自身を表す`Player`オブジェクトです。
これは`room.Me`としてもアクセスできます。

### OnClosed
```C#
void OnClosed(string message);
```

自分自身の退室によって部屋から完全に切断したときに呼ばれます。
マスターにKickされたときも、`OnClosed`が呼ばれます。

### OnOtherPlayerJoined, OnOtherPlayerLeft
```C#
void OnOtherPlayerJoined(Player player);
void OnOtherPlayerLeft(Player player);
```

他のプレイヤーの入室と退室のイベントです。観戦者では呼ばれません。
引数は入退室したプレイヤーの`Player`オブジェクトで、
退室するまでは`room.Players[player.Id]`でもアクセスできます。

### OnMasterPlayerSwitched
```C#
void OnMasterPlayerSwitched(Player previousMaster, Player newMaster);
```

マスタープレイヤーが変更されたイベントです。
これが呼ばれた時点ですでに`room.Master`は新マスターになっています。

マスタープレイヤーが退室するときには、サーバ側で新しいマスターを選出して退室と合わせて通知します。
`OnOtherPlayerLeft`の呼び出しより前に新マスターが設定されるので、マスターが不在になることはありません。

### OnRoomPropertyChanged
```C#
void OnRoomPropertyChanged(
    bool? visible, bool? joinable, bool? watchable, 
    uint? searchGroup, uint? maxPlayers, uint? clientDeadline,
    Dictionary<string, object> publicProps, Dictionary<string, object> privateProps);
```

部屋のプロパティが変更されたイベントです。
変更された引数のみ値が入り、変更されなかったものは`null`になっています。
`room`オブジェクトのプロパティは既に変更されています。
変更前の値が必要な場合は、実装者側で事前に保存しておく必要があります。

### OnPlayerPropertyChanged
```C#
void OnPlayerPropertyChanged(Player player, Dictionary<string, object> props);
```

プレイヤーのプロパティが変更されたイベントです。
引数`props`には、変更されたキーのみ含まれます。

### OnError, OnErrorClosed
```C#
void OnError(Exception exception);
void OnErrorClosed(Exception exception);
```

エラーを通知するイベントです。
`OnError`は呼ばれても接続している状態を維持しています。
一方`OnErrorClosed`が呼ばれたときにはもう退室しています。

引数はWSNet2内部で発生した例外です。
メインスレッド以外で発生することもあるので、そのままスローせずに保持し、引数として渡しています。


## RPC

WSNet2でのメッセージの送受信は基本的にはRPC（Remote Procedure Call）を利用します。
RPCを呼び出すには、最初に対象となるメソッドを`RegisterRPC()`で`Room`に登録します。
これは全てのクライアント（プレイヤーと観戦者）で同じ順序で登録されている必要があります。

### RPC対象メソッドのシグネチャ

```C#
void TargetMethod(string senderId);
void TargetMethod(string senderId, T param);
```

RPC対象メソッドの第一引数は呼び出したプレイヤーのIDです。
さらに第二引数をとることもでき、RPC呼び出し時にパラメータを渡すことができます。
第二引数の型`T`は[シリアライズ可能な型](シリアライズ可能な型)でなければなりません。

### RPCの登録

```C#
int RegisterRPC(Action<string> rpc);
int RegisterRPC(Action<string, T> rpc);
int RegisterRPC(Action<string, T> rpc, T cacheObject = null);
```

RPC対象メソッドを`RegisterRPC()`で登録することで、呼び出し可能になります。
RPCの第二引数がプリミティブ型（真偽値、数値、文字列）**以外**のときは`cacheObject`を割り当てることができ、
実行時の引数となるオブジェクトを使い回してメモリアロケーションを抑制できます。

RPCは登録順にIDが割り当てられ、そのIDによって対象メソッドを識別します。
このため、クライアント間で登録順が異なってしまうと正しくRPCを呼び出せなくなります。

C#のオーバーロード解決が一部不完全なため、明示的な型変換や型引数の指定が必要な場合があります。

### RPCの呼び出し

```C#
int RPC(Action<string> rpc, params string[] targets);
int RPC(Action<string, T> rpc, T param, params string[] targets);
```

第一引数`rpc`には、登録したメソッドを渡します。
対象メソッドに第二引数がある場合は、引数`param`を渡し、これがそのまま登録したメソッドの第二引数になります。

引数`targets`には、RPCを実行するプレイヤーを列挙します。
なにも指定しない場合は、観戦者を含む全クライアントで登録したメソッドが実行されます。
また、`Room.RPCToMaster`を指定することで、その時点のマスターで実行されます。

RPCの呼び出しは全てのクライアント（プレイヤーと観戦者）でできます。

## その他の操作

### 受信イベント処理の一時停止

```C#
void Pause();
void Restart();
```

`Pause()`を呼び出すことで、受信イベントの処理を一時停止することができます。
処理を再開するには`Restart()`を呼び出します。
その間に受信したイベントはバッファに蓄積され、再開後順次処理されます。

一時停止している間もPingの送信は止めないので、ClientDeadline経過による切断の心配はありません。

Unityでのシーン遷移中など、イベントを処理できない間に停止しておくことを想定しています。
他にも、`Create`や`Join`で入室した時の`onSuccess`コールバックで一時停止しておき、
イベントレシーバやRPCの登録が済んだタイミングで再開するという使い方もできます。

### プロパティの変更

#### 自分自身のプロパティ

```C#
int ChangeMyProperty(IDictionary<string, object> props, Action<EvType, IDictionary<string, object>> onErrorResponse = null);
```

`ChangeMyProperty()`で自分自身のプロパティを変更できます。
他人のプロパティは変更できません。
また、観戦者はプロパティを持たないのでこの操作はできません。

実際の変更は`OnPlayerPropertyChanged`が呼ばれるタイミングで適用されます。

引数の辞書には、追加、変更するキーのみ含めます。
含まれなかったキーの値はそのまま保持されます。
このためキーの削除はできないので、値を`null`にするなどで対処してください。

`onErrorResponse`を指定しておくと、サーバ側でのエラーの通知を受け取れます。
成功したことは`OnPlayerPropertyChanged`で確認してください。

#### 部屋のプロパティ

```C#
        public int ChangeRoomProperty(
            bool? visible = null,
            bool? joinable = null,
            bool? watchable = null,
            uint? searchGroup = null,
            uint? maxPlayers = null,
            uint? clientDeadline = null,
            IDictionary<string, object> publicProps = null,
            IDictionary<string, object> privateProps = null,
            Action<EvType,bool?,bool?,bool?,uint?,uint?,uint?,IDictionary<string,object>,IDictionary<string,object>> onErrorResponse = null);
```

部屋のマスターは`ChangeRoomProperty()`で部屋のフラグや各種プロパティを変更できます。
実際の変更は`OnRoomPropertyChanged`が呼ばれるタイミングで適用されます。

変更しない項目の引数は`null`にします。
また、`publicProps`と`privateProps`引数の辞書には、追加、変更するキーのみ含めます。
含まれなかったキーの値はそのまま保持されます。
このため辞書のキーの削除はできないので、値を`null`にするなどで対処してください。

`onErrorResponse`を指定しておくと、サーバ側でのエラーの通知を受け取れます。
成功したことは`OnRoomPropertyChanged`で確認してください。

### マスターの移譲

```C#
int SwitchMaster(Player newMaster, Action<EvType, string> onErrorResponse = null);
```

マスタープレイヤーは、他のプレイヤーにマスターを移譲できます。
実際の交代は`OnMasterPlayerSwitched`が呼ばれるタイミングで適用されます。

`onErrorResponse`を指定しておくと、サーバ側でのエラーの通知を受け取れます。
成功したことは`OnMasterPlayerSwitched`で確認してください。

### 退室

```C#
int Leave();
```

部屋から退室します。
呼び出しただけではまだ退室しておらず、`OnClosed`が呼ばれるタイミングで退室が完了します。
それまでのイベントは届き続けます。

### Kick

```C#
int Kick(Player target, Action<EvType, string> onErrorResponse = null)
```

マスタープレイヤーは、他のプレイヤーを強制退室させることができます。

`onErrorResponse`を指定しておくと、サーバ側でのエラーの通知を受け取れます。
成功したことは`OnOtherPlayerLeft`で確認してください。
