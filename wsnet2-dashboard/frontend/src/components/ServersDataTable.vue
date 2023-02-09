<script setup lang="ts">
import { reactive, h } from "vue";
import type { Server } from "../store/servers";
// UI components
import {
  NDataTable,
  PaginationProps,
  DataTableColumn,
  NCard,
  NTag,
  NGradientText,
} from "naive-ui";

interface Props {
  data?: Server[];
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
        { default: () => "Host Name" }
      );
    },
    key: "hostname",
    render(data: unknown) {
      const row = data as Server;
      return h(
        NTag,
        {
          style: {
            marginRight: "6px",
          },
          type: "info",
        },
        {
          default: () => row.hostname,
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
          default: () => "Public Name",
        }
      );
    },
    key: "public_name",
    render(data: unknown) {
      const row = data as Server;
      return h(NTag, { type: "primary" }, { default: () => row.public_name });
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
        { default: () => "Port (GRPC)" }
      );
    },
    key: "grpc_port",
    sorter: true,
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "warning",
        },
        { default: () => "Port (WS)" }
      );
    },
    key: "ws_port",
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "error",
        },
        {
          default: () => "Status",
        }
      );
    },
    key: "status",
    render(data: unknown) {
      const row = data as Server;
      return h(NTag, { type: "warning" }, { default: () => row.status });
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
        {
          default: () => "Heartbeat",
        }
      );
    },
    key: "heartbeat",
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
      :row-key="(row: Server) => row.id"
      :pagination="pagination"
      @update:sorter="onSort"
    />
  </n-card>
</template>
