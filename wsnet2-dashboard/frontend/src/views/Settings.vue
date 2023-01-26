<script setup lang="ts">
import { ref, onBeforeMount } from "vue";
import settings from "../store/settings";

// UI components
import { NCard, NButton, NIcon, NInput, NGrid, NGi } from "naive-ui";
import { useMessage } from "naive-ui";
import { RefreshOutlined } from "@vicons/material";

const serverAddress = ref<string>();
const message = useMessage();

function applyServerAddress() {
  settings.setServerAddress(String(serverAddress.value));
  message.success("Server address updated.");
}

onBeforeMount(() => {
  if (!settings.serverAddress && import.meta.env.VITE_DEFAULT_SERVER_URI) {
    serverAddress.value = import.meta.env.VITE_DEFAULT_SERVER_URI;
  } else {
    serverAddress.value = settings.serverAddress;
  }
});
</script>

<template>
  <n-card title="Server Settings">
    <n-grid x-gap="12" cols="12">
      <n-gi span="10">
        <n-input
          v-model:value="serverAddress"
          type="text"
          placeholder="The URI of wsnet2-dashboard backend server (e.g. http://192.168.0.1:5555)"
        />
      </n-gi>
      <n-gi span="2">
        <n-button
          strong
          secondary
          style="width: 100%"
          type="success"
          @click="applyServerAddress()"
        >
          <template #icon>
            <n-icon><RefreshOutlined /></n-icon>
          </template>
        </n-button>
      </n-gi>
    </n-grid>
  </n-card>
</template>

<style>
.n-card {
  width: 100%;
}
</style>
