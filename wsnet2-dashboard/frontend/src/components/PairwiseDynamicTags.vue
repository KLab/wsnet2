<script setup lang="ts">
import { ref } from "vue";
import {
  NIcon,
  NText,
  NSpace,
  NInput,
  NTag,
  NButton,
  NTooltip,
} from "naive-ui";
import { Add } from "@vicons/carbon";
import { FormValidationStatus } from "naive-ui/lib/form/src/interface";

interface Props {
  modelValue: Array<[string, string]>;
  tooltip?: string;
}

const inputVal = ref<[string, string]>();
const showInput = ref(false);
const inputStatus = ref<FormValidationStatus>("success");

const props = defineProps<Props>();
const emit = defineEmits(["update:modelValue"]);

function getType(val: string) {
  if (val == "true" || val == "false") return "error";
  if (isNaN(Number(val))) return "warning";
  return "success";
}

function onClose(index: number) {
  props.modelValue.splice(index, 1);
  emit("update:modelValue", props.modelValue);
}

function onInputBlur() {
  inputVal.value = undefined;
  showInput.value = false;
}

function onInputChange(val: [string, string]) {
  if (!val[0] || !val[1]) {
    inputStatus.value = "error";
    return;
  }

  if (!val[0] && !val[1]) {
    inputStatus.value = "warning";
    showInput.value = false;
    return;
  }

  props.modelValue.push(val);
  inputVal.value = undefined;
  showInput.value = false;
  inputStatus.value = "success";
  emit("update:modelValue", props.modelValue);
}
</script>

<template>
  <n-space>
    <n-tag
      type="warning"
      closable
      v-for="(pair, index) in props.modelValue"
      @close="onClose(index)"
    >
      <n-text strong>{{ pair[0] }}</n-text>
      :
      <n-text :type="getType(pair[1])" depth="3">{{ pair[1] }}</n-text>
    </n-tag>
    <n-tooltip trigger="hover">
      <template #trigger>
        <n-input
          v-if="showInput"
          pair
          size="small"
          separator=":"
          :placeholder="['key', 'value']"
          v-model:value="inputVal"
          clearable
          :status="inputStatus"
          @change="onInputChange"
          @blur="onInputBlur"
        />
        <n-button
          v-else
          dashed
          type="success"
          size="small"
          @click="showInput = true"
        >
          <template #icon>
            <n-icon><Add /></n-icon>
          </template>
        </n-button>
      </template>
      {{ tooltip }}
    </n-tooltip>
  </n-space>
</template>
