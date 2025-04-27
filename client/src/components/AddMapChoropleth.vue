<script setup lang="ts">
import type { LegendItem, SubmitChoroplethMapRequest } from '@/types';
import { onMounted, onUnmounted, reactive, ref, useTemplateRef, watch } from 'vue';
import { useRouter } from 'vue-router';
import AddMapLayout from './AddMapLayout.vue';
import type { MapPreviewState } from './MapPreview.vue';

type EditingAction = "defineBounds" | "selectLegend"
type Action = "editing" | "confirming"
interface LegendItemWithId extends LegendItem {
  id: number;
}

const targetHeight = 800;

let selectedImage: File;
let selectedImageBounds: { width: number; height: number; };
let overlayImageBounds: { width: number; height: number; };
const imageSelectedSrc = ref<string | null>(null)
const selectingColorId = ref<number | null>(null);
const isDragging = ref(false);
const selectedEditingAction = ref<EditingAction>("defineBounds");
const mapPreviewState = ref<MapPreviewState>({ state: "init" });
const completedActions = reactive<Record<EditingAction, boolean>>({
  defineBounds: false,
  selectLegend: false
});

const overlayData = reactive({ x: 0, y: 0, scale: 1 });
const legend = reactive<{ items: LegendItemWithId[]; colorTolerance: number; borderTolerance: number; }>({
  items: [],
  colorTolerance: 10,
  borderTolerance: 3,
});
const tag = ref("");

const refImage = useTemplateRef("ref-img");
const overlayImage = useTemplateRef("overlay-img");
const refImageCanvas = useTemplateRef("ref-canvas");

const router = useRouter();

let mapImgResizeObserver: ResizeObserver | undefined;

onMounted(async () => {
  window.addEventListener("mousemove", drag);
  window.addEventListener("wheel", scroll)
  window.addEventListener("mouseup", endDrag);

  const img = new Image();
  img.onload = function () {
    overlayImageBounds = { width: img.naturalWidth, height: img.naturalHeight };
  };
  img.src = "http://localhost:8080/assets/blackwhite.png";
})

onUnmounted(() => {
  window.removeEventListener("mousemove", drag);
  window.removeEventListener("wheel", scroll)
  window.removeEventListener("mouseup", endDrag);

  if (mapImgResizeObserver) {
    mapImgResizeObserver.disconnect();
  }
})

watch(refImageCanvas, canvasElem => {
  if (canvasElem) {
    const canvasContext = canvasElem.getContext('2d')!;

    const canvasImg = new Image();
    canvasImg.src = imageSelectedSrc.value!;

    canvasImg.onload = () => {
      const targetWidth = Math.floor(selectedImageBounds.width * (targetHeight / selectedImageBounds.height));
      canvasElem.width = targetWidth;
      canvasElem.height = targetHeight;
      canvasContext.drawImage(canvasImg, 0, 0, targetWidth, targetHeight);
    }
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
  const elem = event.target as HTMLInputElement;
  if (!elem.files) {
    return;
  }

  const file = elem.files[0];

  selectedImage = file;
  imageSelectedSrc.value = URL.createObjectURL(file);

  const reader = new FileReader();
  reader.onload = (event) => {
    const img = new Image();
    img.onload = function () {
      selectedImageBounds = { width: img.naturalWidth, height: img.naturalHeight };
    };
    img.src = event.target!.result as string;
  };
  reader.readAsDataURL(file);
}

let currId = 0;
function addLegendKey() {
  legend.items.push({ id: ++currId, color: null, value: null });
}

function removeLegendKey(id: number) {
  legend.items = legend.items.filter(k => k.id !== id);
  if (selectingColorId.value === id) {
    selectingColorId.value = null;
  }
}

function selectPixel(event: MouseEvent) {
  const canvasElem = (event.target as HTMLCanvasElement);

  const canvasBounds = canvasElem.getBoundingClientRect();
  const x = (event.clientX - canvasBounds.left) * (canvasElem.width / canvasElem.clientWidth);
  const y = (event.clientY - canvasBounds.top) * (canvasElem.height / canvasElem.clientHeight);

  const imageData = canvasElem.getContext("2d")!.getImageData(x, y, 1, 1, { colorSpace: 'srgb' }).data;
  const color = [imageData[0], imageData[1], imageData[2], imageData[3]] as const;

  legend.items = legend.items.map(k => selectingColorId.value === k.id ? { ...k, color } : k);
  selectingColorId.value = null;
}

function hexColor(color: readonly [number, number, number, number]): string {
  const componentValue = (component: number) => {
    let str = component.toString(16);
    if (str.length === 1) {
      str = '0' + str;
    }

    return str;
  }

  return `#${componentValue(color[0])}${componentValue(color[1])}${componentValue(color[2])}`;
}

function confirmOverlay() {
  completedActions.defineBounds = true;
  if (!completedActions.selectLegend) {
    selectedEditingAction.value = 'selectLegend';
  } else {
    submitData();
  }
}

function confirmLegend() {
  completedActions.selectLegend = true;
  if (!completedActions.defineBounds) {
    selectedEditingAction.value = 'defineBounds';
  } else {
    submitData();
  }
}

async function submitData() {
  if (legend.items.some(l => l.color === null)) {
    alert("Some color values not specified");
    return;
  }

  if (!tag.value) {
    alert("Missing file tag");
    return;
  }

  const imageScale = selectedImageBounds.height / targetHeight;

  const overlayAspectRatio = overlayImageBounds.width / overlayImageBounds.height;

  const data: SubmitChoroplethMapRequest = {
    tag: tag.value,
    overlayLocTopLeftX: Math.floor(overlayData.x * imageScale),
    overlayLocTopLeftY: Math.floor(overlayData.y * imageScale),
    overlayLocBottomRightX: Math.floor((overlayData.x + targetHeight * overlayAspectRatio * overlayData.scale) * imageScale),
    overlayLocBottomRightY: Math.floor((overlayData.y + targetHeight * overlayData.scale) * imageScale),
    colorTolerance: legend.colorTolerance,
    borderTolerance: legend.borderTolerance,
    legend: legend.items,
  };

  const formData = new FormData();
  formData.append("file", selectedImage);
  formData.append("data", JSON.stringify(data));

  mapPreviewState.value = { state: "loading" };

  const res = await fetch("http://localhost:8080/submit-choropleth-map", {
    method: "POST",
    body: formData,
  });

  const computedImageBlob = await res.blob();
  mapPreviewState.value = {
    state: "present",
    computedImageBlob,
    computedMapSrc: URL.createObjectURL(computedImageBlob),
    tag: tag.value
  };
}
</script>

<template>
  <AddMapLayout name="Choropleth" :map-preview-state="mapPreviewState">
    <div class="form-field">
      <label for="tag">Tag</label>
      <input type="text" v-model="tag" name="tag" />
    </div>

    <div class="form-field">
      <label for="map">Choose Map</label>
      <input type="file" id="map" name="map" accept="image/png" @input="selectImage" />
    </div>

    <div v-if="!!imageSelectedSrc">
      <button
        @click="selectedEditingAction = 'defineBounds'"
        :class="{ 'active-button': selectedEditingAction === 'defineBounds' }"
      >
        Define Bounds
        <span v-if="completedActions.defineBounds">✓</span>
      </button>

      <button
        @click="selectedEditingAction = 'selectLegend'"
        :class="{ 'active-button': selectedEditingAction === 'selectLegend' }"
      >
        Define Legend
        <span v-if="completedActions.selectLegend">✓</span>
      </button>

      <div v-if="selectedEditingAction === 'defineBounds'">
        <p>Shift + Scroll to adjust overlay size</p>

        <form @submit.prevent="confirmOverlay">
          <label for="xOffset">X</label>
          <input type="number" v-model="overlayData.x" name="xOffset" />

          <label for="yOffset">Y</label>
          <input type="number" v-model="overlayData.y" name="yOffset" />

          <label for="scale">Scale</label>
          <input type="number" v-model="overlayData.scale" name="scale" step="any" />

          <input type="submit" value="Submit" />
        </form>

        <div class="map-container">
          <img :src="imageSelectedSrc" ref="ref-img" class="ref-img" />

          <img
            class="overlay-img"
            ref="overlay-img"
            src="http://localhost:8080/assets/blackwhite.png"
            draggable="false"
            :style="{ left: overlayData.x + 'px', top: overlayData.y + 'px', height: overlayData.scale * targetHeight + 'px' }"
            @mousedown="startDrag"
          />
        </div>
      </div>
      <div v-else-if="selectedEditingAction === 'selectLegend'">
        <label for="colorTolerance">Color Tolerance</label>
        <input type="number" v-model="legend.colorTolerance" name="colorTolerance" />

        <label for="borderTolerance">Border Tolerance</label>
        <input type="number" v-model="legend.borderTolerance" name="borderTolerance" />

        <button @click="addLegendKey">Add Legend Key</button>

        <div v-for="legendItem of legend.items" :key="legendItem.id" :class="{ 'active-legend-item': legendItem.id === selectingColorId }">
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
          <canvas
            ref="ref-canvas"
            class="ref-canvas"
            :style="{ 'cursor': selectingColorId !== null ? 'crosshair' : 'auto' }"
            @click="selectPixel"
          ></canvas>
        </div>
      </div>
    </div>
  </AddMapLayout>
</template>

<style lang="scss" scoped>
@import "../styles/shared.scss";

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
  margin: auto;
  border: 1px solid black;
}

.ref-canvas {
  height: 800px;
  width: 800px;
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