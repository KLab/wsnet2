import express from "express";
import msgpack from "express-msgpack";
import cors from "cors";
import { ApolloServer } from "apollo-server-express";
// local imports
import { schema } from "./schema";
import { createContext } from "./context";
// import routes
import game from "./routes/game";
import overview from "./routes/overview";

async function init() {
  // consts
  const app = express();
  const server = new ApolloServer({
    schema: schema,
    context: createContext,
  });

  // middlewares
  app.use(
    cors({
      origin: [String(process.env.FRONTEND_ORIGIN)],
    })
  );

  // app.use(msgpack());
  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));
  await server.start();
  server.applyMiddleware({ app, path: "/graphql" });

  // routes
  app.use("/game", game);
  app.use("/overview", overview);
  return app;
}

// start server
init()
  .then((app) => {
    app.listen({
      port: process.env.SERVER_PORT,
      host: "0.0.0.0",
      callback: () => {
        console.log(`Start on port ${String(process.env.SERVER_PORT)}.`);
      },
    });
  })
  .catch((err: Error) => {
    console.log(err.message);
  });
