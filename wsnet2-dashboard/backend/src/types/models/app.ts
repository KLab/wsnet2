import { objectType } from "nexus";
import np from "nexus-prisma";
const { app } = np;

export const appModel = objectType({
  name: app.$name,
  description: app.$description,
  definition(t) {
    t.field(app.id);
    t.field(app.name);
    t.field(app.key);
  },
});
