import { scalarType } from "nexus";
import {
  GraphQLDateTime,
  GraphQLBigInt,
  GraphQLByte,
  GraphQLJSON,
} from "graphql-scalars";
import { binary } from "../../plugins/binary.js";

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
