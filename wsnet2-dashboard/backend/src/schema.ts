import { makeSchema } from "nexus";
import * as types from "./types";

export const schema = makeSchema({
  types,
  plugins: [],
});
