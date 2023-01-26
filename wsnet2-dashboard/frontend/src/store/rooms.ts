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

export interface Room {
  id: string;
  app_id: string;
  host_id: number;
  visible: number;
  joinable: number;
  watchable: number;
  number: number;
  search_group: number;
  max_players: number;
  players: number;
  watchers: number;
  props: Unmarshaled;
  created: string;
}

export interface RoomCreateInput {
  createRoomId: string;
  appId: string;
  hostId: number;
  visible: number;
  joinable: number;
  watchable: number;
  searchGroup: number;
  maxPlayers: number;
  players: number;
  watchers: number;
}

export interface FetchRoomsInput {
  appId: string[] | undefined | null;
  hostId: number | undefined | null;
  visible: number | undefined | null;
  joinable: number | undefined | null;
  watchable: number | undefined | null;
  number: number | undefined | null;
  searchGroup: number | undefined | null;
  maxPlayers: number | undefined | null;
  playersMin: number | undefined | null;
  playersMax: number | undefined | null;
  watchersMin: number | undefined | null;
  watchersMax: number | undefined | null;
  createdBefore: string | undefined | null;
  createdAfter: string | undefined | null;
  useCache: boolean | undefined | null;
}

export interface MockRoomsCreateInput {
  appIds: string[];
  count: number;
}

/**
 * Random integer between min and max. Min and max are both inclusive
 * @param {number} min - Lower bound.
 * @param {number} max - Higher bound.
 */
function randomIntFromInterval(min: number, max: number) {
  return Math.floor(Math.random() * (max - min + 1) + min);
}

async function createRoom(args: RoomCreateInput) {
  const response = await apolloClient.mutate({
    mutation: gql`
      mutation roomCreate(
        $createRoomId: ID!
        $appId: String!
        $hostId: Int!
        $visible: Int!
        $joinable: Int!
        $watchable: Int!
        $searchGroup: Int!
        $maxPlayers: Int!
        $players: Int!
        $watchers: Int!
      ) {
        createRoom(
          id: $createRoomId
          app_id: $appId
          host_id: $hostId
          visible: $visible
          joinable: $joinable
          watchable: $watchable
          search_group: $searchGroup
          max_players: $maxPlayers
          players: $players
          watchers: $watchers
        ) {
          id
          app_id
          host_id
          visible
          joinable
          watchable
          number
          search_group
          max_players
          players
          watchers
          props
          created
        }
      }
    `,
    variables: args,
  });

  if (response.errors) throw Error(response.errors.join("\n"));
  return response.data.createRoom as Room;
}

@Module({ dynamic: true, namespaced: true, name: "rooms", store: store })
class RoomsModule extends VuexModule {
  rooms = Array<Room>();

  @Mutation
  setRooms(rooms: Room[]) {
    this.rooms = rooms;
  }

  @Action({ commit: "setRooms" })
  async fetch(args: FetchRoomsInput): Promise<Room[]> {
    const response = await apolloClient.query({
      query: gql`
        query roomQuery(
          $appId: [String]
          $hostId: Int
          $visible: Int
          $joinable: Int
          $watchable: Int
          $number: Int
          $searchGroup: Int
          $maxPlayers: Int
          $playersMin: Int
          $playersMax: Int
          $watchersMin: Int
          $watchersMax: Int
          $createdBefore: DateTime
          $createdAfter: DateTime
        ) {
          rooms(
            app_id: $appId
            host_id: $hostId
            visible: $visible
            joinable: $joinable
            watchable: $watchable
            number: $number
            search_group: $searchGroup
            max_players: $maxPlayers
            players_min: $playersMin
            players_max: $playersMax
            watchers_min: $watchersMin
            watchers_max: $watchersMax
            created_before: $createdBefore
            created_after: $createdAfter
          ) {
            id
            app_id
            host_id
            visible
            joinable
            watchable
            number
            search_group
            max_players
            players
            watchers
            props
            created
          }
        }
      `,
      variables: {
        appId: args.appId,
        hostId: args.hostId,
        visible: args.visible,
        joinable: args.joinable,
        watchable: args.watchable,
        number: args.number,
        searchGroup: args.searchGroup,
        maxPlayers: args.maxPlayers,
        playersMin: args.playersMin,
        watchersMin: args.watchersMin,
        playersMax: args.playersMax,
        watchersMax: args.watchersMax,
        createdBefore: args.createdBefore,
        createdAfter: args.createdAfter,
      },
      fetchPolicy: args.useCache ? "cache-first" : "network-only",
    });

    if (response.error) throw Error(response.error.message);
    return response.data.rooms as Room[];
  }

  @Action
  async createMockRooms(args: MockRoomsCreateInput) {
    const arg: RoomCreateInput = {
      createRoomId: "",
      appId: "",
      hostId: 0,
      visible: 0,
      joinable: 0,
      watchable: 0,
      searchGroup: 0,
      maxPlayers: 32,
      players: 5,
      watchers: 5,
    };

    const promises = Array<Promise<Room>>();

    for (let index = 0; index < args.appIds.length; index++) {
      arg.appId = args.appIds[index];

      for (let index2 = 0; index2 < args.count; index2++) {
        arg.createRoomId = `${arg.appId}_${index2}`;
        arg.hostId = randomIntFromInterval(0, 10001);
        arg.visible = randomIntFromInterval(0, 1);
        arg.joinable = randomIntFromInterval(0, 1);
        arg.watchable = randomIntFromInterval(0, 1);
        arg.searchGroup = randomIntFromInterval(0, 100);
        arg.players = randomIntFromInterval(1, arg.maxPlayers);
        arg.watchers = randomIntFromInterval(0, arg.maxPlayers);
        promises.push(createRoom(Object.assign({}, arg)));
      }
    }

    const results = await Promise.all(promises);
    return results;
  }
}

export default getModule(RoomsModule);
