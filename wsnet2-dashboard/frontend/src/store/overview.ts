import {
  Module,
  VuexModule,
  Action,
  Mutation,
  getModule,
} from "vuex-module-decorators";
import { store } from ".";
import settings from "./settings";

export interface Overview {
  rooms: [{ host_id: number; hostname: string; num: number }];
  servers: [{ NApp: number; NGameServer: number; NHubServer: number }];
}

@Module({
  dynamic: true,
  namespaced: true,
  name: "overview",
  store: store,
})
class OverviewModule extends VuexModule {
  serverVersion = "";
  graphqlResultLimit = 0;

  @Action
  async fetch(): Promise<Overview> {
    const serverAddress = settings.serverAddress
      ? settings.serverAddress
      : import.meta.env.VITE_DEFAULT_SERVER_URI;
    const response = await fetch(`${serverAddress}/overview`, {
      method: "GET",
      mode: "cors",
      headers: {
        accept: "application/json",
      },
    });

    if (response.ok && response.body != null) {
      const result = await response.json();
      return result as Overview;
    } else {
      let message = "Failed to fetch overview!";
      if (response.body != null) {
        const err = await response.json();
        message = (err as any)["details"];
      }
      throw Error(message);
    }
  }

  @Mutation
  setServerVersion(version: string) {
    this.serverVersion = version;
  }

  @Mutation
  setGraphqlResultLimit(limit: number) {
    this.graphqlResultLimit = limit;
  }

  @Action
  async fetchServerVersion() {
    const serverAddress = settings.serverAddress
      ? settings.serverAddress
      : import.meta.env.VITE_DEFAULT_SERVER_URI;
    const response = await fetch(`${serverAddress}/overview/version`, {
      method: "GET",
      mode: "cors",
      headers: {
        accept: "application/json",
      },
    });
    if (response.ok && response.body != null) {
      const result = await response.json();
      this.context.commit("setServerVersion", result["version"] as string);
    } else {
      const err = await response.json();
      throw Error((err as any)["details"]);
    }
  }

  @Action
  async fetchGraphqlResultLimit() {
    const serverAddress = settings.serverAddress
      ? settings.serverAddress
      : import.meta.env.VITE_DEFAULT_SERVER_URI;
    const response = await fetch(
      `${serverAddress}/overview/graphql_result_limit`,
      {
        method: "GET",
        mode: "cors",
        headers: {
          accept: "application/json",
        },
      }
    );
    if (response.ok && response.body != null) {
      const result = await response.json();
      this.context.commit("setGraphqlResultLimit", result["limit"] as number);
    } else {
      const err = await response.json();
      throw Error((err as any)["details"]);
    }
  }
}

export default getModule(OverviewModule);
