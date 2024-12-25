import { extendType, idArg } from "nexus";
import { Context } from "../../context.js";

export const hubServerQuery = extendType({
  type: "Query",
  definition(t) {
    t.list.field("hubServers", {
      type: "hub_server",
      description: "Get all hub servers",
      resolve(_, __, ctx: Context) {
        return ctx.prisma.hub_server.findMany({
          take: Number(process.env.GRAPHQL_RESULT_MAX_SIZE),
        });
      },
    });

    t.field("hubServerById", {
      type: "hub_server",
      description: "Get unique hub server by id",
      args: {
        id: idArg(),
      },
      resolve(_, { id }, ctx: Context) {
        return ctx.prisma.hub_server.findUnique({
          where: { id: Number(id) },
        });
      },
    });
  },
});
