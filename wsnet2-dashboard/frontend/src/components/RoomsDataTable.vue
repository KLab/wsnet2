<script setup lang="ts">
import { ref, reactive, h } from "vue";
import type { Room } from "../store/rooms";
// UI components
import {
  NDataTable,
  PaginationProps,
  DataTableColumn,
  NCard,
  NTag,
  NIcon,
  NButton,
  NGradientText,
} from "naive-ui";
import { ViewFilled } from "@vicons/carbon";
import RoomDetails from "../components/RoomDetails.vue";
import { renderBoolean, render } from "../utils/render";

interface Props {
  data?: Room[];
  loading?: boolean;
}

const props = defineProps<Props>();
const detailsView = ref<
  InstanceType<typeof RoomDetails> & { show(data: Room): void }
>();

const columns = reactive<DataTableColumn[]>([
  {
    title: "",
    key: "",
    render(data: unknown) {
      const row = data as Room;
      return h(
        NButton,
        {
          circle: true,
          secondary: true,
          type: "primary",
          onclick: () => {
            detailsView.value?.show(row);
          },
        },
        {
          default: () =>
            h(NIcon, { size: 20 }, { default: () => h(ViewFilled) }),
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
      const row = data as Room;
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
          type: "success",
        },
        {
          default: () => "Number",
        }
      );
    },
    key: "number",
    render(data: unknown) {
      const row = data as Room;
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
      const row = data as Room;
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
          type: "warning",
        },
        { default: () => "Visible" }
      );
    },
    key: "visible",
    render(data: unknown) {
      const row = data as Room;
      return renderBoolean(row.visible === 1);
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
        { default: () => "Joinable" }
      );
    },
    key: "joinable",
    render(data: unknown) {
      const row = data as Room;
      return renderBoolean(row.joinable === 1);
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
        { default: () => "Watchable" }
      );
    },
    key: "watchable",
    render(data: unknown) {
      const row = data as Room;
      return renderBoolean(row.watchable === 1);
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
          type: "info",
        },
        { default: () => "Players" }
      );
    },
    key: "players",
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "info",
        },
        { default: () => "Watchers" }
      );
    },
    key: "watchers",
  },
  {
    title() {
      return h(
        NGradientText,
        {
          size: 12,
          type: "danger",
        },
        { default: () => "Props" }
      );
    },
    key: "",
    render(data: unknown) {
      const row = data as Room;
      if (row.props && row.props.length > 0) return render(row.props[0]);
      return h(NTag, { type: "error" }, { default: () => "Null" });
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
      :row-key="(row: Room) => row.id"
      :pagination="pagination"
      @update:sorter="onSort"
    />
    <RoomDetails ref="detailsView" />
  </n-card>
</template>
