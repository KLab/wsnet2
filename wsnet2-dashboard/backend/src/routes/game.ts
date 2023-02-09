import * as express from "express";
import { RoomNumber } from "../pb/roominfo_pb";
import { Timestamp } from "../pb/timestamp_pb";
import { GameClient } from "../pb/gameservice_grpc_pb";
import { credentials } from "@grpc/grpc-js";
import {
  GetRoomInfoReq,
  GetRoomInfoRes,
  KickReq,
  Empty,
} from "../pb/gameservice_pb";

import { PrismaClient } from "@prisma/client";
import binary = require("../plugins/binary.js");

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
        number: RoomNumber.AsObject | undefined;
        searchGroup: number | undefined;
        maxPlayers: number | undefined;
        players: number | undefined;
        watchers: number | undefined;
        publicProps: binary.Unmarshaled;
        privateProps: binary.Unmarshaled;
        created: Timestamp.AsObject | undefined;
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
  const room = src.getRoomInfo();
  return {
    roomInfo: {
      id: room?.getId(),
      appId: room?.getAppId(),
      hostId: room?.getHostId(),
      visible: room?.getVisible(),
      joinable: room?.getJoinable(),
      watchable: room?.getWatchable(),
      number: room?.getNumber()?.toObject(),
      searchGroup: room?.getSearchGroup(),
      maxPlayers: room?.getMaxPlayers(),
      players: room?.getPlayers(),
      watchers: room?.getWatchers(),
      publicProps: room
        ? binary.UnmarshalRecursive(room.getPublicProps_asU8())
        : [null, {}],
      privateProps: room
        ? binary.UnmarshalRecursive(room.getPrivateProps_asU8())
        : [null, {}],
      created: room?.getCreated()?.toObject(),
    },
    clientInfosList: src.getClientInfosList().map((info) => {
      return {
        id: info.getId(),
        isHub: info.getIsHub(),
        props: binary.UnmarshalRecursive(info.getProps_asU8()),
      };
    }),
    masterId: src.getMasterId(),
  };
}

const prisma = new PrismaClient();
const serversCache: { [id: string]: GameClient } = {};
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
  const gameClient = new GameClient(
    `${server.hostname}:${server.grpc_port}`,
    credentials.createInsecure()
  );

  serversCache[hostId.toString()] = gameClient;
  return gameClient;
}

async function getRoomInfo(args: RESTRoomInfoReq) {
  return new Promise<GetRoomInfoRes>((resolve, reject) => {
    getCachedGameClient(args.hostId)
      .then((gameClient) => {
        const req = new GetRoomInfoReq();
        req.setAppId(args.appId);
        req.setRoomId(args.roomId);
        gameClient.getRoomInfo(req, (error, response) => {
          if (error) reject(error);
          resolve(response);
        });
      })
      .catch((err) => reject(err));
  });
}

async function kick(args: RESTKickReq) {
  return new Promise<Empty>((resolve, reject) => {
    getCachedGameClient(args.hostId)
      .then((gameClient) => {
        const req = new KickReq();
        req.setAppId(args.appId);
        req.setRoomId(args.roomId);
        req.setClientId(args.clientId);
        gameClient.kick(req, (error, response) => {
          if (error) reject(error);
          resolve(response);
        });
      })
      .catch((err) => reject(err));
  });
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
    req: express.Request<unknown, Empty.AsObject, RESTKickReq>,
    res: express.Response
  ) => {
    kick(req.body)
      .then((info) => {
        res.status(200).send(info.toObject());
      })
      .catch((err: Error) => res.status(500).send({ details: err.message }));
  }
);

export default router;
