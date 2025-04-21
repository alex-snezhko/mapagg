<script setup lang="ts">
import type { AggregationInputs } from '@/types';

const model = defineModel<AggregationInputs>({ required: true });

</script>

<template>
  <div>
    <div>
      <label for="sampling-rate">Sampling Rate</label>
      <input type="number" name="sampling-rate" v-model="model.samplingRate" />
    </div>
    <div v-for="(tag, i) of model.tags" :id="tag.tag">
      <label :for="tag.tag">{{ tag.tag }}</label>
      <input type="number" :name="tag.tag" v-model="model.tags[i].weight" />
    </div>
  </div>
</template>

<style scoped>
</style>

