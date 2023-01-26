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

export interface App {
  id: string;
  name: string;
  key: string;
}

@Module({ dynamic: true, namespaced: true, name: "apps", store: store })
class AppsModule extends VuexModule {
  apps = Array<App>();

  @Mutation
  setApps(apps: App[]) {
    this.apps = apps;
  }

  @Action({ commit: "setApps" })
  async fetch(useCache: boolean): Promise<App[]> {
    const response = await apolloClient.query({
      query: gql`
        query roomQuery {
          apps {
            id
            name
            key
          }
        }
      `,
      fetchPolicy: useCache ? "cache-first" : "network-only",
    });

    if (response.error) throw Error(response.error.message);
    return response.data.apps as App[];
  }

  @Action
  async createApp(args: App): Promise<App> {
    const response = await apolloClient.mutate({
      mutation: gql`
        mutation createApp($id: ID!, $name: String, $key: String) {
          createApp(id: $id, name: $name, key: $key) {
            id
            name
            key
          }
        }
      `,
      variables: args,
    });

    if (response.errors) throw Error(response.errors.join("\n"));
    return response.data.createApp as App;
  }
}

export default getModule(AppsModule);
