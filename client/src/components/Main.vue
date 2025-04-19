<script setup lang="ts">
import { onMounted, reactive, ref, watch } from "vue";
import Map from "./Map.vue";
import Selectors from "./Selectors.vue";
import type { AggregateDataTag, TagsResponse } from "@/types";

const tags = ref<AggregateDataTag[]>([]);

onMounted(() => {
  getTags();
});

async function getTags() {
  const tagsRes = await fetch("http://localhost:8080/tags");
  const tagsResponse = await tagsRes.json() as TagsResponse;
  if (!tagsResponse.success) {
    alert("Failed to get tags " + tagsResponse.error);
    return;
  }

  tags.value = tagsResponse.data.map(tag => ({ tag, weight: 1 }));
}
</script>

<template>
  <main>
    <RouterLink to="/add-map">Add Map</RouterLink>

    <Selectors v-model="tags" />

    <Map :tags="tags" />
  </main>
</template>

<style scoped>
main {
  width: 100%;
  height: 100%;
}

#map {
  width: 100%;
  height: 100%;
}
</style>
