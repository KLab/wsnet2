<template>
  <div class="block">
    <n-grid class="block" x-gap="12" cols="1 600:3">
      <n-gi>
        <n-card title="Apps" class="hoverable" @click="router.push('/Apps')">
          <n-statistic :value="data?.servers[0].NApp" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card
          title="Game Servers"
          class="hoverable"
          @click="router.push('/GameServers')"
        >
          <n-statistic :value="data?.servers[0].NGameServer" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="Hub Servers" class="hoverable" @click="router.push('/HubServers')">
          <n-statistic :value="data?.servers[0].NHubServer" />
        </n-card>
      </n-gi>
    </n-grid>
    <n-divider />
    <n-card title="Rooms" class="hoverable" @click="router.push('/Rooms')">
      <n-statistic :value="getTotalRoomsNumber()" />
      <DoughnutChart ref="roomsChart" :chartData="testData" :options="options" />
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onBeforeMount } from "vue";
import { useRouter } from "vue-router";
// UI components
import { NGrid, NGi, NCard, NDivider, NStatistic } from "naive-ui";
import { DoughnutChart, ExtractComponentData } from "vue-chart-3";
import {
  Chart,
  ChartData,
  ChartOptions,
  DoughnutController,
  ArcElement,
  Tooltip,
  Legend,
} from "chart.js";

import type { Overview } from "../store/overview";
import overview from "../store/overview";

const router = useRouter();
Chart.register(DoughnutController, ArcElement, Tooltip, Legend);
const roomsChart = ref<ExtractComponentData<typeof DoughnutChart>>();
const options = ref<ChartOptions<"doughnut">>({
  responsive: true,
  plugins: {
    legend: {
      position: "left",
    },
  },
});

const loading = ref(false);
const data = ref<Overview>();

const testData = computed<ChartData<"doughnut">>(() => ({
  labels: data.value?.rooms.map((item) => item.hostname),
  datasets: [
    {
      data: data.value?.rooms.map((item) => item.num) as number[],
      backgroundColor: data.value?.rooms.map(() => generateRandomHexColor()),
    },
  ],
}));

async function apply() {
  loading.value = true;
  try {
    data.value = await overview.fetch();
  } catch (err) {
    alert(`Failed to fetch overview data: \n${err}`);
  } finally {
    loading.value = false;
  }
}

function getTotalRoomsNumber() {
  let sum = 0;
  data.value?.rooms.forEach((element) => {
    sum += element.num;
  });

  return sum;
}

function generateRandomHexColor() {
  return `#${Math.floor(Math.random() * 16777215).toString(16)}`;
}

onBeforeMount(async () => {
  await apply();
});
</script>

<style scoped>
.block {
  margin: 12px;
  padding: 12px;
}

.hoverable {
  cursor: pointer;
}
</style>
