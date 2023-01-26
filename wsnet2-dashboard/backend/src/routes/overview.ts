import * as express from "express";
import mysql from "mysql2/promise";

const router = express.Router();
const sql = mysql.createPool({
  uri: process.env.DATABASE_URL,
  waitForConnections: true,
  connectionLimit: 10,
  queueLimit: 0,
});

//一覧取得
router.get("/", (req: express.Request, res: express.Response) => {
  Promise.all([
    sql.query(`
        SELECT room.host_id, game_server.hostname, COUNT(room.host_id) AS num 
            FROM room
            INNER JOIN game_server ON game_server.id = room.host_id
            GROUP BY room.host_id
      `),
    sql.query(`
        SELECT apps.num AS NApp, game_servers.num AS NGameServer, hub_servers.num AS NHubServer FROM
            (SELECT COUNT(id) AS num FROM app) AS apps,
            (SELECT COUNT(id) AS num FROM game_server) AS game_servers,
            (SELECT COUNT(id) AS num FROM hub_server) AS hub_servers
        `),
  ])
    .then(([[rooms], [servers]]) => {
      res.status(200).send({ rooms: rooms, servers: servers });
    })
    .catch((err: Error) => {
      res.status(500).send({ details: err.message });
    });
});

export default router;
