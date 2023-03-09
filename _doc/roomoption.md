# RoomOption クラス

## 目次

- [概要](#概要)
- [コンストラクタ](#コンストラクタ)
- [各フラグの設定](#各フラグの設定)
- [部屋番号の割り当て](#部屋番号の割り当て)
- [その他の設定](#その他の設定)

## 概要

部屋作成時の各プロパティは`RoomOption`で指定します。
`RoomOption`の各メソッドは`RoomOption`自身を返すので、メソッドチェインとして記述できます。

各プロパティについては[Roomクラス](room.md#部屋のプロパティ)を参照してください。

## コンストラクタ

```C#
RoomOption(
    uint maxPlayers,
    uint searchGroup,
    IDictionary<string, object> publicProps,
    IDictionary<string, object> privateProps);
```

- `maxPlayers`: 最大プレイヤー数
- `searchGroup`: 検索グループ
- `publicProps`: 公開プロパティ
- `privateProps`: 非公開プロパティ

### その他の初期値

- Visible: `true`
- Joinable: `true`
- Watchable: `true`
- WithNumber: `false`
- ClientDeadline: サーバ設定による
- LogLevel: サーバ設定による

## 各フラグの設定

```C#
RoomOption Visible(bool val);
RoomOption Joinable(bool val);
RoomOption Watchable(bool val);
```

グループ検索可能、入室可能、観戦可能の各フラグを設定します。

## 部屋番号の割り当て

```C#
RoomOption WithNumber(bool val);
```

部屋番号を割り当てかを設定します。

## その他の設定

### ClientDeadline

```C#
RoomOption WithClientDeadline(uint sec);
```

`ClientDeadline`を設定します。

- `sec`: 設定値（秒）

### LogLevel

```C#
RoomOption SetLogLevel(LogLevel l);
```

部屋のログレベルを設定します。

- `l`: ログレベル
  - `RoomOption.LogLevel.DEFAULT`
  - `RoomOption.LogLevel.NOLOG`
  - `RoomOption.LogLevel.ERROR`
  - `RoomOption.LogLevel.INFO`
  - `RoomOption.LogLevel.DEBUG`
  - `RoomOption.LogLevel.ALL`
