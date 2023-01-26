import {
  Module,
  VuexModule,
  Mutation,
  getModule,
} from "vuex-module-decorators";
import { store } from ".";
import apolloClient from "../apolloClient";
import { createHttpLink } from "@apollo/client/core";

@Module({
  dynamic: true,
  namespaced: true,
  name: "settings",
  store: store,
  preserveState: localStorage.getItem("vuex") !== null,
})
class SettingsModule extends VuexModule {
  serverAddress = "";
  theme = "dark";

  @Mutation
  setServerAddress(address: string) {
    this.serverAddress = address;
    apolloClient.setLink(
      createHttpLink({
        uri: `${this.serverAddress}/graphql`,
        fetchOptions: {
          mode: "cors",
        },
      })
    );
  }

  @Mutation
  setTheme(theme: string) {
    this.theme = theme;
  }
}

export default getModule(SettingsModule);
