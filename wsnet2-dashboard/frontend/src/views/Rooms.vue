<script setup lang="ts">
import { ref, computed, onBeforeMount } from "vue";
import apps from "../store/apps";
import rooms from "../store/rooms";
import type { App } from "../store/apps";
import type { Room } from "../store/rooms";
import overview from "../store/overview";
import { useMessage } from "naive-ui";

// UI components
import {
  NCard,
  NSpace,
  NSelect,
  NDatePicker,
  NGrid,
  NGi,
  NInputNumber,
  NButton,
  NIcon,
  NTooltip,
} from "naive-ui";
import { CachedFilled } from "@vicons/material";

import RoomsDataTableVue from "../components/RoomsDataTable.vue";
import PropsFilter from "../components/PropsFilter.vue";
import SliderRange from "../components/SliderRange.vue";

const message = useMessage();
const appList = ref<App[]>();
const appIdList = computed(() =>
  appList.value?.map((item) => {
    return { label: item.name, value: item.id };
  })
);
const roomList = ref<Room[]>();
const loading = ref(false);

const selectedAppIds = ref<string[] | null>();
const selectedVisible = ref<number | null>();
const selectedJoinable = ref<number | null>();
const selectedWatchable = ref<number | null>();
const number = ref<number | null>();
const hostId = ref<number | null>();
const maxPlayers = ref<number | null>();
const searchGroup = ref<number | null>();
const booleanSelectOptions = computed(() => [
  { label: "true", value: 1 },
  { label: "false", value: 0 },
]);
const players = ref<[number | null, number | null]>([null, null]);
const watchers = ref<[number | null, number | null]>([null, null]);
const playersRange = ref<
  InstanceType<typeof SliderRange> & { reset(): void }
>();
const watchersRange = ref<
  InstanceType<typeof SliderRange> & { reset(): void }
>();

const createdRange = ref(null);
const filters = ref<Array<[string, string]>>([]);

function reset() {
  selectedAppIds.value = null;
  selectedVisible.value = null;
  selectedJoinable.value = null;
  selectedWatchable.value = null;
  number.value = null;
  hostId.value = null;
  maxPlayers.value = null;
  searchGroup.value = null;
  createdRange.value = null;
  playersRange.value?.reset();
  watchersRange.value?.reset();
}

async function apply(useCache: boolean) {
  loading.value = true;
  try {
    // create a copy of veux state to allow operations on retrieved data(e.g. sorting)
    const fetched = await rooms.fetch({
      appId: selectedAppIds.value,
      hostId: hostId.value,
      visible: selectedVisible.value,
      joinable: selectedJoinable.value,
      watchable: selectedWatchable.value,
      number: number.value,
      searchGroup: searchGroup.value,
      maxPlayers: maxPlayers.value,
      playersMin: players.value ? players.value[0] : undefined,
      playersMax: players.value ? players.value[1] : undefined,
      watchersMin: watchers.value ? watchers.value[0] : undefined,
      watchersMax: watchers.value ? watchers.value[1] : undefined,
      createdBefore: createdRange.value
        ? new Date(createdRange.value[1]).toISOString()
        : undefined,
      createdAfter: createdRange.value
        ? new Date(createdRange.value[0]).toISOString()
        : undefined,
      useCache: useCache,
    });

    // apply props filter
    if (filters.value.length > 0) {
      roomList.value = fetched.filter((room) => checkPropsFilter(room));
    } else {
      roomList.value = [...fetched];
    }

    // check if results reaches limit
    var limit = await overview.fetchGraphqlResultLimit();
    if (fetched.length == limit) {
      message.warning(
        `Number of results reaches the limit(${limit}). Please narrow down your search.`
      );
    }
  } catch (err) {
    alert(`Failed to fetch Room list: \n${err}`);
  } finally {
    loading.value = false;
  }
}

function checkPropsFilter(room: Room) {
  // props null check
  if (!room.props) return false;

  // check filters
  for (let index = 0; index < filters.value.length; index++) {
    const filter = filters.value[index];
    // check if data contains error
    if (!!room.props[1]) return false;

    // check if data is object
    const data = room.props[0];
    if (typeof data != "object" || data?.constructor != Object) return false;

    // check if object has specified content
    const keys = filter[0].split("/");
    let next = data;
    keys.forEach((key) => {
      // check if data has key
      if (!next.hasOwnProperty(key)) return false;
      // typescript care
      const keyTyped = key as keyof typeof next;
      next = next[keyTyped];
    });

    switch (typeof next) {
      case "number":
      case "bigint":
        const numberValue = Number(filter[1]);
        // check if filter is NaN
        if (isNaN(numberValue)) return false;
        // check if value is close enough
        if (Math.abs(numberValue - next) > 0.001) return false;
        break;
      case "string":
        // check directly
        if (next !== filter[1]) return false;
        break;
      case "boolean":
        // check boolean
        if (filter[1] !== String(next)) return false;
      default:
        return false;
    }
  }

  // all filter check passed
  return true;
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
  <n-card title="Rooms">
    <n-space vertical>
      <n-select
        v-model:value="selectedAppIds"
        multiple
        :options="appIdList"
        placeholder="Select target AppIds"
        clearable
      />

      <n-grid x-gap="12" cols="1 400:3">
        <n-gi>
          <n-select
            v-model:value="selectedVisible"
            :options="booleanSelectOptions"
            placeholder="Visible"
            clearable
          />
        </n-gi>
        <n-gi>
          <n-select
            v-model:value="selectedJoinable"
            :options="booleanSelectOptions"
            placeholder="Joinable"
            clearable
          />
        </n-gi>
        <n-gi>
          <n-select
            v-model:value="selectedWatchable"
            :options="booleanSelectOptions"
            placeholder="Watchable"
            clearable
          />
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

      <n-grid x-gap="12" cols="1 400:2">
        <n-gi>
          <SliderRange
            place-holder-min="Minimum number of players"
            place-holder-max="Maximun number of players"
            v-model="players"
            ref="playersRange"
          />
        </n-gi>
        <n-gi>
          <SliderRange
            place-holder-min="Minimum number of watchers"
            place-holder-max="Maximun number of watchers"
            v-model="watchers"
            ref="watchersRange"
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

      <n-card>
        <PropsFilter v-model="filters" />
      </n-card>

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

  <RoomsDataTableVue :data="roomList" :loading="loading" />
</template>

<style>
.n-card {
  width: 100%;
}
</style>
