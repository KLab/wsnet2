import {
  extendType,
  idArg,
  stringArg,
  intArg,
  arg,
  nonNull,
  list,
} from "nexus";
import { Context } from "../../context.js";

interface IConditions {
  OR?: { app_id: string }[];
  room_id?: string;
  host_id?: number;
  number?: number;
  search_group?: number;
  max_players?: number;
  created?: {
    lte?: Date;
    gte?: Date;
  };
  closed?: {
    lte?: Date;
    gte?: Date;
  };
}

export const roomHistoryQuery = extendType({
  type: "Query",
  definition(t) {
    t.list.field("roomHistories", {
      type: "room_history",
      description: "Get all room hitories",
      args: {
        app_id: list(stringArg()),
        room_id: stringArg(),
        host_id: intArg(),
        number: intArg(),
        search_group: intArg(),
        max_players: intArg(),
        created_before: arg({
          type: "DateTime",
        }),
        created_after: arg({
          type: "DateTime",
        }),
        closed_before: arg({
          type: "DateTime",
        }),
        closed_after: arg({
          type: "DateTime",
        }),
      },
      resolve(
        _,
        {
          app_id,
          room_id,
          host_id,
          number,
          search_group,
          max_players,
          created_before,
          created_after,
          closed_before,
          closed_after,
        },
        ctx: Context
      ) {
        const conditions: IConditions = {};
        if (app_id != null)
          conditions.OR = (app_id as string[]).map((id) => {
            return {
              app_id: id,
            };
          });

        if (room_id != null) conditions.room_id = String(room_id);
        if (host_id != null) conditions.host_id = Number(host_id);
        if (number != null) conditions.number = Number(number);
        if (search_group != null)
          conditions.search_group = Number(search_group);
        if (max_players != null) conditions.max_players = Number(max_players);
        if (created_after != null || created_before != null) {
          conditions.created = {};
          if (created_after != null)
            conditions.created.gte = new Date(String(created_after));
          if (created_before != null)
            conditions.created.lte = new Date(String(created_before));
        }
        if (closed_after != null || closed_before != null) {
          conditions.closed = {};
          if (closed_after != null)
            conditions.closed.gte = new Date(String(closed_after));
          if (closed_before != null)
            conditions.closed.lte = new Date(String(closed_before));
        }

        return ctx.prisma.room_history.findMany({
          where: conditions,
          take: Number(process.env.GRAPHQL_RESULT_MAX_SIZE),
        });
      },
    });

    t.field("roomHistoryById", {
      type: "room_history",
      description: "Get unique room history by id",
      args: {
        id: nonNull(idArg()),
      },
      resolve(_, { id }, ctx: Context) {
        return ctx.prisma.room_history.findUnique({
          where: { id: Number(id) },
        });
      },
    });
  },
});
