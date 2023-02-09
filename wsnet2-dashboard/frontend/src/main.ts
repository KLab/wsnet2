// base
import { createApp } from "vue";

// views
import App from "@/App.vue";

// font
import "vfonts/Lato.css";

// plugins
import { store } from "./store";
import router from "./router";

const app = createApp(App);
app.use(router);
app.use(store);
app.mount("#app");

export default app;
