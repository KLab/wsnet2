# WSNet2のユーザ認証

## 概要

クライアントは、認証データをゲームAPIサーバから取得し、その認証データを与えたWSNet2Clientを用いて、WSNet2へアクセスします。
WSNet2は認証データを検証することで、ユーザIDが正規のものかを確認します。

認証データには数分程度の有効期限を設けておきます（[設定ファイルのLobby.authdata_expire](server_setup.md#ファイル内容)で指定）。
期限を過ぎたときには、ゲームAPIサーバから再度取得します。

## ゲームAPIサーバの手順

### 鍵の事前交換

WSNet2と次の情報を事前に共有しておきます。
どちらも適当な長さのASCII文字列です。

* AppID: アプリ識別子
* AppKey: 鍵情報

WSNet2は複数プロジェクトの相乗りが出来るように、AppIDでプロジェクトを特定します。
また、AppKeyはクライアントへは絶対に公開しないでください。

### 認証データ取得API

クライアントからの要求を受けたら、次のように認証データを生成します。

```
app_key = 事前交換した鍵（ASCIIバイト列）
client_id = クライアントのユーザID（UTF-8バイト列）

nonce = 適当な乱数(64bit, BigEndianのバイト列)
timestamp = 現在時刻(Unix秒, 64bit, BigEndianのバイト列)

hmac = HMAC_SHA256(app_key, client_id + nonce + timestamp)

auth_data = Base64(nonce + timestamp + hmac)
```

認証データ(auth_data)は48Byte(nonce:8byte; timestamp:8byte; hmac:32byte)のデータをBase64エンコードした文字列です。
この文字列をクライアントに返します。

## クライアントの手順

### 認証データの取得

ゲームAPIサーバから、認証データ(Base64エンコードした文字列)を受け取ります。
有効期限をクライアント側でも管理し、有効な間は使いまわすこともできます。

### WSNet2Clientへの設定

WSNet2（Lobby）へのアクセスは、[`WSNet2.WSNet2Client`](../wsnet2-unity/Assets/WSNet2/Scripts/Core/WSNet2Client.cs)クラスを利用します。

Unityの場合、[`WSNet2Service.Instance.GetClient()`](../wsnet2-unity/Assets/WSNet2/Scripts/WSNet2Service.cs#L42-L53)の引数として認証データを渡します。
.NETの場合、[`WSNet2Client`のコンストラクタ](../wsnet2-unity/Assets/WSNet2/Scripts/Core/WSNet2Client.cs#L26-L34)に認証データを渡します。

### 認証データの更新

ゲームAPIサーバから新しい認証データを受取ったら、[`WSNet2Client.UpdateAuthData()`](../wsnet2-unity/Assets/WSNet2/Scripts/Core/WSNet2Client.cs#L53-L60)で上書きできます。

Unityの場合は`WSNet2Service.Instance.GetClient()`で取得し直してもかまいません。
`appId`と`clientId`が同じクライアントが既にある場合、そのインスタンスの認証データを上書きして返します。

# その他補足

## Roomアクセス時の認証

Roomへのwebsocket接続時にも同様の認証が行われますが、これはゲームAPIサーバから取得した認証データではなく、Room単位で有効なキーを利用します。
なので、入室や観戦が受理されて以降は、認証データの有効期限を気にしなくてもよいです。
Room用のキーの管理はWSNet2のC#実装側で行っているので、クライアント側で特に触れる必要はありません。

## APIサーバなしでの利用（開発用）

事前交換したAppKeyがあれば、認証情報をクライアント側でも生成できます。
これには、[`WSNet2.AuthDataGenerator`](../wsnet2-unity/Assets/WSNet2/Scripts/Core/AuthDataGenerator.cs)クラスを利用します。
