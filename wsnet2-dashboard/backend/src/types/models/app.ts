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
