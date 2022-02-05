<script setup lang="ts">
import Status from "@/components/Status.vue";
import { Snapshot } from "@/types/Snapshot";

interface WebsiteCardProps {
  name: string;
  url: string;
  snapshots: Snapshot[];
}

const props = defineProps<WebsiteCardProps>();
</script>

<template>
  <div class="website-card">
    <div class="website-card__content">
      <div class="website-card__header">
        <span class="website-card__title">
          {{ props.name }}
        </span>
        <a class="website-card__url" :href="props.url">{{ props.url }}</a>
      </div>
    </div>
    <status :snapshots="props.snapshots"></status>
  </div>
</template>

<style>
.website-card {
  padding: 1rem;
}

.website-card__header {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.website-card__title {
  font-family: "IBM Plex Sans", sans-serif;
  font-size: 1.25rem;
  font-weight: 500;
}

.website-card__url {
  font-family: "Libre Franklin", sans-serif;
  font-size: 1rem;
  color: var(--color-gray);
}

.website-card:not(:last-child) {
  border-bottom: 2px solid #eaeaea;
}
</style>
