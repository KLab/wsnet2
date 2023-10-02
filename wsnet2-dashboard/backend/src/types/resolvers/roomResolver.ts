import {
  extendType,
  idArg,
  stringArg,
  intArg,
  arg,
  nonNull,
  list,
} from "nexus";
import { Context } from "../../context";
import { Prisma, room } from "@prisma/client";

interface IConditions {
  OR?: { app_id: string }[];
  host_id?: number;
  visible?: number;
  joinable?: number;
  watchable?: number;
  number?: number;
  search_group?: number;
  max_players?: number;
  players?: {
    lte?: number;
    gte?: number;
  };
  watchers?: {
    lte?: number;
    gte?: number;
  };
  created?: {
    lte?: Date;
    gte?: Date;
  };
}

export const roomQuery = extendType({
  type: "Query",
  definition(t) {
    t.field("roomById", {
      type: "room",
      description: "Get unique room by id",
      args: {
        id: nonNull(idArg()),
      },
      resolve(_, { id }, ctx: Context) {
        return ctx.prisma.room.findUnique({
          where: { id: String(id) },
        });
      },
    });

    t.list.field("rooms", {
      type: "room",
      description:
        "Query rooms by conditions. All parameters are optional. If no parameter is given, all rooms will be returned.",
      args: {
        app_id: list(stringArg()),
        host_id: intArg(),
        visible: intArg(),
        joinable: intArg(),
        watchable: intArg(),
        number: intArg(),
        search_group: intArg(),
        max_players: intArg(),
        players_min: intArg(),
        watchers_min: intArg(),
        players_max: intArg(),
        watchers_max: intArg(),
        created_before: arg({
          type: "DateTime",
        }),
        created_after: arg({
          type: "DateTime",
        }),
      },
      async resolve(
        _,
        {
          app_id,
          host_id,
          visible,
          joinable,
          watchable,
          number,
          search_group,
          max_players,
          players_min,
          watchers_min,
          players_max,
          watchers_max,
          created_before,
          created_after,
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

        if (host_id != null) conditions.host_id = Number(host_id);
        if (visible != null) conditions.visible = Number(visible);
        if (joinable != null) conditions.joinable = Number(joinable);
        if (watchable != null) conditions.watchable = Number(watchable);
        if (number != null) conditions.number = Number(number);
        if (search_group != null)
          conditions.search_group = Number(search_group);
        if (max_players != null) conditions.max_players = Number(max_players);
        if (players_min != null || players_max != null) {
          conditions.players = {};
          if (players_min != null) conditions.players.gte = Number(players_min);
          if (players_max != null) conditions.players.lte = Number(players_max);
        }
        if (watchers_min != null || watchers_max != null) {
          conditions.watchers = {};
          if (watchers_min != null)
            conditions.watchers.gte = Number(watchers_min);
          if (watchers_max != null)
            conditions.watchers.lte = Number(watchers_max);
        }
        if (created_after != null || created_before != null) {
          conditions.created = {};
          if (created_after != null)
            conditions.created.gte = new Date(String(created_after));
          if (created_before != null)
            conditions.created.lte = new Date(String(created_before));
        }

        // raw result
        const result: room[] = await ctx.prisma.room.findMany({
          where: conditions,
          take: Number(process.env.GRAPHQL_RESULT_MAX_SIZE),
        });

        return result;
      },
    });
  },
});

export const roomMutation = extendType({
  type: "Mutation",
  definition(t) {
    t.field("createRoom", {
      type: "room",
      description: "Create new room",
      args: {
        id: nonNull(idArg()),
        app_id: nonNull(stringArg()),
        host_id: nonNull(intArg()),
        visible: nonNull(intArg()),
        joinable: nonNull(intArg()),
        watchable: nonNull(intArg()),
        number: intArg(),
        search_group: nonNull(intArg()),
        max_players: nonNull(intArg()),
        players: nonNull(intArg()),
        watchers: nonNull(intArg()),
        props: arg({
          type: "Bytes",
          default: undefined,
        }),
        created: arg({
          type: "DateTime",
          default: new Date().toISOString(),
        }),
      },
      resolve(_, args: Prisma.roomCreateInput, ctx: Context) {
        return ctx.prisma.room.create({
          data: args,
        });
      },
    });
  },
});
