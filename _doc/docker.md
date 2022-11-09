## 立ち上げ方

ロカール環境にwsnet2のサーバ群を立ち上げるには、Dockerとdocker-composeを利用します。

[リポジトリ](https://github.jp.klab.com/WSNet/wsnet2)の`server`ディレクトリで`docker-compose build`と`docker-compose up`を叩いてください。

```BASH
$ git clone git@github.jp.klab.com:WSNet/wsnet2
$ cd wsnet2/server
$ docker-compose build
$ docker-compose up
```

## 接続先情報

- LobbyURL: http://localhost:8080
- AppID: testapp
- AppKey: testapppkey

Unityから接続するための`WSNet2Client`は次のように取得します。

```C#
var userId = "user1";
var authgen = new AuthDataGenerator();

var client = WSNet2Service.Instance.GetClient(
                 "http://localhost:8080",
                 "testapp",
                 userId,
                 authgen.Generate("testapppkey", userId));
```

この`testapp`は初期状態で登録されています。
追加するには、DBの`app`テーブルにレコードを追加してください。
追加後、lobby、game、hubを再起動することで反映されます。

DBへのアクセス方法：

```BASH
$ docker exec -it wsnet2-db mysql -uwsnet -pwsnetpass wsnet2
```

## コンテナ一覧

- wsnet2-builder
  - wsnet2のバイナリをビルドする
- wsnet2-lobby
  - Lobbyサーバ
  - クライアントからの部屋の作成や入室、部屋検索リクエストを受け付ける
- wsnet2-game
  - Gameサーバ
  - 部屋を保持、クライアントからのwebsocket接続を受け付けてメッセージを送受信する
- wsnet2-hub
  - 観戦Hubサーバ
  - Gameからのメッセージを観戦クライアントに中継する
- wsnet2-db
  - データベース
  - 部屋の検索などに利用する