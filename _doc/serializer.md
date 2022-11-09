## 目次
- [概要](#概要)
- [IWSNet2Serializableの実装](#IWSNet2Serializableの実装)
- [独自型の登録](#独自型の登録)
- [SerialWriterのメソッド](#SerialWriterのメソッド)
- [SerialReaderのメソッド](#SerialReaderのメソッド)

## 概要

WSNet2のシリアライザは、C#のほとんどのプリミティブ型とIWSNet2Serializableを実装したクラス、
それらを要素として持つリストや辞書をシリアライズできます。（[シリアライズ可能な型](シリアライズ可能な型)）

独自クラスに`IWSNet2Serializable`を実装するときには、
シリアライザの3つのクラス、`WSNet2Serializer`、`SerialReader`、`SerialWriter`を利用します。

ここでは、この3つのクラスを利用した`IWSNet2Serializable`の実装方法について説明します。

## IWSNet2Serializableの実装

### 例
```C#
using WSNet2;

// 様々な型のメンバーを持てる
class UserData : IWSNet2Serializable
{
	string name;
	int[]  numbers;

	// 「シリアライズ可能な型」以外もメンバには持てるが、
	// シリアライズできないので通信に含められない
	Hashtable cache;

	(省略)

	public void Serialize(SerialWriter writer)
	{
		// シリアライズしたいフィールドを順に書き込む
		writer.Write(name);
		writer.Write(numbers);
	}

	public void Deserialize(SerialReader reader, int size)
	{
		// 書き込んだのと同じ順序で取り出す
		name = reader.ReadString();

		// numbersのサイズが同じであれば再利用でき
		// メモリアロケーションを抑制できる
		numbers = reader.ReadInts(numbers);
	}
}


class TeamData : IWSNet2Serializable
{
	UserData leader;
	List<UserData> members;

	(省略)

	public void Serialize(SerialWriter writer)
	{
		// UserDataは IWSNet2Serializable なのでシリアライズできる
		writer.Write(leader);
		writer.Write(members);
	}

	public void Deserialize(SerialReader reader, int size)
	{
		// 書き込んだのと同じ順序で取り出す
		leader = reader.ReadObject(leader);
		members = reader.ReadList(members);
}
```

### IWSNet2Serializableインターフェイス


`IWSNet2Serializable`インターフェイスは次のように2つのメソッドを持ちます。

```C#
interface IWSNet2Serializable
{
    void Serialize(SerialWriter writer);
    void Deserialize(SerialReader reader, int size);
}
```

#### Serialize

`Serialize()`メソッドでは、引数の`SerialWriter`に必要なフィールドを順に書き込んでいきます。
シリアライズ可能な型であれば、`writer.Write()`メソッドが用意されているので、そのまま呼び出せます。

`IWSNet2Serializable`型の場合は、事前に`WSNet2Serializer.Register<T>(classID)`で登録しておく必要があります。

メンバとしてはシリアライズできない型であったとしても、
分解して書き込んでおいて`Deserialize`で再構築するという手段もとれます。

#### Deserialize

`Deserialize()`メソッドでは、引数の`SerialReader`から値を取り出していきます。
この取り出し順序は、`Serialize`で書き込んだのと完全に一致している必要があります。

`IWSNet2Serializable`型のオブジェクトを`reader.ReadObject`などで取り出す場合、
`Serialize`と同様、事前に`WSNet2Serializer.Register<T>(classID)`で登録しておく必要があります。
シリアライズする側とデシリアライズする側で`T`と`classID`の対応関係も一致している必要があります。

また、`writer.Write`ではint型だったが`reader.ReadByte`を呼んだ場合など、型が異なると例外が発生します。

## 独自型の登録

```C#
// classIDが重複しないように登録する
WSNet2Serializer.Register<UserData>(1);
WSNet2Serializer.Register<TeamData>(2);
```

`IWSNet2Serializable`を実装した独自の型は`WSNet2Serializer.Register<T>(classID)`メソッドを呼び出し、
型とclassIDを紐づけて登録することでシリアライズ出来るようになります。

このメソッドではグローバルなデータとして登録されますので、同じ型あるいはclassIDで複数回呼び出すことはできません。

プログラム内で利用する全ての型を、静的コンストラクタで一括で登録するのがよいでしょう。

## SerialWriterのメソッド

`SerialWriter.Write`メソッドは、シリアライズ可能な型が網羅されるようにオーバーロードされています。
一部より広い型を受け付けますが、基本的にはシリアライズ可能な型に絞って利用してください。

```C#
void Write(bool v);
void Write(sbyte v);
void Write(byte v);
void Write(char v);
void Write(short v);
void Write(ushort v);
void Write(int v);
void Write(uint v);
void Write(long v);
void Write(ulong v);
void Write(float v);
void Write(double v);
void Write(string v);
void Write<T>(T v) where T : class, IWSNet2Serializable;
void Write(IEnumerable v);
void Write(IDictionary<string, object> v);
void Write(bool[] vals);
void Write(sbyte[] vals);
void Write(byte[] vals);
void Write(char[] vals);
void Write(short[] vals);
void Write(ushort[] vals);
void Write(int[] vals);
void Write(uint[] vals);
void Write(long[] vals);
void Write(ulong[] vals);
void Write(float[] vals);
void Write(double[] vals);
```

## SerialReaderのメソッド

`SerialReader`から値を読み取るには、型に対応したメソッドを利用します。
書き込まれていたデータとメソッドの型が一致していない場合、例外が発生します。

### プリミティブ型

プリミティブ型は引数なしで単純に値を返します。

```C#
bool   ReadBool();
sbyte  ReadSByte();
byte   ReadByte();
char   ReadChar();
short  ReadShort();
ushort ReadUShort();
int    ReadInt();
uint   ReadUint();
long   ReadLong();
ulong  ReadULong();
float  ReadFloat();
double ReadDouble();
string ReadString();
```

### プリミティブ型の配列

プリミティブ型の配列は再利用オブジェクトを引数として渡すことができます。
再利用オブジェクトの配列長が取り出される配列長と一致している場合、
配列を再確保することなく中身を上書きして返します。

長さが異なる場合は配列を新たに確保します。

```C#
bool[]   ReadBools(bool[] recycle = null);
sbyte[]  ReadSBytes(sbyte[] recycle = null);
byte[]   ReadBytes(byte[] recycle = null);
char[]   ReadChars(char[] recycle = null);
short[]  ReadShorts(short[] recycle = null);
ushort[] ReadUShorts(ushort[] recycle = null);
int[]    ReadInts(int[] recycle = null);
uint[]   ReadUInts(uint[] recycle = null);
long[]   ReadLongs(long[] recycle = null);
ulong[]  ReadULongs(ulong[] recycle = null);
float[]  ReadFloats(float[] recycle = null);
double[] ReadDoubles(double[] recycle = null);
string[] ReadStrings(string[] recycle = null);
```

### IWSNet2Serializable型のデシリアライズ

登録された型またはそれのみ含まれる配列とリストは、次のようなジェネリクスメソッドで読み出せます。
また、再利用オブジェクトを引数に与えることができます。
リストや配列を再利用する場合、そのリストや配列自体も可能なら再利用することに加え、
各要素をデシリアライズするときに、同じ添字番号のオブジェクトを再利用します。

```C#
T       ReadObject<T>(T recycle = default) where T : class, IWSNet2Serializable, new();
List<T> ReadList<T>(List<T> recycle = null) where T : class, IWSNet2Serializable, new();
T[]     ReadArray<T>(T[] recycle = null) where T : class, IWSNet2Serializable, new();
```

### 辞書・リスト・配列のデシリアライズ

辞書・リスト・配列は次のようなメソッドで読みだせ、再利用オブジェクトも引数に与えられます。
各要素をデシリアライズするときに、同じキーや添字番号のものを可能なら再利用します。
リストや配列では可能であればリストや配列自体も再利用しますが、辞書は毎回生成しなおします。

```C#
Dictionary<string, object> ReadDict(IDictionary<string, object> recycle = null);
List<object>               ReadList(List<object> recycle = null);
object[]                   ReadArray(object[] recycle = null);
```
