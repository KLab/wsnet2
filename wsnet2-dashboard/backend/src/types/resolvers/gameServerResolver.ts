import { extendType, idArg } from "nexus";
import { Context } from "../../context";

export const gameServerQuery = extendType({
  type: "Query",
  definition(t) {
    t.list.field("gameServers", {
      type: "game_server",
      description: "Get all game servers",
      resolve(_, __, ctx: Context) {
        return ctx.prisma.game_server.findMany({
          take: Number(process.env.GRAPHQL_RESULT_MAX_SIZE),
        });
      },
    });

    t.field("gameServerById", {
      type: "game_server",
      description: "Get unique game server by id",
      args: {
        id: idArg(),
      },
      resolve(_, { id }, ctx: Context) {
        return ctx.prisma.game_server.findUnique({
          where: { id: Number(id) },
        });
      },
    });
  },
});
