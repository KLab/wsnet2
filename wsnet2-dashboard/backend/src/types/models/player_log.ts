import { objectType } from "nexus";
import { player_log } from "nexus-prisma";

export const playerLogModel = objectType({
  name: player_log.$name,
  description: player_log.$description,
  definition(t) {
    t.field(player_log.id);
    t.field(player_log.room_id);
    t.field(player_log.player_id);
    t.field(player_log.message);
    t.field(player_log.datetime);
  },
});
