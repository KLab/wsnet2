import {
  Module,
  VuexModule,
  Mutation,
  Action,
  getModule,
} from "vuex-module-decorators";
import apolloClient from "../apolloClient";
import gql from "graphql-tag";
import { store } from ".";

export interface Server {
  id: string;
  hostname: string;
  public_name: string;
  grpc_port: number;
  ws_port: number;
  status: number;
  heartbeat: bigint;
}

@Module({ dynamic: true, namespaced: true, name: "servers", store: store })
class ServersModule extends VuexModule {
  gameServers = Array<Server>();
  hubServers = Array<Server>();

  @Mutation
  setGameServers(servers: Server[]) {
    this.gameServers = servers;
  }

  @Mutation
  setHubServers(servers: Server[]) {
    this.hubServers = servers;
  }

  @Action({ commit: "setGameServers" })
  async fetchGameServers(useCache: boolean): Promise<Server[]> {
    const response = await apolloClient.query({
      query: gql`
        query gameServerQuery {
          gameServers {
            id
            hostname
            public_name
            grpc_port
            ws_port
            status
            heartbeat
          }
        }
      `,
      fetchPolicy: useCache ? "cache-first" : "network-only",
    });

    if (response.error) throw Error(response.error.message);
    return response.data.gameServers as Server[];
  }

  @Action({ commit: "setHubServers" })
  async fetchHubServers(useCache: boolean): Promise<Server[]> {
    const response = await apolloClient.query({
      query: gql`
        query hubServerQuery {
          hubServers {
            id
            hostname
            public_name
            grpc_port
            ws_port
            status
            heartbeat
          }
        }
      `,
      fetchPolicy: useCache ? "cache-first" : "network-only",
    });

    if (response.error) throw Error(response.error.message);
    return response.data.hubServers as Server[];
  }
}

export default getModule(ServersModule);
