- [概要](#概要)
- [設定項目](#設定項目)
  - [EvPoolSize](#EvPoolSize)
  - [EvBufInitialSize](#EvBufInitialSize)
  - [MsgPoolSize](#MsgPoolSize)
  - [MsgBufInitialSize](#MsgBufInitialSize)
  - [MaxReconnection](#MaxReconnection)
  - [ConnectTimeoutMilliSec](#ConnectTimeoutMilliSec)
  - [RetryIntervalMilliSec](#RetryIntervalMilliSec)
  - [MaxPingIntervalMilliSec](#MaxPingIntervalMilliSec)
  - [MinPingIntervalMilliSec](#MinPingIntervalMilliSec)

## 概要

WSNet2をプロジェクトに合わせて調整するための設定を[`WSNet2.WSNet2Settings`](/WSNet/wsnet2/blob/master/wsnet2-unity/Assets/WSNet2/Scripts/Core/WSNet2Settings.cs)クラスを通して行えます。

## 設定項目

### EvPoolSize

- 保持できるEventの数
- 初期値: 16 個

入室時の`onSuccess`コールバック中や`room.Pause()`で一時停止している間に受信したEventを保持するためのプールサイズです。
自動拡張されないので、十分な量を設定してください。

### EvBufInitialSize

- 各Eventのバッファサイズの初期値
- 初期値: 256 byte

各イベントを保持するバッファサイズの初期値です。
`EvPoolSize`個のバッファが確保されます。
各バッファは必要に応じて自動的に拡張されます。

### MsgPoolSize

- 保持できるMsgの数
- 初期値: 16 個

送信待ちメッセージを保持するためのプールのサイズです。
回線切断による再接続中のメッセージをプールするだけでなく、再接続時の不達メッセージの再送にも使われます。

### MsgBufInitialSize

- 各Msgのバッファサイズの初期値
- 初期値: 256 byte

各メッセージを保持するバッファサイズの初期値です。
`MsgPoolSize`個のバッファが確保されます。
各バッファは必要に応じて自動的に拡張されます。

### MaxReconnection

- 最大連続再接続試行回数
- 初期値: 5 回

websocket再接続試行回数の上限です。
この回数連続して接続に失敗したら再試行を打ち切り、部屋から切断したことになります。
再接続に成功したら失敗回数のカウントは0に戻ります。

### ConnectTimeoutMillSec

- 接続タイムアウト
- 初期値: 5000 ミリ秒 (=5秒)

websocket接続のタイムアウト時間です。
接続要求に対してこの時間応答が無かった場合、その接続は失敗として扱われます。

### RetryIntervalMilliSec

- 再接続インターバル
- 初期値: 1000 ミリ秒 (=1秒)

接続要求してから、次の再接続を要求するまでの最低間隔です。
この時間以上接続していたり、タイムアウト時間がこの時間より長い場合は、即時に再接続が試行されます。

### MaxPingIntervalMilliSec

- 最大Ping間隔
- 初期値: 10000 ミリ秒 (=10秒)

Ping間隔は部屋に設定された[`ClientDeadline`](Roomの使い方#clientdeadline)に応じて自動計算されますが、
この時間より長くなることはありません。

WSNet2では、Pingメッセージに対応するPongイベントによって観戦者数や各Playerの最終Msg受信時刻を同期するため、
あまり長い間隔にしてしまうとこれらの情報のリアルタイム性が損なわれてしまいます。

### MinPingIntervalMilliSec

- 最小Ping間隔
- 初期値: 1000 ミリ秒 (=1秒)

Ping間隔はRoomの`ClientDeadline`に応じて自動計算されますが、通信頻度が過剰にならないようにするため、この時間より短くなることはありません。
逆に言えば、`ClientDeadline`がこの値より短い場合、タイムアウトしないためのPingが機能しなくなります。
