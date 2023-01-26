import { objectType } from "nexus";
import { room_history } from "nexus-prisma";

export const roomHistoryModel = objectType({
  name: room_history.$name,
  description: room_history.$description,
  definition(t) {
    t.field(room_history.id);
    t.field(room_history.app_id);
    t.field(room_history.host_id);
    t.field(room_history.room_id);
    t.field(room_history.number);
    t.field(room_history.search_group);
    t.field(room_history.max_players);
    t.field(room_history.public_props);
    t.field(room_history.private_props);
    t.field(room_history.player_logs);
    t.field(room_history.created);
    t.field(room_history.closed);
  },
});
