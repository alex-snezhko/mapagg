<script setup lang="ts">
import { onMounted, reactive, ref, watch } from "vue";
import Map from "./Map.vue";
import Selectors from "./Selectors.vue";
import type { AggregateDataTag, AggregationInputs, TagsResponse } from "@/types";

const inputs = reactive<AggregationInputs>({
  samplingRate: 10,
  tags: []
});

const isLoading = ref(false);

onMounted(() => {
  let exitingInputValues: AggregationInputs | null = null;
  const inpsValue = localStorage.getItem("inputs");
  if (inpsValue) {
    exitingInputValues = JSON.parse(inpsValue) as AggregationInputs;
  }

  if (exitingInputValues) {
    inputs.samplingRate = exitingInputValues.samplingRate;
  }

  getTags(exitingInputValues?.tags);
});

watch(inputs, inp => {
  localStorage.setItem("inputs", JSON.stringify(inp));
})

async function getTags(existingTagData: AggregateDataTag[] | undefined) {
  const tagsRes = await fetch("http://localhost:8080/tags");
  const tagsResponse = await tagsRes.json() as TagsResponse;
  if (!tagsResponse.success) {
    alert("Failed to get tags " + tagsResponse.error);
    return;
  }

  inputs.tags = tagsResponse.data.map(tag => ({
    tag,
    weight: existingTagData?.find(td => td.tag === tag)?.weight ?? 1
  }));
}
</script>

<template>
  <main>
    <div v-if="isLoading">Loading...</div>
    <Selectors v-model="inputs" />

    <Map :inputs="inputs" @loading-change="val => isLoading = val" />
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
