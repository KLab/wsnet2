<script setup lang="ts">
import { reactive, h } from "vue";
import type { ClientInfo } from "../store/roomInfo";
// UI components
import {
  NDataTable,
  PaginationProps,
  DataTableColumn,
  NCard,
  NIcon,
  NButton,
  NGradientText,
  NText,
  NTag,
  NSpace,
} from "naive-ui";
import { useMessage, useDialog } from "naive-ui";
import { RemoveCircleFilled } from "@vicons/material";
import { renderBoolean, render } from "../utils/render";
import roomInfo from "../store/roomInfo";

interface Props {
  appId?: string;
  roomId?: string;
  hostId?: number;
  masterId?: string;
  data?: ClientInfo[];
  loading: boolean;
}

const props = defineProps<Props>();
const message = useMessage();
const dialog = useDialog();
const emit = defineEmits(["update"]);

const columns = reactive<DataTableColumn[]>([
  {
    title: "",
    key: "",
    render(data: unknown) {
      const row = data as ClientInfo;
      return h(
        NButton,
        {
          circle: true,
          secondary: true,
          type: "error",
          onclick: () => {
            dialog.warning({
              title: "Kick Player",
              content: row.id,
              positiveText: "Confirm",
              negativeText: "Cancel",
              onPositiveClick: () => {
                roomInfo
                  .kick({
                    appId: String(props.appId),
                    roomId: String(props.roomId),
                    hostId: props.hostId ? props.hostId : -1,
                    clientId: row.id,
                  })
                  .then(() => {
                    message.success(`Player kicked: ${row.id}`);
                    emit("update");
                  })
                  .catch((err) => {
                    alert(`Kick player failed: \n${err}`);
                  });
              },
            });
          },
        },
        {
          default: () =>
            h(NIcon, { size: 20 }, { default: () => h(RemoveCircleFilled) }),
        }
      );
    },
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "danger",
        },
        { default: () => "Id" }
      );
    },
    key: "id",
    sorter: true,
    render(data: unknown) {
      const row = data as ClientInfo;
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NTag,
              { type: row.id === props.masterId ? "error" : "info" },
              {
                default: () =>
                  row.id === props.masterId ? "MASTER" : "member ",
              }
            ),
            row.id,
          ],
        }
      );
    },
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "warning",
        },
        { default: () => "Is Hub" }
      );
    },
    key: "isHub",
    render(data: unknown) {
      const row = data as ClientInfo;
      return renderBoolean(row.isHub);
    },
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "info",
        },
        { default: () => "Props" }
      );
    },
    key: "props",
    render(data: unknown) {
      const row = data as ClientInfo;
      return h(
        NGradientText,
        { size: 14 },
        { default: () => render(row.props[0]) }
      );
    },
  },
]);

const pagination = reactive<PaginationProps>({
  pageSizes: [5, 10, 25, 50, 100],
  showSizePicker: true,
});

type SortState = {
  columnKey: string | number;
  order: "ascend" | "descend" | false;
};

function onSort(sortState: SortState) {
  switch (sortState.columnKey) {
    case "id": {
      props.data?.sort((a, b) => {
        if (sortState.order === "ascend") {
          return a.id > b.id ? 1 : -1;
        } else {
          return a.id < b.id ? 1 : -1;
        }
      });
      break;
    }
  }
}
</script>

<template>
  <n-card>
    <n-data-table
      size="small"
      ref="table"
      :loading="loading"
      :columns="columns"
      :data="data"
      :row-key="(row: ClientInfo) => row.id"
      :pagination="pagination"
      @update:sorter="onSort"
    />
  </n-card>
</template>
