WSNet2
======

## これは何？

WSNet2はWebSocketをベースとしたモバイルオンラインゲーム向けのリアルタイム通信システムです。

### 特徴

#### WebSocketベース

クライアント・サーバ間は、HTTP(S)とWebSocketという広く使われているプロトコルで通信するため、
特別な設定をすることなくほとんどの環境から接続することができます。

#### 自動再接続・メッセージ再送機能

特にモバイル端末では、モバイル回線・Wifiの切り替え時や、移動による一時的な切断が起こります。
このときWSNet2は自動で再接続して未送受信のメッセージを同期するため、利用アプリは切断を気にせず通信を続けられます。

#### スケーラブルな観戦

ゲームの部屋を管理するサーバ（Gameサーバ）だけでなく、観戦用にメッセージ中継をするサーバ（Hubサーバ）も備えています。
Hubサーバを増やすことでGameサーバの負荷を最小限に抑えつつ大規模な観戦を実現できます。

#### Unity/.Net 両対応

クライアントの開発言語はC#をメインターゲットとしており、Unityのゲームに組み込むことができます。
また.Netでも利用できるため、サーバサイドゲームロジックをより軽量に実行することができます。

### 動作・開発環境

- サーバ
  - Go 1.18 以降
  - MySQL 5.7 以降
- クライアント
  - C#
    - Unity 2020 以降
    - .Net 5.0 以降
  - Go 1.18 以降

## 使ってみる

### サーバの起動

```shell
$ cd server
$ docker compose up
```

VMやリモート環境で起動する場合は、
`docker-compose.yml`にてgame/hubの接続ホスト名を環境変数`WSNET2_GAME_PUBLICNAME`で適切に指定してください。

### サンプルゲーム

![Titleシーン](_doc/sample_title.png)
![Gameシーン](_doc/sample_game.png)

`wsnet2-unity`ディレクトリをUnityで開き、`Assets/Sample/Title.unity`シーンを実行します。

#### Unityクライアントでゲームロジックを動かす場合

タイトルシーンで「部屋作成」ボタンを押して、部屋を作り対戦相手の入室を待ちます。
対戦相手のBotを次のように起動します。

```shell
$ cd wsnet2-dotnet/WSNet2.Sample
$ dotnet run -- -b
```
Lobbyサーバを`http://localhost:8080`以外で起動している場合は、
Titleシーン画面のLobbyに入力したうえで、Bot起動時にも`-s`オプションで指定してください。

#### サーバサイドロジックを使う場合

次のようにゲームロジッククライアント（MasterClient）とBotを起動します。
MasterClientが部屋を作り、BotとUnityクライアントの入室を待ち受けます。

```shell
$ cd wsnet2-dotnet/WSNet2.Sample
$ dotnet run -m -b
```

UnityクライアントのTitleシーンで「ランダム入室」を押してMasterClientの待ち受ける部屋に入室します。
