<script setup lang="ts">
import type { AggregationInputs } from '@/types';
import Slider from './Slider.vue';
import { ref } from 'vue';

const model = defineModel<AggregationInputs>({ required: true });

</script>

<template>
  <div class="selectors">
    <div class="input-value">
      <label for="sampling-rate">Sampling Rate</label>
      <input type="number" name="sampling-rate" v-model="model.samplingRate" />
    </div>
    <div class="tag-values">
      <h2>Tag Values</h2>
      <div v-for="(tag, i) of model.tags" :key="tag.tag" class="input-value">
        <label>{{ tag.tag }}</label>
        <Slider v-model="model.tags[i].weight" />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.selectors {
  padding: 12px;
  background-color: rgba(100, 100, 100, 0.85);
  width: 280px;
  border-radius: 4px;
}

.tag-values {
  h2 {
    font-size: 1.125em;
    margin-top: 12px;
    font-weight: bold;
    color: #eee;
  }
}

.input-value {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: white;

  input {
    width: 10ch;
  }
}
</style>

