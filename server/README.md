# WSNet2 Server

## 概要


WSNet2 Server は3種類のサーバーを提供します。

![外部構成](./_doc/wsnet2-external.drawio.png)

<dl>
<dt>Lobbyサーバー</dt>
<dd>外部からの部屋作成リクエストに応じてRoomを作成したり、参加リクエストに応じてgame/hubサーバーのRoomにプレイヤーを登録し、接続先のアドレスを案内します。またRoomの検索やランダム入室も受け付けます。</dd>
<dt>Gameサーバー</dt>
<dd>Roomを提供します。Roomに参加しGameサーバーに接続したプレーヤーは、Room内の他のプレーヤーと通信します。
<dt>Hubサーバー</dt>
<dd>Roomを観戦するためのhubを提供します。hubが仲介することで、大量の観戦者がいてもGameサーバーで行われているゲームの性能への影響を最小限に抑えます。</dd>
</dl>

## 内部構成

![内部構成](./_doc/wsnet2-internal.drawio.png)

Lobbyサーバーは外部のプレーヤーやサーバーから、Roomの作成、検索、参加、観戦のリクエストを受けます。このAPIはHTTPベースで、メッセージ本文はmsgpackを利用しています。

内部の通信（Lobby→Game、Lobby→Hub)はgRPCを使います。

作成・参加した部屋の情報をLobbyから受取ったClientは、WebSocketでGameサーバーへ接続します。

例: 部屋の作成

```mermaid
sequenceDiagram
    actor Client
    participant Lobby
    participant Game
    participant DB

    Client->>+Lobby: CreateRoom (HTTP)
    Lobby->>Game: Create (gRPC)
    activate Game
    Game->>DB: 部屋登録
    Game-->>Lobby: JoinedRoomRes
    Lobby-->>-Client: OK
    activate Client

    Client->>Game: Connect (WebSocket)
    Client->>Game: Leave
    Game-->>Client: Closed
    deactivate Client
    Game-->>DB: 部屋削除
    deactivate Game
````

## Hubサーバーを経由した観戦

Hubは観戦クライアントにとってのGameサーバであると同時に、Gameサーバにとってのクライアントとして振る舞います。

```mermaid
sequenceDiagram
    actor Watcher
    participant Lobby
    participant Hub
    participant Game

    activate Game

    Watcher->>+Lobby: WatchRoom (HTTP)
    Lobby->>Hub: Watch (gRPC)
    Hub->>Game: Watch (gRPC)
	Game-->>Hub: JoinedRoomRes
    activate Hub
	Hub-->>Lobby: JoinedRoomRes
    Hub->>Game: Connect (WebSocket)
    Lobby-->>-Watcher: OK
    activate Watcher

    Watcher->>Hub: Connect (WebSocket)

    Game-->>Hub: Closed
    deactivate Game
    Hub-->>Watcher: Closed
    deactivate Hub
    deactivate Watcher
```
