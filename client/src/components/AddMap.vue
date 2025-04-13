<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, useTemplateRef, watch, watchEffect } from 'vue';

type Action = "defineBounds" | "selectLegend"
interface LegendItem {
  id: number;
  color: readonly [number, number, number] | null;
  value: number | null;
}

let imageSelected: File;
const imageSelectedSrc = ref<string | null>(null)
const selectingColorId = ref<number | null>(null);
const isDragging = ref(false);
const selectedAction = ref<Action>("defineBounds")
const completedActions = reactive<Record<Action, boolean>>({
  defineBounds: false,
  selectLegend: false
});

const overlayData = reactive({ x: 0, y: 0, scale: 1, imgHeight: 0, imgWidth: 0 });
const legend = ref<LegendItem[]>([]);
const fileTag = ref("");

const refImage = useTemplateRef("ref-img");
const overlayImage = useTemplateRef("overlay-img");
const refImageCanvas = useTemplateRef("ref-canvas");

let mapImgResizeObserver: ResizeObserver | undefined;
let canvasImgResizeObserver: ResizeObserver | undefined;

onMounted(() => {
  window.addEventListener("mousemove", drag);
  window.addEventListener("wheel", scroll)
  window.addEventListener("mouseup", endDrag);
})

onUnmounted(() => {
  window.removeEventListener("mousemove", drag);
  window.removeEventListener("wheel", scroll)
  window.removeEventListener("mouseup", endDrag);

  if (mapImgResizeObserver) {
    mapImgResizeObserver.disconnect();
  }
})

watch(refImage, img => {
  if (img) {
    mapImgResizeObserver = new ResizeObserver(entries => {
      for (const entry of entries) {
        overlayData.imgHeight = entry.contentRect.height;
        overlayData.imgWidth = entry.contentRect.width;
      }
    })

    mapImgResizeObserver.observe(img);
  }
})

watch(refImageCanvas, canvasElem => {
  if (canvasElem) {
    const canvasContext = canvasElem.getContext('2d')!;

    const canvasImg = new Image();
    canvasImg.src = imageSelectedSrc.value!;
    canvasImg.onload = () => {
      canvasContext.drawImage(canvasImg, 0, 0, canvasImg.width, canvasImg.height);
    }
    canvasElem.height = 800;
    canvasElem.width = 800;

    // canvasElem.height = entry.contentRect.height;
    // canvasElem.width = 800

    // canvasElem.height = 800;

    // canvasImgResizeObserver = new ResizeObserver(entries => {
    //   for (const entry of entries) {
    //     canvasElem.width = entry.contentRect.width;
    //     canvasElem.height = entry.contentRect.height;
    //   }
    // })

    // canvasImgResizeObserver.observe(canvasElem);
  }
})

function drag(event: MouseEvent) {
  if (isDragging.value) {
    overlayData.x += event.movementX;
    overlayData.y += event.movementY;
  }
}

function scroll(event: WheelEvent) {
  if (event.shiftKey) {
    const scale = Math.max(0, overlayData.scale + event.deltaY * 0.0002);;
    overlayData.scale = Math.round(scale * 1000) / 1000;
  }
}

function startDrag() {
  isDragging.value = true;
}

function endDrag() {
  isDragging.value = false;
}

function selectImage(event: Event) {
  console.log(event)

  const elem = event.target as HTMLInputElement;
  if (!elem.files) {
    return;
  }

  const file = elem.files[0];

  imageSelected = file;
  imageSelectedSrc.value = URL.createObjectURL(file);
}

let currId = 0;
function addLegendKey() {
  legend.value.push({ id: ++currId, color: null, value: null });
}

function removeLegendKey(id: number) {
  legend.value = legend.value.filter(k => k.id !== id);
  if (selectingColorId.value === id) {
    selectingColorId.value = null;
  }
}

function selectPixel(event: MouseEvent) {
  console.log(event);

  const canvasElem = (event.target as HTMLCanvasElement);

  const imageData = canvasElem.getContext("2d")!.getImageData(event.offsetX, event.offsetY, 1, 1).data;
  const color = [imageData[0], imageData[1], imageData[2]] as const;
  console.log(color);

  legend.value = legend.value.map(k => selectingColorId.value === k.id ? { ...k, color } : k);
  selectingColorId.value = null;
}

function hexColor(color: readonly [number, number, number]): string {
  const componentValue = (component: number) => {
    let str = component.toString(16);
    if (str.length === 1) {
      str = '0' + str;
    }

    return str;
  }

  const a = `#${componentValue(color[0])}${componentValue(color[1])}${componentValue(color[2])}`;
  console.log(a)
  return a;
}

function confirmOverlay() {
  completedActions.defineBounds = true;
  if (!completedActions.selectLegend) {
    selectedAction.value = 'selectLegend';
  } else {
    submitData();
  }
}

function confirmLegend() {
  completedActions.selectLegend = true;
  if (!completedActions.defineBounds) {
    selectedAction.value = 'defineBounds';
  } else {
    submitData();
  }
}

function submitData() {
  // if (!refImage.value || !overlayImage.value) {
  //   return;
  // }

  // const refImageRect = overlayImage.value.getBoundingClientRect();
  // const overlayImageRect = overlayImage.value.getBoundingClientRect();

  // const topLeftDx = overlayImageRect.left - refImageRect.left;
  // const topLeftDy = overlayImageRect.top - refImageRect.top;
  // const bottomRightDx = overlayImageRect.right - refImageRect.right;
  // const bottomRightDy = overlayImageRect.bottom - refImageRect.bottom;

  // const topLeftDx = overlayData.x.left - refImageRect.left;
  // const topLeftDy = overlayImageRect.top - refImageRect.top;
  // const bottomRightDx = overlayImageRect.right - refImageRect.right;
  // const bottomRightDy = overlayImageRect.bottom - refImageRect.bottom;

  if (legend.value.some(l => l.color === null)) {
    alert("Some color values not specified");
    return;
  }

  if (!fileTag.value) {
    alert("Missing file tag");
    return;
  }

  const formData = new FormData();
  formData.append("file", imageSelected);
  formData.append("data", JSON.stringify({
    tag: fileTag.value,
    topLeftDx: overlayData.x,
    topLeftDy: overlayData.y,
    bottomRightDx: overlayData.x + overlayData.imgWidth * overlayData.scale,
    bottomRightDy: overlayData.y + overlayData.imgHeight * overlayData.scale,
    legend: legend.value,
  }));

  fetch("http://localhost:8080/submit-map", {
    method: "POST",
    body: formData,
  });
}

</script>

<template>
  <label for="map">Choose Map</label>
  <input type="file" id="map" name="map" accept="image/png" @input="selectImage" />

  <label for="tag">File Tag</label>
  <input v-model="fileTag" name="tag" />

  <div v-if="!!imageSelectedSrc">
    <button
      @click="selectedAction = 'defineBounds'"
      :class="{ 'active-button': selectedAction === 'defineBounds' }"
    >
      Define Bounds
      <span v-if="completedActions.defineBounds">✓</span>
    </button>

    <button
      @click="selectedAction = 'selectLegend'"
      :class="{ 'active-button': selectedAction === 'selectLegend' }"
    >
      Define Legend
      <span v-if="completedActions.selectLegend">✓</span>
    </button>

    <div v-if="selectedAction === 'defineBounds'">
      <p>Shift + Scroll to adjust overlay size</p>

      <form @submit.prevent="confirmOverlay">
        <label for="scale">X</label>
        <input v-model="overlayData.x" type="number" />

        <label for="scale">Y</label>
        <input v-model="overlayData.y" type="number" />

        <label for="scale">Scale</label>
        <input v-model="overlayData.scale" type="number" name="scale" />

        <input type="submit" value="Submit" />
      </form>

      <div class="map-container">
        <img :src="imageSelectedSrc" ref="ref-img" class="ref-img" />

        <img
          class="overlay-img"
          ref="overlay-img"
          src="http://localhost:8080/assets/blackwhite.png"
          draggable="false"
          :style="{ left: overlayData.x + 'px', top: overlayData.y + 'px', height: overlayData.scale * overlayData.imgHeight + 'px' }"
          @mousedown="startDrag"
        />
      </div>
    </div>
    <div v-else-if="selectedAction === 'selectLegend'">
      <button @click="addLegendKey">Add Legend Key</button>

      <div v-for="legendItem of legend" :key="legendItem.id" :class="{ 'active-legend-item': legendItem.id === selectingColorId }">
        <button @click="selectingColorId = legendItem.id">
          Color
          <div class="color-preview" :style="legendItem.color !== null ? { 'background-color': hexColor(legendItem.color) } : { 'border': '1px solid gray' }">

          </div>
        </button>

        <label for="value">Value</label>
        <input type="number" name="value" v-model="legendItem.value" />

        <button @click="removeLegendKey(legendItem.id)">X</button>
      </div>

      <button @click="confirmLegend">Submit</button>

      <div class="map-container">
        <!-- <img
          :src="imageSelected"
          ref="ref-img"
          class="ref-img"
          :style="{ 'cursor': selectingColorId !== null ? 'crosshair' : 'auto' }"
          @click="selectPixel"
        /> -->

        <canvas
          ref="ref-canvas"
          class="ref-canvas"
          :style="{ 'cursor': selectingColorId !== null ? 'crosshair' : 'auto' }"
          @click="selectPixel"
        ></canvas>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.active-button {
  color: white;
  background-color: gray;
}

.active-legend-item {
  border: 1px solid red;
}

.map-container {
  position: relative;
}

.ref-img {
  height: 800px;
  // display: block;
  margin: auto;
  border: 1px solid black;
}

.ref-canvas {
  height: 800px;
  display: block;
  margin: auto;
  border: 1px solid black;
}

.color-preview {
  display: inline-block;
  width: 10px;
  height: 10px;
}

.overlay-img {
  position: absolute;
  opacity: 0.3;
  border: 1px solid black;
  // pointer-events: none;
}
</style>