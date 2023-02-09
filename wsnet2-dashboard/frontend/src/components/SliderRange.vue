<script setup lang="ts">
import { ref } from "vue";
import { NInputNumber, NGrid, NGi } from "naive-ui";

interface Props {
  placeHolderMin: string;
  placeHolderMax: string;
  modelValue: [number | null, number | null];
}

const props = defineProps<Props>();
const emit = defineEmits(["update:modelValue"]);
const value = ref<[number | null, number | null]>([null, null]);

function reset() {
  value.value[0] = null;
  value.value[1] = null;
  emit("update:modelValue", value);
}

defineExpose({ reset });
</script>

<template>
  <n-grid x-gap="12" cols="2">
    <n-gi>
      <n-input-number
        v-model:value="value[0]"
        :placeholder="placeHolderMin"
        clearable
        @update-value="emit('update:modelValue', value)"
      />
    </n-gi>
    <n-gi>
      <n-input-number
        v-model:value="value[1]"
        :placeholder="placeHolderMax"
        clearable
        @update-value="emit('update:modelValue', value)"
      />
    </n-gi>
  </n-grid>
</template>
