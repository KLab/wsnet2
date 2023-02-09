import { extendType, idArg, stringArg, nonNull } from "nexus";
import { Context } from "../../context";
import { Prisma } from "@prisma/client";

export const appQuery = extendType({
  type: "Query",
  definition(t) {
    t.list.field("apps", {
      type: "app",
      description: "Get all apps",
      resolve(_, __, ctx: Context) {
        return ctx.prisma.app.findMany();
      },
    });

    t.field("appById", {
      type: "app",
      description: "Get unique app by id",
      args: {
        id: idArg(),
      },
      resolve(_, { id }, ctx: Context) {
        return ctx.prisma.app.findUnique({
          where: { id: String(id) },
        });
      },
    });
  },
});

export const appMutation = extendType({
  type: "Mutation",
  definition(t) {
    t.field("createApp", {
      type: "app",
      description: "Create new app",
      args: {
        id: nonNull(idArg()),
        name: stringArg(),
        key: stringArg(),
      },
      resolve(_, args: Prisma.appCreateInput, ctx: Context) {
        return ctx.prisma.app.create({
          data: args,
        });
      },
    });
  },
});
