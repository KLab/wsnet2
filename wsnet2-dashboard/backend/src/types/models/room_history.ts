import { objectType } from "nexus";
import np from "nexus-prisma";
const { room_history } = np;

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
    t.field(room_history.created);
    t.field(room_history.closed);
    t.list.field("player_logs", {
      type: "player_log",
      resolve(parent, _args, ctx) {
        return ctx.prisma.player_log.findMany({
          where: { room_id: parent.room_id },
          orderBy: { id: "asc" },
          take: Number(process.env.GRAPHQL_RESULT_MAX_SIZE),
        });
      },
    });
  },
});
