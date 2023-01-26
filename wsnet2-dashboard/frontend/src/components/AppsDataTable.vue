<script setup lang="ts">
import { reactive, h } from "vue";
import type { App } from "../store/apps";
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
  data?: App[];
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
        { default: () => "Name" }
      );
    },
    key: "name",
    render(data: unknown) {
      const row = data as App;
      return h(
        NTag,
        {
          style: {
            marginRight: "6px",
          },
          type: "info",
        },
        {
          default: () => row.name,
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
          default: () => "Key",
        }
      );
    },
    key: "key",
    render(data: unknown) {
      const row = data as App;
      return h(NTag, { type: "success" }, { default: () => row.key });
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
      :row-key="(row: App) => row.id"
      :pagination="pagination"
      @update:sorter="onSort"
    />
  </n-card>
</template>
