<script setup lang="ts">
import MapPreview, { type MapPreviewState } from './MapPreview.vue';

interface Props {
  mapPreviewState: MapPreviewState;
  name: string;
}

defineProps<Props>();
</script>

<template>
  <div class="layout">
    <div>
      <h2>{{ name }}</h2>
      <slot></slot>
    </div>

    <MapPreview :state="mapPreviewState" />
  </div>
</template>

<style lang="scss" scoped>
.layout {
  display: flex;
  flex-wrap: wrap;
  padding: 40px;
  gap: 120px;
  justify-content: center;
  align-items: center;
}

h2 {
  text-align: center;
  padding-bottom: 40px;
  font-weight: bold;
}
</style>
