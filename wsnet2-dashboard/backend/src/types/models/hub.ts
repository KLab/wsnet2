import { objectType } from "nexus";
import np from "nexus-prisma";
const { hub } = np;

export const hubModel = objectType({
  name: hub.$name,
  description: hub.$description,
  definition(t) {
    t.field(hub.id);
    t.field(hub.host_id);
    t.field(hub.room_id);
    t.field(hub.watchers);
    t.field(hub.created);
  },
});
