import { objectType } from "nexus";
import np from "nexus-prisma";
const { room } = np;

export const roomModel = objectType({
  name: room.$name,
  description: room.$description,
  definition(t) {
    t.field(room.id);
    t.field(room.app_id);
    t.field(room.host_id);
    t.field(room.visible);
    t.field(room.joinable);
    t.field(room.watchable);
    t.field(room.number);
    t.field(room.search_group);
    t.field(room.max_players);
    t.field(room.players);
    t.field(room.watchers);
    t.field(room.props);
    t.field(room.created);
  },
});
