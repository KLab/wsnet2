import { Module, VuexModule, Action, getModule } from "vuex-module-decorators";
import { decodeAsync } from "@msgpack/msgpack";
import { store } from ".";
import settings from "./settings";

export interface Overview {
  rooms: [{ host_id: number; hostname: string; num: number }];
  servers: [{ NApp: number; NGameServer: number; NHubServer: number }];
}

@Module({ dynamic: true, namespaced: true, name: "overview", store: store })
class OverviewModule extends VuexModule {
  @Action
  async fetch(): Promise<Overview> {
    const serverAddress = settings.serverAddress
      ? settings.serverAddress
      : import.meta.env.VITE_DEFAULT_SERVER_URI;
    const response = await fetch(`${serverAddress}/overview`, {
      method: "GET",
      mode: "cors",
      headers: {
        accept: "application/msgpack",
      },
    });

    if (response.ok && response.body != null) {
      const result = await decodeAsync(response.body);
      return result as Overview;
    } else {
      let message = "Failed to fetch overview!";
      if (response.body != null) {
        const err = await decodeAsync(response.body);
        message = (err as any)["details"];
      }
      throw Error(message);
    }
  }
}

export default getModule(OverviewModule);
