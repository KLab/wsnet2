# シリアライズ可能な型

## 目次

- [概要](#概要)
- [型一覧](#型一覧)
  - [プリミティブ型](#プリミティブ型)
  - [IWSNet2Serializable型](#IWSNet2Serializable型)
  - [辞書型](#辞書型)
  - [配列・リスト](#配列リスト)
- [Nullの扱い](#Nullの扱い)

## 概要

部屋とプレイヤーのプロパティ辞書の値や、RPCの引数には、シリアライズ可能な型のみ指定できます。
WSNet2のシリアライザでは、型は基本的には保存されます。
シリアライズやデシリアライズの詳細は[シリアライザの使い方](serializer.md)を参照してください。

## 型一覧

### プリミティブ型

次のプリミティブ型はそのままシリアライズできます。

- `bool`
- `sbyte`
- `byte`
- `char`
- `short`
- `ushort`
- `int`
- `uint`
- `long`
- `ulong`
- `float`
- `double`
- `string`

### IWSNet2Serializable型

[`IWSNet2Serializable`](serializer.md#IWSNet2Serializableインターフェイス)を実装した型は、
`WSNet2Serializer`クラスに登録することでシリアライズ出来るようになります。

```C#
static void Register<T>(byte classID) where T : class, IWSNet2Serializable, new();
```

型ごとに1byteのclassIDを割り当てるため、最大256種類登録できます。
全てのクライアントで型とclassIDの対応は一致している必要があります。

### 辞書型

キーが`string`で、値が「シリアライズ可能な型」の辞書`Dictionary<string, object>`型がシリアライズできます。
値として辞書やリストを含めることもできます。

### 配列・リスト

シリアライズ可能なプリミティブ型の配列（`int[]`など）はそのままシリアライズできます。

`IWSNet2Serializable`を実装した単一の型`T`について、`List<T>`と`T[]`もシリアライズ可能です。

さらに、全要素が「シリアライズ可能な型」の`List<object>`や`object[]`もシリアライズでき、
この場合要素にリストや辞書をネストして含めることができます。

## Nullの扱い

文字列や`IWSNet2Serializable`、辞書、配列、リストは`null`にすることができ、
型なしのNullとしてシリアライズされます。

デシリアライズ時には指定の型の`null`になります。
