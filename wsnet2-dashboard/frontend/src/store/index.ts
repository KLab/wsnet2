import { createStore } from "vuex";
import VuexPersistence from "vuex-persist";

const vuexLocal = new VuexPersistence({
  storage: window.localStorage,
  modules: ["settings"],
});

export const store = createStore({
  plugins: [vuexLocal.plugin],
});
