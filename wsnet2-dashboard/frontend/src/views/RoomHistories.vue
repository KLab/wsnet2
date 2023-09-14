<script setup lang="ts">
import { ref, computed, onBeforeMount } from "vue";
import apps from "../store/apps";
import roomHistories from "../store/roomHistories";
import type { App } from "../store/apps";
import type { RoomHistory } from "../store/roomHistories";
import overview from "../store/overview";
import RoomHistoryDataTable from "../components/RoomHistoryDataTable.vue";
import { useMessage } from "naive-ui";

// UI components
import {
  NCard,
  NSpace,
  NSelect,
  NDatePicker,
  NGrid,
  NGi,
  NInput,
  NInputNumber,
  NButton,
  NIcon,
  NTooltip,
} from "naive-ui";
import { CachedFilled } from "@vicons/material";

const message = useMessage();
const appList = ref<App[]>();
const appIdList = computed(() =>
  appList.value?.map((item) => {
    return { label: item.name, value: item.id };
  })
);
const list = ref<RoomHistory[]>();
const loading = ref(false);

const selectedAppIds = ref<string[] | null>();
const roomId = ref<string | null>();
const hostId = ref<number | null>();
const number = ref<number | null>();
const maxPlayers = ref<number | null>();
const searchGroup = ref<number | null>();

const createdRange = ref(null);
const closedRange = ref(null);

function reset() {
  selectedAppIds.value = null;
  roomId.value = null;
  hostId.value = null;
  number.value = null;
  maxPlayers.value = null;
  searchGroup.value = null;
  createdRange.value = null;
  closedRange.value = null;
}

// check if all parameters are empty except selectedAppIds
function checkEmptyParameters() {
  return (
    !roomId.value &&
    !hostId.value &&
    !number.value &&
    !maxPlayers.value &&
    !searchGroup.value &&
    !createdRange.value &&
    !closedRange.value
  );
}

async function apply(useCache: boolean) {
  loading.value = true;

  // check if all parameters are empty and send a warning
  if (checkEmptyParameters()) {
    message.warning(
      "No search parameters set. Please fill in at least one parameter."
    );
    loading.value = false;
    return;
  }

  try {
    // create a copy of veux state to allow operations on retrieved data(e.g. sorting)
    const fetched = await roomHistories.fetch({
      appId: selectedAppIds.value,
      roomId: roomId.value,
      hostId: hostId.value,
      number: number.value,
      searchGroup: searchGroup.value,
      maxPlayers: maxPlayers.value,
      createdBefore: createdRange.value
        ? new Date(createdRange.value[1]).toISOString()
        : undefined,
      createdAfter: createdRange.value
        ? new Date(createdRange.value[0]).toISOString()
        : undefined,
      closedBefore: closedRange.value
        ? new Date(closedRange.value[1]).toISOString()
        : undefined,
      closedAfter: closedRange.value
        ? new Date(closedRange.value[0]).toISOString()
        : undefined,
      useCache: useCache,
    });

    list.value = [...fetched];

    // send a warning if number of results reaches the limit
    var limit = await overview.fetchGraphqlResultLimit();
    if (fetched.length == limit) {
      message.warning(
        `Number of results reaches the GraphQL query limit of ${limit}, please narrow down your search.`
      );
    }
  } catch (err) {
    alert(`Failed to fetch RoomHistory  list: \n${err}`);
  } finally {
    loading.value = false;
  }
}

onBeforeMount(async () => {
  try {
    appList.value = await apps.fetch(false);
  } catch (err) {
    alert(`Failed to fetch AppId list: \n${err}`);
  }
});
</script>

<template>
  <n-card title="Room Histories">
    <n-space vertical>
      <n-select
        v-model:value="selectedAppIds"
        multiple
        :options="appIdList"
        placeholder="Select target AppIds"
        clearable
      />

      <n-grid x-gap="12" cols="1">
        <n-gi>
          <n-input v-model:value="roomId" placeholder="Room Id" clearable />
        </n-gi>
      </n-grid>

      <n-grid x-gap="12" cols="1 400:4">
        <n-gi>
          <n-input-number
            v-model:value="hostId"
            placeholder="Host Id"
            clearable
          />
        </n-gi>
        <n-gi>
          <n-input-number
            v-model:value="number"
            placeholder="Number"
            clearable
          />
        </n-gi>
        <n-gi>
          <n-input-number
            v-model:value="maxPlayers"
            placeholder="Max players"
            clearable
          />
        </n-gi>
        <n-gi>
          <n-input-number
            v-model:value="searchGroup"
            placeholder="Search group"
            clearable
          />
        </n-gi>
      </n-grid>

      <n-date-picker
        v-model:value="createdRange"
        type="datetimerange"
        end-placeholder="Created before..."
        start-placeholder="Created after..."
        clearable
        input-readonly
      />

      <n-date-picker
        v-model:value="closedRange"
        type="datetimerange"
        end-placeholder="Closed before..."
        start-placeholder="Closed after..."
        clearable
        input-readonly
      />

      <n-grid x-gap="12" cols="12">
        <n-gi span="6">
          <n-button
            strong
            secondary
            round
            type="warning"
            style="width: 100%"
            @click="reset"
            >Reset</n-button
          >
        </n-gi>
        <n-gi span="4">
          <n-tooltip trigger="hover">
            <template #trigger>
              <n-button
                strong
                secondary
                round
                type="info"
                style="width: 100%"
                @click="apply(true)"
              >
                Apply
              </n-button>
            </template>
            Apply filters on cached results
          </n-tooltip>
        </n-gi>
        <n-gi span="2">
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
            Refetch data from server without using cache
          </n-tooltip>
        </n-gi>
      </n-grid>
    </n-space>
  </n-card>

  <RoomHistoryDataTable :data="list" :loading="loading" />
</template>

<style>
.n-card {
  width: 100%;
}
</style>
