import { Unmarshaled } from "./roomInfo";
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

export interface RoomHistory {
  id: string;
  app_id: string;
  room_id: string;
  host_id: number;
  number: number;
  search_group: number;
  max_players: number;
  public_props: Unmarshaled;
  private_props: Unmarshaled;
  player_logs: object;
  created: string;
  closed: string;
}

export interface FetchRoomHistoriesInput {
  appId: string[] | undefined | null;
  roomId: string | undefined | null;
  hostId: number | undefined | null;
  number: number | undefined | null;
  searchGroup: number | undefined | null;
  maxPlayers: number | undefined | null;
  createdBefore: string | undefined | null;
  createdAfter: string | undefined | null;
  closedBefore: string | undefined | null;
  closedAfter: string | undefined | null;
  useCache: boolean | undefined | null;
}

@Module({
  dynamic: true,
  namespaced: true,
  name: "roomHistories",
  store: store,
})
class RoomHistoriesModule extends VuexModule {
  roomHistories = Array<RoomHistory>();

  @Mutation
  setRoomHistories(roomHistories: RoomHistory[]) {
    this.roomHistories = roomHistories;
  }

  @Action({ commit: "setRoomHistories" })
  async fetch(args: FetchRoomHistoriesInput): Promise<RoomHistory[]> {
    const response = await apolloClient.query({
      query: gql`
        query roomHistoryQuery(
          $appId: [String]
          $roomId: String
          $hostId: Int
          $number: Int
          $searchGroup: Int
          $maxPlayers: Int
          $createdBefore: DateTime
          $createdAfter: DateTime
          $closedBefore: DateTime
          $closedAfter: DateTime
        ) {
          roomHistories(
            app_id: $appId
            room_id: $roomId
            host_id: $hostId
            number: $number
            search_group: $searchGroup
            max_players: $maxPlayers
            created_before: $createdBefore
            created_after: $createdAfter
            closed_before: $closedBefore
            closed_after: $closedAfter
          ) {
            id
            app_id
            host_id
            room_id
            number
            search_group
            max_players
            public_props
            private_props
            player_logs
            created
            closed
          }
        }
      `,
      variables: {
        appId: args.appId,
        roomId: args.roomId,
        hostId: args.hostId,
        number: args.number,
        searchGroup: args.searchGroup,
        maxPlayers: args.maxPlayers,
        createdBefore: args.createdBefore,
        createdAfter: args.createdAfter,
        closedBefore: args.closedBefore,
        closedAfter: args.closedAfter,
      },
      fetchPolicy: args.useCache ? "cache-first" : "network-only",
    });

    if (response.error) throw Error(response.error.message);
    return response.data.roomHistories as RoomHistory[];
  }
}

export default getModule(RoomHistoriesModule);
