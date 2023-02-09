<script setup lang="ts">
import { NCollapse, NCollapseItem, NIcon, NSpace, NTag } from "naive-ui";
import { EditFilter } from "@vicons/carbon";
import PairwiseDynamicTags from "./PairwiseDynamicTags.vue";

interface Props {
  modelValue: Array<[string, string]>;
}

const props = defineProps<Props>();
const tooltip = `
   Provide key-value pairs to filter props.
   Convoluted dictionary keys can be written as "key1/key2 : value".
`;
</script>

<template>
  <n-collapse>
    <n-collapse-item title="Props Filter">
      <template #header-extra>
        <n-tag round type="success">{{ props.modelValue.length }}</n-tag>
      </template>
      <template #arrow>
        <n-icon>
          <EditFilter />
        </n-icon>
      </template>
      <n-space>
        <PairwiseDynamicTags v-model="props.modelValue" :tooltip="tooltip" />
      </n-space>
    </n-collapse-item>
  </n-collapse>
</template>
