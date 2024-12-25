import { makeSchema } from "nexus";
import * as types from "./types/index.js";

export const schema = makeSchema({
  types,
  plugins: [],
});
