# Backend

[日本語](README-ja.md)

Backend for wsnet2-dashboard.
It provides APIs of `GraphQL` and `GRPC` for the dashboard.

## Environment variables

| Name            | Description                                          | Example                                         |
| --------------- | ---------------------------------------------------- | ----------------------------------------------- |
| SERVER_PORT     | Port of the dashboard's server                       | "5555"                                          |
| DATABASE_URL    | Database uri（including database name and password） | "mysql://wsnet:wsnetpass@localhost:3306/wsnet2" |
| FRONTEND_ORIGIN | IP address of the dashboard frontend（for CORS）     | "http://localhost:3000"                         |

## About the GraphQL

The [ORM](https://en.wikipedia.org/wiki/Object%E2%80%93relational_mapping) being used is [Prisma](https://www.prisma.io/).

- GraphQL schema auto generation：
  - Set the database uri in `.env`.
  - Run `npm run generate` at the backend's root directory.
  - Auto generated schema is stored at [`prisma/schema.prisma`](prisma/schema.prisma).
- GraphQL APIs available to Apollo server is implemented at [`src/types`](src/types/).

## About the GRPC

The GRPC communication with wsnet2-server is carried out at the backend. The dashboard and the backend communicates through REST API(using `application/msgpack`).

- GRPC code auto generation
  - [Prepare the environment for protobuf compilation](https://grpc.io/docs/protoc-installation/).
  - Set the directory of proto files in `PROTO_PATH` inside [`Makefile`](Makefile).
  - Run `make` at the backend's root directory.
