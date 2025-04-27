<script setup lang="ts">
import { onMounted, onUnmounted, ref, useTemplateRef } from 'vue';
const model = defineModel<number>({ required: true });
const sliderBar = useTemplateRef('slider-bar');
const dragging = ref(false);

onMounted(() => {
  window.addEventListener('mousemove', listenDrag)
  window.addEventListener('mouseup', mouseUp);
});

onUnmounted(() => {
  window.removeEventListener('mousemove', listenDrag);
  window.removeEventListener('mouseup', mouseUp);
})

const SLIDER_WIDTH = 100;
const SLIDER_HEIGHT = 4;
const SLIDER_RADIUS = 8;

function listenDrag(event: MouseEvent) {
  if (dragging.value) {
    const sliderLeft = sliderBar.value!.getBoundingClientRect().left;

    const unboundVal = (event.clientX - sliderLeft) / SLIDER_WIDTH;
    model.value = Math.min(Math.max(unboundVal, 0), 1);
  }
}

function mouseUp() {
  dragging.value = false;
}
</script>

<template>
  <div
    class="slider-bar"
    ref="slider-bar"
    :style="{ width: `${SLIDER_WIDTH}px`, height: `${SLIDER_HEIGHT}px` }"
  >
    <div class="slider-left" :style="{ width: `${model * SLIDER_WIDTH}px`, height: `${SLIDER_HEIGHT}px` }"></div>
    <div
      class="slider-dot"
      :style="{
        left: `${(model * SLIDER_WIDTH) - SLIDER_RADIUS}px`,
        top: `${-SLIDER_RADIUS + (SLIDER_HEIGHT / 2)}px`,
        cursor: dragging ? 'grabbing' : 'grab'
      }"
      @mousedown="dragging = true"
    >
      <svg :height="SLIDER_RADIUS * 2" :width="SLIDER_RADIUS * 2" view-box="0 0 {{ sliderRadius * 2 }} {{ sliderRadius * 2 }}">
        <circle :cx="SLIDER_RADIUS" :cy="SLIDER_RADIUS" :r="SLIDER_RADIUS" fill="black" />
        <circle :cx="SLIDER_RADIUS" :cy="SLIDER_RADIUS" :r="SLIDER_RADIUS - 2" fill="#ddd" />
      </svg>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.slider-bar {
  background-color: #ccc;
  border-radius: 2px;
  position: relative;
}

.slider-left {
  background-color: #0ebb0e;
  border-radius: 2px;
  position: absolute;
}

.slider-dot {
  position: absolute;
  line-height: 0.4;
}
</style>
