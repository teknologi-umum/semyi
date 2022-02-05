<script setup lang="ts">
import { Snapshot } from "@/types/Snapshot";
import { computed, onMounted, ref } from "vue";

const CONTAINER_HEIGHT = 30;
const BAR_AMOUNT = 100;
const GAP = 4;

const container = ref<HTMLDivElement | null>(null);
let containerWidth = ref(0);

const barWidth = computed(() => {
  const value = (containerWidth.value - GAP * (BAR_AMOUNT - 1)) / BAR_AMOUNT;
  return value < 1 ? 0 : value;
});
const barRadius = computed(() => {
  const value = barWidth.value / 2;
  return value < 1 ? 0 : value;
});

onMounted(() => {
  if (container?.value !== null) {
    containerWidth.value = container.value.clientWidth;
  }
});

function showTooltip(e) {}

interface StatusProps {
  snapshots: Snapshot[];
}
const props = defineProps<StatusProps>();
</script>

<template>
  <div class="status" ref="container">
    <svg
      :width="containerWidth"
      :height="30"
      xmlns="http://www.w3.org/2000/svg"
      version="1.1"
      :viewBox="`0 0 ${containerWidth} ${CONTAINER_HEIGHT}`"
    >
      <rect
        class="status__bar"
        v-for="i in Array(BAR_AMOUNT)
          .fill(0)
          .map((_, i) => i)
          .reverse()"
        :key="i"
        :width="barWidth"
        :height="CONTAINER_HEIGHT"
        :x="i * (barWidth + GAP)"
        :y="0"
        :fill="
          props.snapshots?.[i]?.statusCode !== undefined &&
          props.snapshots?.[i]?.statusCode !== null
            ? props.snapshots[i].statusCode === 200
              ? 'var(--color-emerald)'
              : 'var(--color-red)'
            : 'var(--color-light-gray)'
        "
        fill-opacity="1"
        :rx="barRadius"
        :ry="barRadius"
      />
    </svg>
  </div>
</template>

<style>
.status {
  margin-top: 1rem;
  width: 100%;
  overflow: hidden;
}
</style>
