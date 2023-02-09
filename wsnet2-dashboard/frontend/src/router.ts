import { createRouter, createWebHashHistory } from "vue-router";
import Home from "@/views/Home.vue";
import Apps from "@/views/Apps.vue";
import GameServers from "@/views/GameServers.vue";
import HubServers from "@/views/HubServers.vue";
import Rooms from "@/views/Rooms.vue";
import RoomHistories from "./views/RoomHistories.vue";
import Settings from "./views/Settings.vue";

// vue-router
const routes = [
  {
    path: "/",
    component: Home,
    meta: {
      title: "WSNet2 - Dashboard",
    },
  },
  {
    path: "/Apps",
    component: Apps,
    meta: {
      title: "WSNet2 - Dashboard: Apps",
    },
  },
  {
    path: "/GameServers",
    component: GameServers,
    meta: {
      title: "WSNet2 - Dashboard: Game Servers",
    },
  },
  {
    path: "/HubServers",
    component: HubServers,
    meta: {
      title: "WSNet2 - Dashboard: Hub Servers",
    },
  },
  {
    path: "/Rooms",
    component: Rooms,
    meta: {
      title: "WSNet2 - Dashboard: Rooms",
    },
  },
  {
    path: "/RoomHistories",
    component: RoomHistories,
    meta: {
      title: "WSNet2 - Dashboard: Room Histories",
    },
  },
  {
    path: "/Settings",
    component: Settings,
    meta: {
      title: "WSNet2 - Dashboard: Settings",
    },
  },
];
const router = createRouter({
  // 4. Provide the history implementation to use. We are using the hash history for simplicity here.
  history: createWebHashHistory(),
  routes, // short for `routes: routes`
});

// This callback runs before every route change, including on page load.
router.beforeEach((to, from, next) => {
  // This goes through the matched routes from last to first, finding the closest route with a title.
  // e.g., if we have `/some/deep/nested/route` and `/some`, `/deep`, and `/nested` have titles,
  // `/nested`'s will be chosen.
  const nearestWithTitle = to.matched
    .slice()
    .reverse()
    .find((r) => r.meta && r.meta.title);

  // Find the nearest route element with meta tags.
  const nearestWithMeta = to.matched
    .slice()
    .reverse()
    .find((r) => r.meta && r.meta.metaTags);

  const previousNearestWithMeta = from.matched
    .slice()
    .reverse()
    .find((r) => r.meta && r.meta.metaTags);

  // If a route with a title was found, set the document (page) title to that value.
  if (nearestWithTitle) {
    document.title = nearestWithTitle.meta.title as string;
  } else if (previousNearestWithMeta) {
    document.title = previousNearestWithMeta.meta.title as string;
  }

  next();
});

export default router;
