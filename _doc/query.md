# Query クラス

## 目次
- [概要](#概要)
- [Null判定](#Null判定)
  - `IsNull(key)`
  - `IsNotNull(key)`
- [同値の判定](#同値の判定)
  - `Equal(key, value)`
  - `NotEqual(key, value)`
- [数値範囲の判定](#数値範囲の判定)
  - `LessThan(key, val)`
  - `LessEqual(key, val)`
  - `GreaterEqual(key, val)`
  - `GreaterThan(key, val)`
  - `Between(key, min, max)`
- [リストに含まれるかの判定](#リストに含まれるかの判定)
  - Contain(key, val)
  - NotContain(key, val)
- [論理結合](#論理結合)
  - `And(queries...)`
  - `Or(queries...)`
- 具体例
  - [最近戦ったユーザとのマッチングを回避する](#最近戦ったユーザとのマッチングを回避する)

## 概要

入室時や部屋検索のときに、部屋の公開プロパティ(`Room.PublicProps`)を使ってフィルタリングすることができます。
このフィルタリング条件は[`Query`](../wsnet2-unity/Assets/WSNet2/Scripts/Core/Query.cs)オブジェクトで指定します。
`Query`の条件にマッチする部屋のみ、検索や入室の対象になります。

`Query`クラスの各メソッドは`Query`オブジェクト自身を返すので、メソッドチェインの形で記述できます。

## Null判定

```C#
Query IsNull(string key);
Query IsNotNull(string key);
```

公開プロパティの`key`の値が`null`のとき`IsNull()`はマッチします。
型に関わらず値が入っているとき`IsNull`は非マッチとなります。
`IsNotNull()`はその逆です。

`key`が存在しないときはどちらもマッチしません。

## 同値の判定

```C#
Query Equal(string key, T val);
Query Not(string key, T val);
```

型`T`は、`bool`, `char`, `string`,
`sbyte`, `byte`, `short`, `ushort`, `int`, `uint`, `long`, `ulong`,
`float`, `double` のいずれかです。

公開プロパティの`key`の値が、`T`型で`val`と等しいときに`Equal()`はマッチします。
T型で`val`と異なるときに`Not()`はマッチします。

`key`が存在しない時、値の型が`T`と異なる時はどちらも常にマッチしません。

## 数値範囲の判定

```C#
Query LessThan(string key, T val);
Query LessEqual(string key, T val);
Query GreaterEqual(string key, T val);
Query GreaterThan(string key, T val);
Query Between(string key, T min, T max);
```

型`T`は、
`sbyte`, `byte`, `short`, `ushort`, `int`, `uint`, `long`, `ulong`,
`float`, `double` のいずれかです。

公開プロパティの`key`の値が`T`型で、`val`との大小関係が合致しているときにマッチします。
`Between()`は`min`、`max`を範囲に含みます。

`key`が存在しない時、値の型が`T`と異なる時はいずれも常にマッチしません。

## リストに含まれるかの判定

```C#
Query Contain(string key, T val);
Query NotContain(string key, T val);
```

型`T`は、`bool`, `char`, `string`,
`sbyte`, `byte`, `short`, `ushort`, `int`, `uint`, `long`, `ulong`,
`float`, `double` のいずれかです。

公開プロパティの`key`の値が[配列またはリスト型](serializable.md#配列リスト)で、その要素として`T`型で`val`と等しいものが含まれるとき`Contain()`はマッチします。
そのような要素が含まれないときに`NotContain()`はマッチします。

`key`が存在しないときや配列またはリスト型では無かったときは、いずれもマッチしません。

## 論理結合

```C#
Query And(param Query[] queries);
Query Or(param Query[] queries);
```

`And()`は、与えられた`Query`の全てがマッチしているときにマッチします。
`Or()`は、与えられた`Query`の少なくとも1つがマッチしているときにマッチします。

空のQueryは常にマッチします。

## 具体例

### 最近戦ったユーザとのマッチングを回避する

部屋の公開プロパティに、次の情報を設定しておきます。
- `HostUserId`: ホストのユーザID
- `RecentlyMatchedIds`: ホストが最近対戦したユーザのIDのリスト

マッチング条件のクエリで次のように指定することで、
ホスト・ゲストのどちらから見ても最近対戦したユーザとのマッチングを避けることができます。

```C#
var query = new Query();

// 自身の最近対戦したユーザにホストが含まれないこと
for (var recentlyMatchedId in recentlyMatchedIds)
{
    query.NotEqual("HostUserId", recentlyMatchedId);
}

// ホストの最近対戦したユーザリストに自身が含まれないこと
query.NotContain("RecentlyMatchedIds", myId);
```
