<script setup lang="ts">
import { ref } from "vue";
import {
  NModal,
  NSpace,
  NTag,
  NIcon,
  NDescriptions,
  NDescriptionsItem,
} from "naive-ui";

import { Error as Errormark, Checkmark } from "@vicons/carbon";

import type { Room } from "../store/rooms";
import type { RoomInfo } from "../store/roomInfo";

import roomInfo from "../store/roomInfo";
import ClientsDataTable from "./ClientsDataTable.vue";
import UnknownObject from "./UnknownObject.vue";

const showModal = ref(false);
const room = ref<Room>();
const loading = ref(false);
const details = ref<RoomInfo>();

async function show(data: Room) {
  room.value = data;
  showModal.value = true;
  await updateClientList();
}

async function updateClientList() {
  if (!room?.value) return;
  loading.value = true;
  try {
    details.value = await roomInfo.fetch({
      appId: room.value.app_id,
      roomId: room.value.id,
      hostId: room.value.host_id,
    });
  } catch (err) {
    showModal.value = false;
    alert(`Failed to get RoomInfo: \n${err}`);
  } finally {
    loading.value = false;
  }
}

defineExpose({ show });
</script>

<template>
  <n-modal
    v-model:show="showModal"
    id="detail"
    preset="card"
    title="Room Details"
  >
    <n-space vertical>
      <n-descriptions bordered :column="4">
        <n-descriptions-item label="Id">{{ room?.id }}</n-descriptions-item>
        <n-descriptions-item label="App Id">
          <n-tag type="info">{{ room?.app_id }}</n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="Number">
          <n-tag type="primary">{{ room?.number }}</n-tag>
        </n-descriptions-item>
        <n-descriptions-item label="Created">
          <n-tag type="warning">{{
            new Date(Date.parse(String(room?.created))).toLocaleString("ja-JP")
          }}</n-tag>
        </n-descriptions-item>
      </n-descriptions>
      <n-descriptions bordered :column="8">
        <n-descriptions-item label="Host Id">{{
          room?.host_id
        }}</n-descriptions-item>
        <n-descriptions-item label="Visible">
          <n-icon
            :color="room?.visible ? 'green' : 'red'"
            :component="room?.visible ? Checkmark : Errormark"
          ></n-icon>
        </n-descriptions-item>
        <n-descriptions-item label="Joinable">
          <n-icon
            :color="room?.joinable ? 'green' : 'red'"
            :component="room?.joinable ? Checkmark : Errormark"
          ></n-icon>
        </n-descriptions-item>
        <n-descriptions-item label="Watchable">
          <n-icon
            :color="room?.watchable ? 'green' : 'red'"
            :component="room?.watchable ? Checkmark : Errormark"
          ></n-icon>
        </n-descriptions-item>
        <n-descriptions-item label="Search Group">{{
          details?.roomInfo?.searchGroup
        }}</n-descriptions-item>
        <n-descriptions-item label="Max Players">{{
          details?.roomInfo?.maxPlayers
        }}</n-descriptions-item>
        <n-descriptions-item label="Players">{{
          details?.roomInfo?.players
        }}</n-descriptions-item>
        <n-descriptions-item label="Watchers">{{
          details?.roomInfo?.watchers
        }}</n-descriptions-item>
      </n-descriptions>
      <n-descriptions bordered>
        <n-descriptions-item label="Public Props">
          <UnknownObject :data="details?.roomInfo?.publicProps[0]" />
        </n-descriptions-item>
        <n-descriptions-item label="Private Props">
          <UnknownObject :data="details?.roomInfo?.privateProps[0]" />
        </n-descriptions-item>
      </n-descriptions>
      <ClientsDataTable
        :app-id="room?.app_id"
        :room-id="room?.id"
        :host-id="room?.host_id"
        :master-id="details?.masterId"
        :data="details?.clientInfosList.filter((item) => !!item.id)"
        :loading="loading"
        @update="updateClientList"
      />
    </n-space>

    <template #footer>Footer</template>
  </n-modal>
</template>

<style>
#detail {
  margin: 5%;
}
</style>
