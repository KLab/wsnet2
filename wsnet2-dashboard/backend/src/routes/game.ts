import * as express from "express";
import { Client, createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import { RoomNumber } from "../pb/roominfo_pb.js";
import { Timestamp } from "../pb/timestamp_pb.js";
import {
  Game,
  GetRoomInfoReq,
  GetRoomInfoRes,
  KickReq,
  Empty,
} from "../pb/gameservice_pb.js";

import { PrismaClient } from "@prisma/client";
import * as binary from "../plugins/binary.js";

export interface RESTRoomInfoReq {
  appId: string;
  roomId: string;
  hostId: number;
}

export interface RESTKickReq {
  appId: string;
  roomId: string;
  hostId: number;
  clientId: string;
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
        publicProps: binary.Unmarshaled;
        privateProps: binary.Unmarshaled;
        created: Timestamp | undefined;
      }
    | undefined;
  clientInfosList: ClientInfo[];
  masterId: string;
}

export interface ClientInfo {
  id: string;
  isHub: boolean;
  props: binary.Unmarshaled;
}

function ConvertRoomInfo(src: GetRoomInfoRes): RoomInfo {
  const room = src.roomInfo;
  return {
    roomInfo: {
      id: room?.id,
      appId: room?.appId,
      hostId: room?.hostId,
      visible: room?.visible,
      joinable: room?.joinable,
      watchable: room?.watchable,
      number: room?.number,
      searchGroup: room?.searchGroup,
      maxPlayers: room?.maxPlayers,
      players: room?.players,
      watchers: room?.watchers,
      publicProps: room
        ? binary.UnmarshalRecursive(room?.publicProps)
        : [null, {}],
      privateProps: room
        ? binary.UnmarshalRecursive(room?.privateProps)
        : [null, {}],
      created: room?.created,
    },
    clientInfosList: src.clientInfos.map((info) => {
      return {
        id: info.id,
        isHub: info.isHub,
        props: binary.UnmarshalRecursive(info.props),
      };
    }),
    masterId: src.masterId,
  };
}

const prisma = new PrismaClient();
const serversCache: { [id: string]: Client<typeof Game> } = {};
const router = express.Router();

async function getCachedGameClient(hostId: number) {
  if (Object.prototype.hasOwnProperty.call(serversCache, hostId.toString()))
    return serversCache[hostId];
  const server = await prisma.game_server.findUnique({
    where: {
      id: hostId,
    },
  });

  if (server == null)
    throw new Error("Can't find game server by given hostId!");

  const gameClient = createClient(Game, createGrpcTransport({
    baseUrl: `${server.hostname}:${server.grpc_port}`,
    interceptors: [],
  }));

  serversCache[hostId.toString()] = gameClient;
  return gameClient;
}

async function getRoomInfo(args: RESTRoomInfoReq) {
  //return new Promise<GetRoomInfoRes>((resolve, reject) => {
  const gameClient = await getCachedGameClient(args.hostId);
  const req = {
    appId: args.appId,
    roomId: args.roomId,
  };
  return await gameClient.getRoomInfo(req);
}

async function kick(args: RESTKickReq) {
  const gameClient = await getCachedGameClient(args.hostId);
  const req = {
    appId: args.appId,
    roomId: args.roomId,
    clientId: args.clientId,
  };
  return await gameClient.kick(req);
}

// router implementations
router.post(
  "/getRoomInfo",
  (
    req: express.Request<unknown, RoomInfo, RESTRoomInfoReq>,
    res: express.Response
  ) => {
    getRoomInfo(req.body)
      .then((roomInfo) => {
        res.status(200).send(ConvertRoomInfo(roomInfo));
      })
      .catch((err: Error) => res.status(500).send({ details: err.message }));
  }
);

router.post(
  "/kick",
  (
    req: express.Request<unknown, Empty, RESTKickReq>,
    res: express.Response
  ) => {
    kick(req.body)
      .then((info) => {
        res.status(200).send(info);
      })
      .catch((err: Error) => res.status(500).send({ details: err.message }));
  }
);

export default router;
