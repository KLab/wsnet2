import express from "express";
import cors from "cors";
import { ApolloServer } from "apollo-server-express";
// local imports
import { schema } from "./schema.js";
import { createContext } from "./context.js";
// import routes
import game from "./routes/game.js";
import overview from "./routes/overview.js";

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
      // host: "0.0.0.0",
      callback: () => {
        console.log(`Start on port ${String(process.env.SERVER_PORT)}.`);
      },
    });
  })
  .catch((err: Error) => {
    console.log(err.message);
  });
