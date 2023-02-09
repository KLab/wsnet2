import { Module, VuexModule, Action, getModule } from "vuex-module-decorators";
import { store } from ".";
import settings from "./settings";
import { encode, decodeAsync } from "@msgpack/msgpack";

export interface RoomInfoReq {
  appId: string;
  roomId: string;
  hostId: number;
}

export interface KickReq {
  appId: string;
  roomId: string;
  hostId: number;
  clientId: string;
}

export type Unmarshaled = [unknown, object | null];

export interface RoomNumber {
  number: number;
}

export interface Timestamp {
  timestamp?: any;
}

export interface RoomInfo {
  roomInfo:
    | {
        id: string | undefined;
        appId: string | undefined;
        hostId: number | undefined;
        visible: boolean | undefined;
        joinable: boolean | undefined;
        watchable: boolean | undefined;
        number: RoomNumber | undefined;
        searchGroup: number | undefined;
        maxPlayers: number | undefined;
        players: number | undefined;
        watchers: number | undefined;
        publicProps: Unmarshaled;
        privateProps: Unmarshaled;
        created: Timestamp | undefined;
      }
    | undefined;
  clientInfosList: ClientInfo[];
  masterId: string;
}

export interface ClientInfo {
  id: string;
  isHub: boolean;
  props: Unmarshaled;
}

@Module({ dynamic: true, namespaced: true, name: "roomInfo", store: store })
class RoomInfoModule extends VuexModule {
  @Action
  async fetch(args: RoomInfoReq): Promise<RoomInfo> {
    const serverAddress = settings.serverAddress
      ? settings.serverAddress
      : import.meta.env.VITE_DEFAULT_SERVER_URI;
    const response = await fetch(`${serverAddress}/game/getRoomInfo`, {
      method: "POST",
      mode: "cors",
      headers: {
        accept: "application/msgpack",
        "content-type": "application/msgpack",
      },
      body: encode(args),
    });

    if (response.ok && response.body != null) {
      const result = await decodeAsync(response.body);
      return result as RoomInfo;
    } else {
      let message: string;
      if (response.body != null) {
        const err = await decodeAsync(response.body);
        message = (err as any)["details"];
      } else {
        message = "Failed to fetch room info!";
      }

      throw Error(message);
    }
  }

  @Action
  async kick(args: KickReq): Promise<void> {
    const serverAddress = settings.serverAddress
      ? settings.serverAddress
      : import.meta.env.VITE_DEFAULT_SERVER_URI;
    const response = await fetch(`${serverAddress}/game/kick`, {
      method: "POST",
      mode: "cors",
      headers: {
        accept: "application/msgpack",
        "content-type": "application/msgpack",
      },
      body: encode(args),
    });

    if (!response.ok) {
      let message: string;
      if (response.body != null) {
        const err = await decodeAsync(response.body);
        message = (err as any)["details"];
      } else {
        message = "Failed to kick!";
      }

      throw Error(message);
    }
  }
}

export default getModule(RoomInfoModule);
