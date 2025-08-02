<script setup lang="ts">
import type { LegendItem, SubmitChoroplethMapRequest } from '@/types';
import { onMounted, onUnmounted, reactive, ref, useTemplateRef, watch } from 'vue';
import { useRouter } from 'vue-router';
import AddMapLayout from './AddMapLayout.vue';
import type { MapPreviewState } from './MapPreview.vue';
import CheckMarkIcon from './icons/CheckMarkIcon.vue';
import InfoIcon from './icons/InfoIcon.vue';
import InputOptions from './InputOptions.vue';

type InputMode = "image-import" | "data-import";

const inputMode = ref<InputMode | null>(null);

const inputOptions = [
  { value: "image-import", display: "Import Image" },
  { value: "data-import", display: "Import Data from CSV" },
]

interface DataImportInputs {
	geoJsonNameProperty: string;
	csvNameColumn: string;
	csvValueColumn: string;
	lowerBoundThreshold: number | null;
	upperBoundThreshold: number | null;
	allowNameMatchingLeniency: boolean;
	skipMissing: boolean;
}

const dataImportInputs = reactive<DataImportInputs>({
	geoJsonNameProperty: "",
	csvNameColumn: "Name",
	csvValueColumn: "Value",
	lowerBoundThreshold: null,
	upperBoundThreshold: null,
	allowNameMatchingLeniency: false,
	skipMissing: false,
});

type EditingAction = "defineBounds" | "selectLegend"
interface LegendItemWithId extends LegendItem {
  id: number;
}

const targetHeight = ref(0);

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

let currLegendId = 0;
const legend = reactive<{ items: LegendItemWithId[]; colorTolerance: number; borderTolerance: number; }>({
  items: [
    {
      id: currLegendId++,
      color: null,
      value: null,
    }
  ],
  colorTolerance: 10,
  borderTolerance: 3,
});
const tag = ref("");
const magnificationLevel = ref(2);
const magnificationEnabled = ref(false);
const mouseOverPosition = reactive({ x: 0, y: 0 });

const refImage = useTemplateRef("ref-img");
const refImageCanvas = useTemplateRef("ref-canvas");

const MAGNIFIER_SIZE = 160;

const router = useRouter();

let mapImgResizeObserver: ResizeObserver | undefined;

onMounted(async () => {
  window.addEventListener("mousemove", drag);
  window.addEventListener("wheel", wheel)
  window.addEventListener("mouseup", endDrag);

  const img = new Image();
  img.onload = function () {
    overlayImageBounds = { width: img.naturalWidth, height: img.naturalHeight };
  };
  img.src = "http://localhost:8080/assets/blackwhite.png";
})

onUnmounted(() => {
  window.removeEventListener("mousemove", drag);
  window.removeEventListener("wheel", wheel)
  window.removeEventListener("mouseup", endDrag);

  if (mapImgResizeObserver) {
    mapImgResizeObserver.disconnect();
  }
})

let refImageResizeObserver;
watch(refImage, imgElem => {
  if (imgElem) {
    refImageResizeObserver = new ResizeObserver(entries => {
      for (const entry of entries) {
        if (entry.contentRect.height > 0) {
          targetHeight.value = entry.contentRect.height;
        }
      }
    });
    
    refImageResizeObserver.disconnect();
    refImageResizeObserver.observe(imgElem);
  }
})

watch(refImageCanvas, canvasElem => {
  if (canvasElem) {
    const canvasContext = canvasElem.getContext('2d')!;

    const canvasImg = new Image();
    canvasImg.src = imageSelectedSrc.value!;

    canvasImg.onload = () => {
      setTimeout(() => {
      const targetWidth = Math.floor(selectedImageBounds.width * (targetHeight.value / selectedImageBounds.height));
      canvasElem.width = targetWidth;
      canvasElem.height = targetHeight.value;
      canvasContext.drawImage(canvasImg, 0, 0, targetWidth, targetHeight.value);
      }, 1000);
    }
  }
})

function drag(event: MouseEvent) {
  if (isDragging.value) {
    overlayData.xOffset += event.movementX;
    overlayData.yOffset += event.movementY;
  }
}

function wheel(event: WheelEvent) {
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

let geoJsonFile: File;
let locationValuesFile: File;
async function uploadFile(event: Event, file: "geojson" | "location-values") {
  const elem = event.target as HTMLInputElement;
  if (!elem.files || elem.files.length > 1) {
    alert("Unexpectedly not one file selected");
    return;
  }

  if (file === "geojson") {
    geoJsonFile = elem.files[0];
  } else {
    locationValuesFile = elem.files[0];
  }
}

function addLegendKey() {
  legend.items.push({ id: currLegendId++, color: null, value: null });
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

function onCanvasMouseMove(event: MouseEvent) {
  mouseOverPosition.x = event.layerX;
  mouseOverPosition.y = event.layerY;
}

function onCanvasWheel(event: WheelEvent) {
  if (event.shiftKey) {
    const newLevel = Math.max(1, magnificationLevel.value - (event.deltaY * 0.005));
    magnificationLevel.value = Math.round(newLevel * 1000) / 1000;
  }
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

function toggleSelectingColor(legendItem: LegendItemWithId) {
  if (selectingColorId.value === legendItem.id) {
    selectingColorId.value = null;
  } else {
    selectingColorId.value = legendItem.id;
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

  const imageScale = selectedImageBounds.height / targetHeight.value;

  const overlayAspectRatio = overlayImageBounds.width / overlayImageBounds.height;

  const data: SubmitChoroplethMapRequest = {
    tag: tag.value,
    overlayLocTopLeftX: Math.floor(overlayData.xOffset * imageScale),
    overlayLocTopLeftY: Math.floor(overlayData.yOffset * imageScale),
    overlayLocBottomRightX: Math.floor((overlayData.xOffset + targetHeight.value * overlayAspectRatio * overlayData.scale) * imageScale),
    overlayLocBottomRightY: Math.floor((overlayData.yOffset + targetHeight.value * overlayData.scale) * imageScale),
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
    <div class="form-fields choropleth-form-fields">
      <div class="form-field">
        <label for="tag">Tag</label>
        <input type="text" v-model="tag" name="tag" />
      </div>

      <hr />

      <InputOptions v-model="inputMode" :options="inputOptions" />

      <template v-if="inputMode === 'image-import'">
        <div class="form-field">
          <label for="map">Choose Map</label>
          <input type="file" id="map" name="map" accept="image/png" @input="selectImage" />
        </div>

        <div v-if="!!imageSelectedSrc" class="define-map-info">
          <div class="tab-buttons">
            <button
              @click="selectedEditingAction = 'defineBounds'"
              class="tab-button"
              :class="{ 'active-tab': selectedEditingAction === 'defineBounds' }"
            >
              <span>Define Bounds</span>
              <CheckMarkIcon v-if="completedActions.defineBounds" class="check-mark-icon" />
            </button>

            <button
              @click="selectedEditingAction = 'selectLegend'"
              class="tab-button"
              :class="{ 'active-tab': selectedEditingAction === 'selectLegend' }"
            >
              <span>Define Legend</span>
              <CheckMarkIcon v-if="completedActions.selectLegend" class="check-mark-icon" />
            </button>
          </div>

          <div v-show="selectedEditingAction === 'defineBounds'" class="tab-content">
            <div class="form-fields">
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

              <div class="info-box">
                <InfoIcon class="info-icon" />
                <div>
                  <p>Click and drag overlay to move</p>
                  <p>Shift + Scroll to adjust overlay scale</p>
                </div>
              </div>

              <div class="map-container">
                <div class="map-img-container">
                  <img :src="imageSelectedSrc" ref="ref-img" class="ref-img" />

                  <img
                    class="overlay-img"
                    src="http://localhost:8080/assets/blackwhite.png"
                    draggable="false"
                    :style="{ left: overlayData.xOffset + 'px', top: overlayData.yOffset + 'px', height: overlayData.scale * targetHeight + 'px' }"
                    @mousedown="startDrag"
                  />
                </div>
              </div>

              <button class="btn submit-btn" @click="confirmOverlay">Confirm Overlay Bounds</button>
            </div>
          </div>
          <div v-show="selectedEditingAction === 'selectLegend'" class="tab-content">
            <div class="form-fields">
              <h2>Legend Colors</h2>

              <div
                v-for="legendItem of legend.items"
                :key="legendItem.id"
                class="legend-item"
                :class="{ 'active-legend-item': legendItem.id === selectingColorId }"
              >
                <button class="btn select-color-button" @click="toggleSelectingColor(legendItem)">
                  <span class="select-color-text">
                    {{ legendItem.id === selectingColorId ? 'Selecting Color' : 'Select Color' }}
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

                <div class="select-color-value">
                  <label for="value">Value</label>
                  <input type="number" name="value" v-model="legendItem.value" />
                  <button
                    :disabled="legend.items.length === 1"
                    class="cancel-button"
                    :class="{ 'disabled-btn': legend.items.length === 1 }"
                    @click="removeLegendKey(legendItem.id)"
                  >
                    <span>🗙</span>
                  </button>
                </div>
              </div>

              <button class="btn add-legend-key-button" @click="addLegendKey">Add Legend Key</button>

              <div class="form-field">
                <label for="magnification-enabled">Enable Magnification</label>
                <input type="checkbox" v-model="magnificationEnabled" name="magnification-enabled" />
              </div>

              <div v-if="magnificationEnabled" class="form-field">
                <label for="magnification-level">Magnification Level</label>
                <input type="number" step="any" v-model="magnificationLevel" name="magnification-level" />
              </div>

              <div class="map-container">
                <canvas
                  ref="ref-canvas"
                  class="ref-canvas"
                  :style="{ 'cursor': selectingColorId !== null ? 'crosshair' : 'auto' }"
                  @click="selectPixel"
                  @mousemove="onCanvasMouseMove"
                  @wheel="onCanvasWheel"
                ></canvas>

                <div
                  v-if="selectingColorId !== null && magnificationEnabled"
                  class="magnifying-glass"
                  alt="Magnifying glass"
                  :style="{
                    top: `${mouseOverPosition.y - MAGNIFIER_SIZE / 2}px`,
                    left: `${mouseOverPosition.x - MAGNIFIER_SIZE / 2}px`,
                    width: `${MAGNIFIER_SIZE}px`,
                    height: `${MAGNIFIER_SIZE}px`,
                    backgroundImage: `url(${imageSelectedSrc})`,
                    backgroundSize: `${640 * magnificationLevel}px ${640 * magnificationLevel}px`,
                    backgroundPosition: `-${mouseOverPosition.x * magnificationLevel - MAGNIFIER_SIZE / 2 + 2}px -${mouseOverPosition.y * magnificationLevel - MAGNIFIER_SIZE / 2 + 2}px`
                  }"
                ></div>
              </div>

              <div class="form-field">
                <label for="colorTolerance">Color Tolerance</label>
                <input type="number" v-model="legend.colorTolerance" name="colorTolerance" />
              </div>

              <div class="form-field">
                <label for="borderTolerance">Border Tolerance</label>
                <input type="number" v-model="legend.borderTolerance" name="borderTolerance" />
              </div>

              <button class="btn submit-btn" @click="confirmLegend">Confirm Legend</button>
            </div>
          </div>
        </div>
      </template>
      <template v-else-if="inputMode === 'data-import'">
        <div class="form-field">
          <label for="geojson">GeoJSON Geography Definition</label>
          <input type="file" name="geojson" accept=".json,.geojson" @input="uploadFile($event, 'geojson')" />
        </div>

        <div class="form-field">
          <label for="location-values">Location Values CSV</label>
          <input type="file" name="location-values" accept=".csv" @input="uploadFile($event, 'location-values')" />
        </div>

        <div class="form-field">
          <label for="geojson-name-property">GeoJSON Name Property</label>
          <input type="text" v-model="dataImportInputs.geoJsonNameProperty" name="geojson-name-property" />
        </div>

        <div class="form-field">
          <label for="csv-name-column">Location CSV Name Column</label>
          <input type="text" v-model="dataImportInputs.csvNameColumn" name="csv-name-column" />
        </div>

        <div class="form-field">
          <label for="csv-value-column">GeoJSON Name Property</label>
          <input type="text" v-model="dataImportInputs.csvValueColumn" name="csv-value-column" />
        </div>

        <div class="form-field">
          <label for="lower-bound-threshold">Lower Bound Threshold</label>
          <input type="number" step="any" v-model="dataImportInputs.lowerBoundThreshold" name="lower-bound-threshold" />
        </div>

        <div class="form-field">
          <label for="upper-bound-threshold">Upper Bound Threshold</label>
          <input type="number" step="any" v-model="dataImportInputs.upperBoundThreshold" name="upper-bound-threshold" />
        </div>

        <div class="form-field">
          <label for="allow-name-matching-leniency">Allow Name Leniency</label>
          <input type="checkbox" v-model="dataImportInputs.allowNameMatchingLeniency" name="allow-name-matching-leniency" />
        </div>

        <div class="form-field">
          <label for="skip-missing-values">Skip Missing Values</label>
          <input type="checkbox" v-model="dataImportInputs.skipMissing" name="skip-missing-values" />
        </div>
      </template>
    </div>
  </AddMapLayout>
</template>

<style lang="scss" scoped>
@import "../styles/shared.scss";

.choropleth-form-fields {
  width: 682px;
}

.tab-buttons {
  display: flex;
  align-items: stretch;
}

.tab-button {
  display: flex;
  gap: 6px;
  align-items: center;
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

@keyframes blinking {
  $color: rgb(200, 217, 230);
  $end-color: rgb(135, 179, 212);
  0%, 25% {
    background-color: $color;
  }

  50% {
    background-color: $end-color;
  }

  75%, 100% {
    background-color: $color;
  }
}

.active-legend-item {
  .select-color-button {
    animation: blinking 2s linear infinite;
  }
}

.define-map-info {
  margin-top: 24px;
}

.map-container {
  height: 640px;
  width: 640px;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  border: 1px solid #bbb;
  border-radius: 4px;
}

.map-img-container {
  position: relative;
}

.check-mark-icon {
  width: 1em;
  height: 1em;
}

.ref-img {
  max-height: 640px;
  max-width: 640px;
  margin: auto;
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

.select-color-value label {
  margin-right: 8px;
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
  width: 22px;
  height: 22px;
  color: white;
  background-color: rgb(211, 55, 55);
  border-radius: 50%;
  border: none;
  font-size: 0.875em;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;

  &:not(.disabled-btn):hover {
    background-color: rgb(174, 18, 18);
  }
}

.disabled-btn {
  cursor: auto;
  background-color: rgb(228, 141, 141);
}

.info-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 12px;
  border: 1px solid #aaa;
  border-radius: 8px;
  background-color: #eff3f6;
  font-size: 0.875em;
  // width: fit-content;
  color: #444;
  box-shadow: 0 2px 2px #ddd;
  margin: 4px 0;

  p {
    line-height: 1.6;
  }
}

.magnifying-glass {
  background-repeat: no-repeat;
  border-radius: 8px;
  border: 2px solid black;
  pointer-events: none;
  position: absolute;
}

.info-icon {
  width: 22px;
  height: 22px;
}
</style>