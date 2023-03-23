# WSNet2

## これは何？

WSNet2 は WebSocket をベースとしたモバイルオンラインゲーム向けのリアルタイム通信システムです。

### 特徴

- **WebSocket ベース**: ほとんどの環境から特別な設定なく接続することができます
- **自動再接続・再送**: 利用アプリは一時的な切断を気にすることなく実装できます
- **スケーラブルな観戦**: 中継サーバ（Hub）を増やすことで大規模な観戦に対応できます
- **Unity/.Net 両対応**: C#のゲームロジックをより軽量にサーバ上でも実行できます

### 動作・開発環境

- サーバ
  - Go 1.19 以降
  - MySQL 5.7 以降
- クライアント
  - C#
    - Unity 2020 以降
    - .Net 5.0 以降
  - Go 1.19 以降

## 使ってみる

### WSNet2 サーバ群の起動

```shell
$ cd server
$ docker compose up
```

VM やリモート環境で起動する場合は、
`compose.yaml`にて game/hub の接続ホスト名を環境変数`WSNET2_GAME_PUBLICNAME`で適切に指定してください。

### サンプルゲーム

![Titleシーン](_doc/sample_title_s.png)
![Gameシーン](_doc/sample_game_s.png)

#### Unity クライアントの使い方

`wsnet2-unity`ディレクトリを Unity で開き、`Assets/Sample/Title.unity`シーンを実行します。

- テキストボックス
  - **Lobby**: WSNet2 の lobby の URL です
  - **AppID/Key**: WSNet2 に登録している AppID と Key です（Docker では`testapp`/`testapppkey`が登録されています）
  - **UserID**: WSNet2 がユーザを識別する ID です。対戦するにはお互い異なる UserID にします
- ボタン
  - **CPU 対戦**: オフラインで CPU と対戦します（WSNet2 に接続しません）
  - **部屋作成**: WSNet2 に部屋を作り、相手プレイヤーを待ち受けます
  - **ランダム入室**: プレイヤーを待っている部屋に入室します
  - **ランダム観戦**: 対戦中のゲームを観戦します

#### Unity クライアントでゲームロジックを動かす例

Unity クライアントのタイトルシーンで「部屋作成」ボタンを押して、部屋を作り対戦相手の入室を待ちます。
対戦相手の Bot を次のように起動します。

```shell
$ cd wsnet2-dotnet/WSNet2.Sample
$ dotnet run -- -b
```

Lobby サーバを`http://localhost:8080`以外で起動している場合は、
Title シーン画面の Lobby に入力したうえで、Bot 起動時にも`-s`オプションで指定してください。

#### サーバサイドゲームロジックで対戦する例

次のようにゲームロジッククライアント（MasterClient）と Bot を起動します。
MasterClient が部屋を作り、Bot と Unity クライアントの入室を待ち受けます。

```shell
$ cd wsnet2-dotnet/WSNet2.Sample
$ dotnet run -- -m -b
```

Unity クライアントの Title シーンで「ランダム入室」を押して MasterClient の待ち受ける部屋に入室します。

#### 観戦する例

次のようにゲームロジッククライアント（MasterClient）と Bot を 2 つ起動します。
MasterClient が部屋を作り、Bot が入室して対戦を始めます。

```shell
$ cd wsnet2-dotnet/WSNet2.Sample
$ dotnet run -- -m -b 2
```

Unity クライアントの Title シーンで「ランダム観戦」を押して観戦します。

## Documentations

### 使い方

- [C#クライアントの使い方](_doc/csharp_client.md)
- [Docker を使ったローカルでの起動](_doc/docker.md)
- [サーバの構築](_doc/server_setup.md)
- [ダッシュボード](wsnet2-dashboard/README-ja.md)

### 機能詳細

- [WSNet2 のユーザ認証](_doc/user_auth.md)
- [シリアライズ可能な型](_doc/serializable.md)
- [シリアライザの使い方](_doc/serializer.md)
- [WSNet2 の Logger](_doc/logger.md)

### クラス詳細

- [WSNet2Client](_doc/wsnet2client.md)
- [RoomOption](_doc/roomoption.md)
- [Room](_doc/room.md)
- [Query](_doc/query.md)
- [WSNet2Settings](_doc/wsnet2settings.md)
