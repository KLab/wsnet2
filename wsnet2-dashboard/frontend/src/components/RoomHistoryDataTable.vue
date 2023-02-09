<script setup lang="ts">
import { reactive, h } from "vue";
import type { RoomHistory } from "../store/roomHistories";
// UI components
import {
  NDataTable,
  PaginationProps,
  DataTableColumn,
  NCard,
  NTag,
  NGradientText,
} from "naive-ui";
import { render } from "../utils/render";

interface Props {
  data?: RoomHistory[];
  loading?: boolean;
}

const props = defineProps<Props>();
const columns = reactive<DataTableColumn[]>([
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
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "info",
        },
        { default: () => "App Id" }
      );
    },
    key: "app_id",
    render(data: unknown) {
      const row = data as RoomHistory;
      return h(
        NTag,
        {
          style: {
            marginRight: "6px",
          },
          type: "info",
        },
        {
          default: () => row.app_id,
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
          type: "info",
        },
        { default: () => "Room Id" }
      );
    },
    key: "room_id",
    sorter: true,
    render(data: unknown) {
      const row = data as RoomHistory;
      return h(
        NTag,
        {
          style: {
            marginRight: "6px",
          },
          type: "info",
        },
        {
          default: () => row.room_id,
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
          type: "success",
        },
        {
          default: () => "Number",
        }
      );
    },
    key: "number",
    render(data: unknown) {
      const row = data as RoomHistory;
      return h(NTag, { type: "primary" }, { default: () => row.number });
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
        { default: () => "Created" }
      );
    },
    key: "created",
    sorter: true,
    render(data: unknown) {
      const row = data as RoomHistory;
      return h(
        NTag,
        {
          style: {
            marginRight: "6px",
          },
          type: "warning",
        },
        {
          default: () =>
            new Date(Date.parse(row.created)).toLocaleString("ja-JP"),
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
          type: "error",
        },
        { default: () => "Closed" }
      );
    },
    key: "closed",
    sorter: true,
    render(data: unknown) {
      const row = data as RoomHistory;
      return h(
        NTag,
        {
          style: {
            marginRight: "6px",
          },
          type: "error",
        },
        {
          default: () =>
            new Date(Date.parse(row.closed)).toLocaleString("ja-JP"),
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
        },
        { default: () => "Host Id" }
      );
    },
    key: "host_id",
    sorter: true,
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "info",
        },
        { default: () => "Search Group" }
      );
    },
    key: "search_group",
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "info",
        },
        { default: () => "Max Players" }
      );
    },
    key: "max_players",
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "danger",
        },
        { default: () => "Public Props" }
      );
    },
    key: "",
    render(data: unknown) {
      const row = data as RoomHistory;
      return render(row.public_props[0]);
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
        { default: () => "Private Props" }
      );
    },
    key: "",
    render(data: unknown) {
      const row = data as RoomHistory;
      return render(row.private_props[0]);
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
        { default: () => "Player Logs" }
      );
    },
    key: "",
    render(data: unknown) {
      const row = data as RoomHistory;
      return render(row.player_logs);
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
    case "room_id": {
      props.data?.sort((a, b) => {
        if (sortState.order === "ascend") {
          return a.room_id > b.room_id ? 1 : -1;
        } else {
          return a.room_id < b.room_id ? 1 : -1;
        }
      });
      break;
    }
    case "created": {
      props.data?.sort((a, b) => {
        if (sortState.order === "ascend") {
          return a.created > b.created ? 1 : -1;
        } else {
          return a.created < b.created ? 1 : -1;
        }
      });
      break;
    }
    case "closed": {
      props.data?.sort((a, b) => {
        if (sortState.order === "ascend") {
          return a.closed > b.closed ? 1 : -1;
        } else {
          return a.closed < b.closed ? 1 : -1;
        }
      });
      break;
    }
    case "host_id": {
      props.data?.sort((a, b) =>
        sortState.order === "ascend"
          ? a.host_id - b.host_id
          : b.host_id - a.host_id
      );
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
      :row-key="(row: RoomHistory) => row.id"
      :pagination="pagination"
      @update:sorter="onSort"
    />
  </n-card>
</template>
