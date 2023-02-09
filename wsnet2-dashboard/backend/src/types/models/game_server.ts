import { objectType } from "nexus";
import { game_server } from "nexus-prisma";

export const gameServerModel = objectType({
  name: game_server.$name,
  description: game_server.$description,
  definition(t) {
    t.field(game_server.id);
    t.field(game_server.hostname);
    t.field(game_server.public_name);
    t.field(game_server.grpc_port);
    t.field(game_server.ws_port);
    t.field(game_server.status);
    t.field(game_server.heartbeat);
  },
});
