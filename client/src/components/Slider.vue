<script setup lang="ts">
import { onMounted, onUnmounted, ref, useTemplateRef } from 'vue';
interface Props {
  sliderBarWidth: number;
  sliderBarHeight: number;
  sliderRadius: number;
}

const model = defineModel<number>({ required: true });
const { sliderBarWidth, sliderBarHeight, sliderRadius } = defineProps<Props>();
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

function listenDrag(event: MouseEvent) {
  if (dragging.value) {
    const sliderLeft = sliderBar.value!.getBoundingClientRect().left;

    const unboundVal = (event.clientX - sliderLeft) / sliderBarWidth;
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
    :style="{ width: `${sliderBarWidth}px`, height: `${sliderBarHeight}px` }"
  >
    <div class="slider-left" :style="{ width: `${model * sliderBarWidth}px`, height: `${sliderBarHeight}px` }"></div>
    <div
      class="slider-dot"
      :style="{
        left: `${(model * sliderBarWidth) - sliderRadius}px`,
        top: `${-sliderRadius + (sliderBarHeight / 2)}px`,
        cursor: dragging ? 'grabbing' : 'grab'
      }"
      @mousedown="dragging = true"
    >
      <svg :height="sliderRadius * 2" :width="sliderRadius * 2" view-box="0 0 {{ sliderRadius * 2 }} {{ sliderRadius * 2 }}">
        <circle :cx="sliderRadius" :cy="sliderRadius" :r="sliderRadius" fill="#333" />
        <circle :cx="sliderRadius" :cy="sliderRadius" :r="sliderRadius - 1.5" fill="#eee" />
      </svg>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.slider-bar {
  background-color: #ccc;
  border-radius: 4px;
  position: relative;
}

.slider-left {
  background-color: #0ebb0e;
  border-radius: 4px;
  position: absolute;
}

.slider-dot {
  position: absolute;
  line-height: 0.4;
}
</style>
