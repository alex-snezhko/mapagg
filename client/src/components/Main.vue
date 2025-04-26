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
  <div class="map-container">
    <img src="/assets/loading_spinner.svg" v-if="isLoading" class="loading-spinner" />
    <Selectors v-model="inputs" class="selectors-box" />

    <Map :inputs="inputs" class="aggregate-map" @loading-change="val => isLoading = val" />
  </div>
</template>

<style lang="scss" scoped>
.map-container {
  width: 100%;
  height: 100%;

  position: relative;
}

.loading-spinner {
  z-index: 10;
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.selectors-box {
  position: absolute;
  z-index: 10;
  top: 8px;
  left: 8px;
}

.aggregate-map {
  z-index: 0;
}
</style>
