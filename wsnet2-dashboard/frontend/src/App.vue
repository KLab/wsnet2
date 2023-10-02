<script setup lang="ts">
import { ref, h, Component, onBeforeMount } from "vue";
import { useRouter } from "vue-router";
import {
  lightTheme,
  darkTheme,
  NConfigProvider,
  NDialogProvider,
  NMessageProvider,
  NLayout,
  NLayoutSider,
  NMenu,
  NIcon,
  MenuOption,
  NTag,
  NSpace,
} from "naive-ui";
import {
  AppFolder24Filled,
  BoxMultipleSearch24Regular,
  WeatherSunny16Regular,
  WeatherMoon16Filled,
  Home16Filled,
  History24Regular,
} from "@vicons/fluent";
import {
  GameConsole,
  ColorPalette,
  Settings,
  SettingsAdjust,
} from "@vicons/carbon";
import { DeviceHubFilled } from "@vicons/material";

import settings from "./store/settings";
import overview from "./store/overview";

import apolloClient from "./apolloClient";
import { createHttpLink } from "@apollo/client/core";

const collapsed = ref(true);
const activeKey = ref(null);
const appVersion = import.meta.env.PACKAGE_VERSION;

const theme = ref(darkTheme);
switch (settings.theme) {
  case "light":
    theme.value = lightTheme;
    break;
  case "dark":
    theme.value = darkTheme;
    break;

  default:
    break;
}

const router = useRouter();
const serverAddress = settings.serverAddress
  ? settings.serverAddress
  : import.meta.env.VITE_DEFAULT_SERVER_URI;

// go to settings if no server address is available
if (!serverAddress) {
  router.push("/Settings");
  alert("You need to set server address before using the dashboard!");
}

// update graphql server address
apolloClient.setLink(
  createHttpLink({
    uri: `${serverAddress}/graphql`,
    fetchOptions: {
      mode: "cors",
    },
  })
);

// update server info
onBeforeMount(async () => {
  // fetch server version
  await overview.fetchServerVersion();
  // fetch graphql result limit
  await overview.fetchGraphqlResultLimit();
});

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

function onMenuUpdate(key: string, item: MenuOption) {
  switch (key) {
    // routes
    case "home": {
      router.push("/");
      break;
    }
    case "apps": {
      router.push("/Apps");
      break;
    }
    case "gameServers": {
      router.push("/GameServers");
      break;
    }
    case "hubServers": {
      router.push("/HubServers");
      break;
    }
    case "rooms": {
      router.push("/Rooms");
      break;
    }
    case "roomHistories": {
      router.push("/RoomHistories");
      break;
    }
    // settings
    case "settings": {
      router.push("/Settings");
      break;
    }
    case "light": {
      theme.value = lightTheme;
      settings.setTheme("light");
      break;
    }
    case "dark": {
      theme.value = darkTheme;
      settings.setTheme("dark");
      break;
    }
  }
}

const menuOptions = [
  {
    label: "Home",
    key: "home",
    icon: renderIcon(Home16Filled),
  },
  {
    label: "Apps",
    key: "apps",
    icon: renderIcon(AppFolder24Filled),
  },
  {
    label: "Game Servers",
    key: "gameServers",
    icon: renderIcon(GameConsole),
  },
  {
    label: "Hub Servers",
    key: "hubServers",
    icon: renderIcon(DeviceHubFilled),
  },
  {
    label: "Rooms",
    key: "rooms",
    icon: renderIcon(BoxMultipleSearch24Regular),
  },
  {
    label: "Room Histories",
    key: "roomHistories",
    icon: renderIcon(History24Regular),
  },
  {
    label: "Settings",
    key: "settings",
    icon: renderIcon(Settings),
    children: [
      {
        label: "Detailed Settings",
        key: "settings",
        icon: renderIcon(SettingsAdjust),
      },
      {
        label: "Theme",
        key: "theme",
        icon: renderIcon(ColorPalette),
        type: "group",
        children: [
          {
            label: "Light",
            key: "light",
            icon: renderIcon(WeatherSunny16Regular),
          },
          {
            label: "Dark",
            key: "dark",
            icon: renderIcon(WeatherMoon16Filled),
          },
        ],
      },
    ],
  },
];
</script>

<template>
  <n-config-provider :theme="theme">
    <n-dialog-provider>
      <n-message-provider>
        <n-layout has-sider position="absolute">
          <n-layout-sider
            bordered
            collapse-mode="width"
            :collapsed-width="64"
            :width="240"
            :collapsed="collapsed"
            show-trigger="bar"
            @collapse="collapsed = true"
            @expand="collapsed = false"
          >
            <n-space justify="center">
              <n-tag :bordered="false" type="success">
                {{ `v${appVersion}` }}
              </n-tag>
            </n-space>
            <n-menu
              :collapsed="collapsed"
              :collapsed-width="64"
              :collapsed-icon-size="22"
              :options="menuOptions"
              v-model:value="activeKey"
              :on-update:value="onMenuUpdate"
            />
          </n-layout-sider>
          <n-layout>
            <router-view />
          </n-layout>
        </n-layout>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>

<style>
#app {
  background-color: black;
}
</style>
