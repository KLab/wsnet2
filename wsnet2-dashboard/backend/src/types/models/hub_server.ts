import { objectType } from "nexus";
import np from "nexus-prisma";
const { hub_server } = np;

export const hubServerModel = objectType({
  name: hub_server.$name,
  description: hub_server.$description,
  definition(t) {
    t.field(hub_server.id);
    t.field(hub_server.hostname);
    t.field(hub_server.public_name);
    t.field(hub_server.grpc_port);
    t.field(hub_server.ws_port);
    t.field(hub_server.status);
    t.field(hub_server.heartbeat);
  },
});
