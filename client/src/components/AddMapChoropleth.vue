<script setup lang="ts">
import type { LegendItem, SubmitChoroplethMapRequest } from '@/types';
import { onMounted, onUnmounted, reactive, ref, useTemplateRef, watch } from 'vue';
import { useRouter } from 'vue-router';
import AddMapLayout from './AddMapLayout.vue';
import type { MapPreviewState } from './MapPreview.vue';
import Slider from './Slider.vue';

type EditingAction = "defineBounds" | "selectLegend"
interface LegendItemWithId extends LegendItem {
  id: number;
}

const targetHeight = 640;

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

const overlayData = reactive({ xOffset: 0, yOffset: 0, scale: 1 });
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
    overlayData.xOffset += event.movementX;
    overlayData.yOffset += event.movementY;
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
    overlayLocTopLeftX: Math.floor(overlayData.xOffset * imageScale),
    overlayLocTopLeftY: Math.floor(overlayData.yOffset * imageScale),
    overlayLocBottomRightX: Math.floor((overlayData.xOffset + targetHeight * overlayAspectRatio * overlayData.scale) * imageScale),
    overlayLocBottomRightY: Math.floor((overlayData.yOffset + targetHeight * overlayData.scale) * imageScale),
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
  <AddMapLayout name="Add Choropleth Map Dataset" :map-preview-state="mapPreviewState">
    <div class="form-fields">
      <div class="form-field">
        <label for="tag">Tag</label>
        <input type="text" v-model="tag" name="tag" />
      </div>

      <hr />

      <div class="form-field">
        <label for="map">Choose Map</label>
        <input type="file" id="map" name="map" accept="image/png" @input="selectImage" />
      </div>
    </div>

    <div v-if="!!imageSelectedSrc">
      <button
        @click="selectedEditingAction = 'defineBounds'"
        class="tab-button"
        :class="{ 'active-tab': selectedEditingAction === 'defineBounds' }"
      >
        Define Bounds
        <span v-if="completedActions.defineBounds">✓</span>
      </button>

      <button
        @click="selectedEditingAction = 'selectLegend'"
        class="tab-button"
        :class="{ 'active-tab': selectedEditingAction === 'selectLegend' }"
      >
        Define Legend
        <span v-if="completedActions.selectLegend">✓</span>
      </button>

      <div v-if="selectedEditingAction === 'defineBounds'" class="tab-content">
        <p>Shift + Scroll to adjust overlay size</p>

        <form @submit.prevent="confirmOverlay" class="form-fields">
          <div class="form-field">
            <label for="xOffset">Overlay X Offset</label>
            <input type="number" v-model="overlayData.xOffset" name="xOffset" />
          </div>

          <div class="form-field">
            <label for="yOffset">Overlay Y Offset</label>
            <input type="number" v-model="overlayData.yOffset" name="yOffset" />
          </div>

          <div class="form-field">
            <label for="scale">Overlay Scale</label>
            <input type="number" v-model="overlayData.scale" name="scale" step="any" />
          </div>

          <input type="submit" value="Submit" />
        </form>

        <div class="map-container">
          <img :src="imageSelectedSrc" ref="ref-img" class="ref-img" />

          <img
            class="overlay-img"
            ref="overlay-img"
            src="http://localhost:8080/assets/blackwhite.png"
            draggable="false"
            :style="{ left: overlayData.xOffset + 'px', top: overlayData.yOffset + 'px', height: overlayData.scale * targetHeight + 'px' }"
            @mousedown="startDrag"
          />
        </div>
      </div>
      <div v-else-if="selectedEditingAction === 'selectLegend'" class="tab-content">
        <div class="form-fields">
          <h2>Legend Colors</h2>

          <div
            v-for="legendItem of legend.items"
            :key="legendItem.id"
            class="legend-item"
            :class="{ 'active-legend-item': legendItem.id === selectingColorId }"
          >
            <button class="standard-button select-color-button" @click="selectingColorId = legendItem.id">
              <span class="select-color-text">
                Select Color
              </span>
              <span
                class="color-preview"
                :style="legendItem.color !== null ? { 'background-color': hexColor(legendItem.color) } : { 'border': '1px solid gray' }"
              >
                <span v-if="legendItem.color === null" class="unselected-color-question-mark">
                  ?
                </span>
              </span>
            </button>

            <div>
              <label for="value">Value</label>
              <input type="number" name="value" v-model="legendItem.value" />
              <button class="cancel-button" @click="removeLegendKey(legendItem.id)"><span>🗙</span></button>
            </div>
          </div>

          <button class="standard-button add-legend-key-button" @click="addLegendKey">Add Legend Key</button>

          <hr />

          <div class="form-field">
            <label for="colorTolerance">Color Tolerance</label>
            <input type="number" v-model="legend.colorTolerance" name="colorTolerance" />
          </div>

          <div class="form-field">
            <label for="borderTolerance">Border Tolerance</label>
            <input type="number" v-model="legend.borderTolerance" name="borderTolerance" />
          </div>

          <button @click="confirmLegend">Submit</button>
        </div>

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

.tab-button {
  padding: 8px 20px;
  font-size: 1.0625em;
  font-weight: 500;
  color: #777;
  background-color: #f3f3f3;
  border: 1px solid #bbb;
  border-radius: 6px 6px 0 0;
  margin-bottom: -1px;
}

.tab-button:not(:first-child) {
  margin-left: -1px;
}

.active-tab {
  color: black;
  background-color: white;
  border-bottom-color: white;
}

.tab-content {
  padding: 20px;
  border: 1px solid #bbb;
}

.legend-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.active-legend-item {
  border: 1px solid red;
}

.map-container {
  position: relative;
  overflow: hidden;
}

.ref-img {
  height: 640px;
  margin: auto;
  border: 1px solid black;
}

.ref-canvas {
  height: 640px;
  width: 640px;
  display: block;
  margin: auto;
  border: 1px solid black;
}

.select-color-button span {
  display: inline-block;
  vertical-align: middle;
}

.select-color-text {
  font-size: 1.125em;
}

.color-preview {
  margin-left: 4px;
  width: 16px;
  height: 16px;
  font-size: 0.875em;
  position: relative;

  .unselected-color-question-mark {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }
}

.overlay-img {
  position: absolute;
  opacity: 0.3;
  border: 1px solid black;
}

.cancel-button {
  margin-left: 12px;
  width: 24px;
  color: white;
  background-color: rgb(255, 51, 51);
  border-radius: 50%;
  border: none;
  height: 24px;
  font-size: 1em;
  display: inline-flex;
  align-items: center;
  justify-content: center;

  &:hover {
    background-color: rgb(224, 0, 0);
  }
}
</style>