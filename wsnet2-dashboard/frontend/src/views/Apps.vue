<script setup lang="ts">
import { ref, onBeforeMount } from "vue";
import app from "../store/apps";
import type { App } from "../store/apps";
import overview from "../store/overview";
import { useMessage } from "naive-ui";

// UI components
import { NCard, NButton, NIcon, NTooltip } from "naive-ui";
import { CachedFilled } from "@vicons/material";

import AppsDataTable from "../components/AppsDataTable.vue";

const message = useMessage();
const list = ref<App[]>();
const loading = ref(false);

async function apply(useCache: boolean) {
  loading.value = true;
  try {
    // create a copy of veux state to allow operations on retrieved data(e.g. sorting)
    list.value = [...(await app.fetch(useCache))];

    // check if results reaches limit
    var limit = await overview.fetchGraphqlResultLimit();
    if (list.value.length == limit) {
      message.warning(
        `Number of results reaches the limit(${limit}). Please narrow down your search.`
      );
    }
  } catch (err) {
    alert(`Failed to fetch game server list: \n${err}`);
  } finally {
    loading.value = false;
  }
}

onBeforeMount(async () => {
  await apply(false);
});
</script>

<template>
  <n-card title="Apps">
    <n-tooltip trigger="hover">
      <template #trigger>
        <n-button
          strong
          secondary
          round
          type="success"
          style="width: 100%"
          @click="apply(false)"
        >
          <template #icon>
            <n-icon><CachedFilled /></n-icon>
          </template>
        </n-button>
      </template>
      Refresh
    </n-tooltip>
  </n-card>

  <AppsDataTable :data="list" :loading="loading" />
</template>

<style>
.n-card {
  width: 100%;
}
</style>
