import { Module, VuexModule, Action, getModule } from "vuex-module-decorators";
import { store } from ".";
import settings from "./settings";

export interface Overview {
  rooms: [{ host_id: number; hostname: string; num: number }];
  servers: [{ NApp: number; NGameServer: number; NHubServer: number }];
}

@Module({ dynamic: true, namespaced: true, name: "overview", store: store })
class OverviewModule extends VuexModule {
  //一覧取得
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

  // パッケージのバージョンを取得
  @Action
  async fetchVersion(): Promise<string> {
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
      return result["version"] as string;
    } else {
      let message = "Failed to fetch version!";
      if (response.body != null) {
        const err = await response.json();
        message = (err as any)["details"];
      }
      throw Error(message);
    }
  }

  // GraphQLの最大取得件数を取得
  @Action
  async fetchGraphqlResultLimit(): Promise<number> {
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
      return result["limit"] as number;
    } else {
      let message = "Failed to fetch graphql_result_limit!";
      if (response.body != null) {
        const err = await response.json();
        message = (err as any)["details"];
      }
      throw Error(message);
    }
  }
}

export default getModule(OverviewModule);
