# wsnet2-dashboard

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](README.md)

![demo](images/demo.gif)

wsnet2 のダッシュボード。

サーバー・部屋状況が見えるほか、詳細データ確認やプレイヤーのキックも行える。

## アプリ構成：

|                | 構成                    | リンク                     |
| -------------- | ----------------------- | -------------------------- |
| Frontend       | Vue3 + NaiveUI          | [詳細](frontend/README.md) |
| Backend（BFF） | Nodejs + Prisma + Nexus | [詳細](backend/README.md)  |

## Docker での立ち上げ手順

### wsnet2-server とまとめてローカルで立ち上げる場合

1. `frontend` の WEB アプリをビルド
   ```bash
   cd wsnet2-dashboard
   docker compose run --rm frontbuilder
   ```
2. `wsnet2-server` と `wsnet2-dashboard` を一緒に立ち上げる：
   ```bash
   docker compose -f compose.yaml -f ../server/compose.yaml up
   ```
   （`server`ディレクトリと`wsnet2-dashboard`ディレクトリのそれぞれで`docker compose up`することもできます）

3. （オプショナル）テスト用部屋生成：
   ```bash
   docker-compose exec game /repo/server/bin/wsnet2-bot --lobby="http://lobby:8080" static 3600
   ```
4. `http://localhost:8081` から `wsnet2-dashboard` の WEB アプリへアクセス

### wsnet2-server が既に別の環境で用意されている場合

1. [wsnet2-server 本体](https://github.jp.klab.com/WSNet/wsnet2/tree/master/server)が起動されていることを確認
2. `frontend/.env`、`backend/.env` に記載されている環境変数を必要に応じて編集

   - [Backend 環境変数](backend/README.md#%E7%92%B0%E5%A2%83%E5%A4%89%E6%95%B0)
   - [Frontend 環境変数](frontend/README.md#%E7%92%B0%E5%A2%83%E5%A4%89%E6%95%B0)

3. `docker-compose up` でダッシュボードを立ち上げる
4. 環境変数”FRONTEND_ORIGIN”のアドレスから `wsnet2-dashboard` の WEB アプリへアクセス

## ダッシュボード利用説明

ダッシュボードを利用する前に、”Settings”でサーバーの IP アドレスを設定する必要があります。初めて wsnet2-dashboard にアクセスする時に設定のプロンプトが出てきます（Frontend の環境変数”VITE_DEFAULT_SERVER_URI”を設定した場合はデフォルトでそのアドレスが使われるため、初期サーバー IP アドレス設定を飛ばすことが可能）。

wsnet2-dashboard のメニューアイテム：

| 名前           | 説明                                             |
| -------------- | ------------------------------------------------ |
| Home           | ホーム画面、サーバー数や部屋分布の情報が見れます |
| Apps           | wsnet2 を利用しているアプリ一覧                  |
| Game Servers   | wsnet2 のゲームサーバー一覧                      |
| Hub Servers    | wsnet2 のハブサーバー一覧                        |
| Rooms          | wsnet2 の部屋検索・情報確認（キック機能含む）    |
| Room Histories | wsnet2 の部屋履歴検索                            |
| Settings       | wsnet2-dashboard の各種設定                      |

### 部屋関連検索（Rooms / Room Histories）

各種フィルターを手動で指定することで、条件に合う部屋を検索します。
全てのフィルターがオプショナルで、任意の組み合わせが可能です。

#### 部屋検索で設定できるフィルター：

| 種類                      | 説明                                          |
| ------------------------- | --------------------------------------------- |
| Target AppIds             | 検索される AppId（複数指定可能）              |
| Visible                   | 部屋が公開されてるかどうか                    |
| Joinable                  | 部屋に参加可能かどうか                        |
| Wachable                  | 部屋を観戦可能かどうか                        |
| Host Id                   | 検索したいホストの Id（Game Server と紐づく） |
| Number                    | 検索したい部屋の番号                          |
| Max Players               | 部屋の最大プレイヤー数                        |
| Search Group              | 部屋のサーチグループ                          |
| Minimum number of players | 部屋の最少プレイヤー数                        |
| Maximum number of players | 部屋の最大プレイヤー数                        |
| Minimum number of wachers | 部屋の最少観戦者数                            |
| Maximum number of wachers | 部屋の最大観戦者数                            |
| Created (Before/After)    | 部屋が作成された時刻範囲（片方のみ指定可能）  |

#### 部屋履歴検索で設定できるフィルター：

| 種類                   | 説明                                          |
| ---------------------- | --------------------------------------------- |
| Target AppIds          | 検索される AppId（複数指定可能）              |
| Room Id                | 検索したい部屋の Id                           |
| Host Id                | 検索したいホストの Id（Game Server と紐づく） |
| Number                 | 検索したい部屋の番号                          |
| Max Players            | 部屋の最大プレイヤー数                        |
| Search Group           | 部屋のサーチグループ                          |
| Created (Before/After) | 部屋が作成された時刻範囲（片方のみ指定可能）  |
| Closed (Before/After)  | 部屋が閉じられた時刻範囲（片方のみ指定可能）  |

#### Props フィルター（共通）

部屋が持つカスタムデータ（Props）にフィルターを掛ける。

- 「＋」ボタンを押すと新しい項目を追加する事が可能。
- 指定できる value タイプ：

  | 種類   | 例               |
  | ------ | ---------------- |
  | 文字列 | "abc"            |
  | 数字   | "123", "1.1"     |
  | ブール | "true" / "false" |

- 入れ子構造を持つ Props の場合、親の Key のチェーンで指定可能：

  ```javascript
  (例)
  {
      a: {
          aa: "value",
          bb: {
              aaa: "value",
              bbb: "value"
          }
      }
  }

  bbbにフィルターを掛ける場合は
  a/bb/bbb : value
  ```

#### ボタン説明

| 名前    | 説明                                         |
| ------- | -------------------------------------------- |
| Reset   | 設定したフィルターを全部クリアする           |
| Apply   | （キャッシュされたデータに）フィルターを適用 |
| Refresh | （キャッシュを使わずに）フィルターを適用     |

#### 部屋詳細画面

部屋データの左側にある「👁‍🗨」ボタンをクリックすると部屋詳細画面が表示されます。
ここで部屋のプライベート Props やクライアントリストを確認可能。

- クライアントリストの左側にある「-」ボタンでプレイヤーをキックできる。
