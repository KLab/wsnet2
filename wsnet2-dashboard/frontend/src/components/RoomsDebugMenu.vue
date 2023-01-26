<script setup lang="ts">
import { ref } from "vue";
import {
  NButton,
  NCollapse,
  NCollapseItem,
  NIcon,
  NSpace,
  useDialog,
} from "naive-ui";
import { Debug } from "@vicons/carbon";
import apps from "../store/apps";
import rooms from "../store/rooms";

const loading = ref(false);
const appIds = ["testapp1", "testapp2"];
const dialog = useDialog();

async function createMockApps() {
  loading.value = true;

  try {
    for (let index = 0; index < appIds.length; index++) {
      await apps.createApp({
        id: appIds[index],
        name: `${appIds[index]} name`,
        key: "test key",
      });
    }
    dialog.success({
      title: "Create Mock Apps",
      content: `Successfully created apps: ${appIds}`,
    });
  } catch (err) {
    dialog.error({
      title: "Create Mock Apps",
      content: String(err),
    });
  }

  loading.value = false;
}

async function createMockRooms() {
  loading.value = true;

  try {
    const result = await rooms.createMockRooms({ appIds, count: 10 });
    dialog.success({
      title: "Create Mock Rooms",
      content: `Successfully created ${result.length} rooms`,
    });
  } catch (err) {
    dialog.error({
      title: "Create Mock Rooms",
      content: String(err),
    });
  }

  loading.value = false;
}

function test() { }
</script>

<template>
  <n-collapse>
    <template #arrow>
      <n-icon>
        <Debug />
      </n-icon>
    </template>
    <n-collapse-item title="Debug Menu">
      <n-space>
        <n-button ghost type="error" @click="createMockApps">Create Mock Apps</n-button>
        <n-button ghost type="error" :loading="loading" @click="createMockRooms">Create Mock Rooms</n-button>
        <n-button ghost type="error" :loading="loading" @click="test">Test</n-button>
      </n-space>
    </n-collapse-item>
  </n-collapse>
</template>
