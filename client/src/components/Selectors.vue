<script setup lang="ts">
import type { AggregateDataTag } from '@/types';

const model = defineModel<AggregateDataTag[]>({ required: true });

</script>

<template>
  <div>
    <div v-for="(tag, i) of model" :id="tag.tag">
      <label :for="tag.tag">{{ tag.tag }}</label>
      <input type="number" :name="tag.tag" v-model="model[i].weight" />
    </div>
  </div>
</template>

<style scoped>
</style>

