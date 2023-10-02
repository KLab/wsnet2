# Backend

[English](README.md)

wsnet2-dashboard 専用バックエンド（所謂 BFF）、
フロントエンドが使う`GraphQL`と`GRPC`の API を提供する。

## 環境変数

| 環境変数名              | 説明                                                      | 例                                              |
| ----------------------- | --------------------------------------------------------- | ----------------------------------------------- |
| SERVER_PORT             | dashboard サーバーに使われるポート番号                    | "5555"                                          |
| DATABASE_URL            | データベースの指定（DB 名やパスワードも必要）             | "mysql://wsnet:wsnetpass@localhost:3306/wsnet2" |
| FRONTEND_ORIGIN         | ダッシュボードフロントエンドの IP アドレス（CORS 指定用） | "http://localhost:3000"                         |
| GRAPHQL_RESULT_MAX_SIZE | GraphQL の一回のクエリーで取得できるアイテムの数の上限値  | "1000"                                          |

## アプリのビルド

データベースからスキーマを生成するので、`db`を起動しておく必要がある。

```bash
docker compose -f server/compose.yaml up -d db
cd wsnet2-dashboard
docker compose run --rm backend make
```

生成されたコードは `wsnet2-dashboard/backend/dist` に格納される。

## GraphQL について

[ORM](https://en.wikipedia.org/wiki/Object%E2%80%93relational_mapping)は[Prisma](https://www.prisma.io/)を採用。

- GraphQL スキーマの自動生成：
  - データベースの環境変数を`.env`に正しく設定する
  - backend のルートディレクトリで`npm run generate`
  - 自動生成されたスキーマは[`prisma/schema.prisma`](prisma/schema.prisma)に書き込まれる
- Apollo Server で公開する GraphQL の API は別途手動で指定（実装）する必要がある、場所は[`src/types`](src/types/)

## gRPC について

wsnet2 本体との gRPC 通信はバックエンドで行う。ダッシュボードへのデータ転送は通常の REST で行う（データフォマットは`application/msgpack`のみサポート）。

gRPC コードの自動生成は、`make` により自動的に行われる。
（`server/pb/*.proto` から生成し、`wsnet2-dashboard/backend/src/pb` 以下に格納する）

## 開発ノート

### フレームワークと開発環境

#### nodejs と express

今回の wsnet2-dashboard は、フロントエンドがメインです。バックエンドはデータや機能を持つ wsnet2 本体と dashboard フロントエンドを繋ぐためのもので、所謂「Backends for Frontends（BFF）」。開発環境選定については、開発効率とフロントエンドとの連携のしやすさを考えて、軽めでフロントエンドと同じ javascript で書ける nodejs（javascript ランタイム）+express（バックエンドフレームワーク）にしました。

> 参考：[step by steo BFF](https://speakerdeck.com/yosuke_furukawa/step-by-step-bff)

#### vscode と typescript

開発環境はフロントエンド同様、vscode + typescript の構成にしています。リポジトリも一緒なので、このように vscode でフロントエンドと同時に開発でき、連携コストを最小に出来ました。

![](../images/vscode.png)

### GraphQL 実装詳細

wsnet2 本体のデータベースから情報を取得して、dashboard のフロントエンドに渡すための API は GraphQL を使います。

GraphQL のおさらい（CircleCI の「[GraphQL とは？メリットや概要を入門ガイドで学ぶ](https://circleci.com/ja/blog/introduction-to-graphql/)」より）：

> GraphQL とは API のために作られたクエリ言語であり、既存のデータに対するクエリを実行するランタイムです。
> 理解できる完全な形で API 内のデータについて記述します。
> GraphQL を利用すれば、クライアント側から必要な内容だけを問い合わせられると共に、漸次的に API を進化させることが容易になり、強力な開発者ツールを実現できます。

GraphQL と通常の RESTful API に比べて、最大の特徴はクライアントの方が欲しい情報だけを取得する事、これによって高い柔軟性を持ちながら、ネットワーク通信量を最小限に抑えられる。

dashboard のようなフィルターでデータを絞る操作をよくやるアプリケーションでは、GraphQL を使えば、実装する API は少ない数で済むメリットがあります。

GraphQL はフロントエンド開発者にとって、少ない API で幅広いことができるので、非常に使いやすいですが、バックエンドでの実装は伝統的な RESTful API よりは大変でした。昔はよくこう言われてます：「フロントエンド開発者は GraphQL で楽できるが、その分、バックエンド開発者の負担が増えている」。実際、GraphQL が登場してから数年経ったが、バックエンドの実装が大変だった事情があり、まだそこまで普及出来てなかった。しかし最近ではいろんな開発補助ツールが充実してきているので、GraphQL のバックエンド開発もだいぶやりやすくなりました。

今回のバックエンド GraphQL API 実装は prisma + nexus の組み合わせを使いました。

- [Prisma](https://www.prisma.io/)： 所謂 ORM（Object-Relational Mapping）、複数の種類の DB をサポートし、SQL を直接書くことなく、オブジェクトのメソッドで DB 操作ができる
- [Nexus](https://nexusjs.org/)：GraphQL API の実装補助ツール
- [nexus-prisma](https://nexus.prisma.io/)：Nexus と Prisma を連携させるためのプラグイン

バックエンド側 GraphQL 実装の流れ

- Prisma で DB をから`prisma.schema`を自動生成する（`npx prisma db pull && npx prisma generate`）
- `prisma.schema`の Nexus 関連設定を更新する

  ```
  generator client {
    provider      = "prisma-client-js"
  }

  generator nexusPrisma {
    provider      = "nexus-prisma"
  }

  datasource db {
    provider = "mysql"
    url      = env("DATABASE_URL")
  }
  ```

- Nexus で GraphQL が扱う model を実装する

  ```typescript
  // 例：types/models/app.ts
  import { objectType } from "nexus";
  import { app } from "nexus-prisma";

  export const appModel = objectType({
    name: app.$name,
    description: app.$description,
    definition(t) {
      t.field(app.id);
      t.field(app.name);
      t.field(app.key);
    },
  });
  ```

- Nexus で GraphQL API の resolver を実装する

  ```typescript
  // 例：types/resolvers/appResolver.ts
  import { extendType, idArg, stringArg, nonNull } from "nexus";
  import { Context } from "../../context";
  import { Prisma } from "@prisma/client";

  export const appQuery = extendType({
    type: "Query",
    definition(t) {
      t.list.field("apps", {
        type: "app",
        description: "Get all apps",
        resolve(_, __, ctx: Context) {
          return ctx.prisma.app.findMany();
        },
      });

      t.field("appById", {
        type: "app",
        description: "Get unique app by id",
        args: {
          id: idArg(),
        },
        resolve(_, { id }, ctx: Context) {
          return ctx.prisma.app.findUnique({
            where: { id: String(id) },
          });
        },
      });
    },
  });
  ```

- 必要に応じて、カスタムスカラー（Bytes/JSON とか）を定義する

  ```typescript
  // 例：types/models/scalers.ts
  import { scalarType } from "nexus";
  import {
    GraphQLDateTime,
    GraphQLBigInt,
    GraphQLByte,
    GraphQLJSON,
  } from "graphql-scalars";
  import binary = require("../../plugins/binary");

  export const dateTimeScalar = GraphQLDateTime;
  export const bigIntScalar = GraphQLBigInt;

  export const jsonScalar = scalarType({
    name: "Json",
    asNexusMethod: "json",
    description: "Json custom scalar type",
    parseValue: GraphQLJSON.parseValue,
    serialize: GraphQLJSON.serialize,
    parseLiteral: GraphQLJSON.parseLiteral,
  });

  export const bytesScalar = scalarType({
    name: "Bytes",
    asNexusMethod: "bytes",
    description: "Bytes custom scalar type",
    parseValue: GraphQLByte.parseValue,
    serialize(value: Uint8Array) {
      return binary.UnmarshalRecursive(value);
    },
    parseLiteral: GraphQLByte.parseLiteral,
  });
  ```

- 実装した GraphQL を express に登録する

  ```typescript
  import express from "express";
  import { ApolloServer } from "apollo-server-express";
  import { makeSchema } from "nexus";
  import * as types from "./types";
  import { createContext } from "./context";

  const app = express();

  const schema = makeSchema({
    types,
    plugins: [],
  });

  const server = new ApolloServer({
    schema: schema,
    context: createContext,
  });

  // apollo serverを先に起動する必要がある
  await server.start();
  // graphqlのpathを指定する
  server.applyMiddleware({ app, path: "/graphql" });
  ```

### Go コードの転用（gopherjs）

wsnet2 本体に Go で実装されたシリアライザーを BFF の nodejs 側でも使えるように、関連処理を gopherjs で javascript に変換してます（[backend/src/plugins/binary.js](backend/src/plugins/binary.js)）。Go 側のシリアライザーに変更が入る場合、こちらの javascript コードも併せて更新する必要があります。
