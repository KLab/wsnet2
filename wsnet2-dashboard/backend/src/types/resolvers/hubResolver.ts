import { extendType, idArg } from "nexus";
import { Context } from "../../context";

export const hubQuery = extendType({
  type: "Query",
  definition(t) {
    t.list.field("hubs", {
      type: "hub",
      description: "Get all hubs",
      resolve(_, __, ctx: Context) {
        return ctx.prisma.hub.findMany({
          take: Number(process.env.GRAPHQL_RESULT_MAX_SIZE),
        });
      },
    });

    t.field("hubById", {
      type: "hub",
      description: "Get unique hub by id",
      args: {
        id: idArg(),
      },
      resolve(_, { id }, ctx: Context) {
        return ctx.prisma.hub.findUnique({
          where: { id: Number(id) },
        });
      },
    });
  },
});
