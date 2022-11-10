# サーバの構築

## 目次

- [サーバプログラムのビルド](#サーバプログラムのビルド)
- [データベースの構築](#データベースの構築)
- [サーバ設定ファイル](#サーバ設定ファイル)
  - [ファイルの内容](#ファイルの内容)
  - [環境変数による設定](#環境変数による設定)

## サーバプログラムのビルド

WSNet2のサーバはGoで書かれたプログラムです。
ビルドするにはGo（1.18以降）のほか、GNU MakeとUnix風シェルが必要です。

[`server`ディレクトリ](../server)にて、`make`コマンドを実行すると、

`bin`ディレクトリに次の実行ファイルが作られます。
WSNet2の機能として必要なのはLobby、Game、Hubの3種類のサーバです。

- **wsnet2-lobby**: Lobbyサーバ
- **wsnet2-game**: Gameサーバ
- **wsnet2-hub**: Hubサーバ
- **wsnet2-bot**: 負荷試験やシナリオ試験用のbotクライアント
- **wsnet2-tool**: サーバや部屋の情報を閲覧するコマンドラインツール

## データベースの構築

MySQL互換RDBMSを利用します。

必要なテーブルは[`sql/10-schema.sql`](../server/sql/10-schema.sql)に定義されています。

- **app**: 登録アプリ識別子と鍵
- **game_server**: Gameサーバの接続情報と状態
- **hub_server**: Hubサーバの接続情報と状態
- **room**: 稼働中の部屋
- **hub**: 稼働中の観戦用部屋
- **room_history**: 終了した部屋

最初に`app`テーブルにAppIDとKeyを登録します。この情報はゲームAPIサーバと共有するもので[ユーザ認証](user_auth.md#鍵の事前交換)に使われます。

その他のテーブルは自動で書き込まれるため、空のままにします。

## サーバ設定ファイル

サーバプログラム（wsnet2-lobby、wsnet2-game、wsnet2-hub）の起動には、
コマンドライン引数としてTOML形式の設定ファイルを指定します。

サーバ種別ごと別テーブルに分けているので、同一のファイルを全種類のサーバに使えます。

### ファイルの内容

```toml
#
# RDBMSに関する設定
#
[Database]
host = "wsnet2-db"
port = 3306
dbname = "wsnet2"

# 接続ユーザ
user = "wsnet"
password = "password"
# user, passwordの代わりに、authfileで指定することもできます
authfile = "/path/to/authfile" # "user:password" が書かれたファイル

conn_max_lifetime = "3m" # 接続を再利用できる最大時間（デフォルト:3m）

#
# Lobbyサーバに関する設定
#
[Lobby]
hostname = ""     # 指定した場合、ホスト名がマッチしたHTTPリクエストのみ受け付ける
net = "tcp"       # ソケット種別。"unix": UNIXドメインソケット
unixpath = ""     # net="unix"のときのunixドメインソケットのパス
port = 8080       # net="tcp"のときのポート番号
pprof_prot = 3080 # pprofの待受けポート番号

valid_heartbeat = "5s" # Game,Hubの最終HeartBeat時刻の有効期間（デフォルト:5s）
authdata_expire = "1m" # 認証データの有効期間（デフォルト:1m）
api_timeout = "5s"     # LobbyAPIの内部タイムアウト時間（デフォルト:5s）
db_max_conns = 0       # 最大DB接続数
hub_max_watchers = 10000 # Hubサーバの最大収容観戦者数

# ログ設定
loglevel = 5 # 基本ログレベル（デフォルト:2）
log_stdout_level = 4       # stdoutのログレベル
log_stdout_console = false # stdoutのログフォーマットを開発用にする
# ログファイル出力とローテーションの設定
log_path = ""         # 空ならファイルに出力しない
log_max_size = 500
log_max_backups = 0
log_max_age = 0
log_compress = false

#
# Gameサーバの設定
#
[Game]
hostname = "wsnet2-game"                # ローカルホスト名（Lobby, Hubからのアクセス）
public_name = "wsnet2-game.example.com" # 公開ホスト名（クライアントからのアクセス）
grpc_port = 19000                       # gRPC待受けポート（Lobby, Hubからのアクセス）
websocket_port = 8000                   # WebSocket待受けポート（クライアント、Hubからのアクセス）
pprof_port = 3000
# WebSocket接続にTLSを使用する場合の設定
# 空ならTLSを使用しない
tls_cert = "/path/to/cert_file"
tls_key = "/path/to/key_file"

retry_count = 5        # ユニークな部屋番号生成のリトライ回数（デフォルト:5）
max_room_num = 999999  # 部屋番号の最大値。3桁に制限したいときは 999 とする
max_rooms = 1000       # 最大部屋数（デフォルト：1000）
max_clients = 5000     # 最大クライアント数（デフォルト：5000）
db_max_conns = 0       # 最大DB接続数
heartbeat_interval = "2s" # HeartBeat時刻更新間隔。{Lobby,Hub}.valid_heartbeatより短くする。
# 部屋の初期値
default_max_players = 10 # 部屋あたりの最大プレイヤー数（デフォルト:10）
default_deadline = 5     # クライアントタイムアウト判定時間（秒; デフォルト:5）
default_loglevel = 2     # 部屋のログレベル
# client設定
event_buf_size = 128     # イベント再送バッファ数（デフォルト:128）
wait_after_close = "30s" # 部屋終了後の再接続データ再送可能時間（デフォルト:30s）
auth_key_len = 32               # 接続のユーザ認証用の鍵のサイズ

# ログ設定（Lobbyと同じ）
loglevel = 2
log_stdout_level = 4
log_stdout_console = false
log_path = ""
log_max_size = 500
log_max_backups = 0
log_max_age = 0
log_compress = false

#
# Hubサーバの設定
#
[Hub]
# 基本的にGameと同じ
hostname = "wsnet2-hub"
public_name = "wsnet2-hub.example.com"
grpc_port = 19010
websocket_port = 8010
pprof_port = 3010
tls_cert = "/path/to/cert/file"
tls_key = "/path/to/key/file"
max_clients = 5000
default_loglevel = 2
valid_heartbeat = "5s"     # Gameの最終HeartBeat時刻の有効期間（デフォルト:5s）
heartbeat_interval = "2s"
nodecount_interval = "1s"  # Hubを経由している観戦者数の同期間隔（デフォルト:1s）
db_max_conns = 0
event_buf_size = 128
wait_after_close = "30s"
auth_key_len = 32
loglevel = 2
log_stdout_level = 4
log_stdout_console = false
log_path = ""
log_max_size = 500
log_max_backups = 0
log_max_age = 0
log_compress = false
```

### 環境変数による設定

GameとHubの`hostname`、`public_name`、`grpc_port`、`websocket_port`は次の環境変数で上書きできます。
複数台構成ではホスト名を環境変数で指定することで、設定ファイルを共通にできます。

- `WSNET2_GAME_HOSTNAME`
- `WSNET2_GAME_PUBLICNAME`
- `WSNET2_GAME_GRPCPORT`
- `WSNET2_GAME_WSPORT`
