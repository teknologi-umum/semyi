<script setup lang="ts">
import WebsiteCard from "@/components/WebsiteCard.vue";
import FAKE_WEBSITES from "@/fake/fakeWebsites";
import FAKE_SNAPSHOTS from "@/fake/fakeSnapshots";
</script>

<template>
  <div class="overview">
    <h1 class="overview__title">Overview</h1>
    <div class="overview__websites">
      <website-card
        v-for="website in FAKE_WEBSITES"
        :key="website.name"
        :name="website.name"
        :url="website.url"
        :snapshots="FAKE_SNAPSHOTS[website.name]"
      />
    </div>
  </div>
</template>

<style>
.overview {
  margin: 0 auto;
  padding: 0 1rem;
  max-width: 960px;
}

.overview__title {
  font-size: 2.5rem;
  font-family: "IBM Plex sans", sans-serif;
  font-weight: 800;
  margin: 2rem 0;
}

.overview__websites {
  border-radius: 0.25rem;
  border: 2px solid #eaeaea;
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
  box-shadow: 0 0.25rem 1rem rgba(0, 0, 0, 0.05);
}
</style>
