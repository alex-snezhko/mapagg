<script setup lang="ts">
import { reactive, ref } from 'vue';
import MapPreview, { type MapPreviewState } from './MapPreview.vue';

const inputs = reactive({
  minThresholdRadiusMiles: 0.1,
  maxThresholdRadiusMiles: 0.5,
  latCol: "Latitude",
  longCol: "Longitude",
  tag: ""
})

const mapPreviewState = ref<MapPreviewState>({ state: "init" });

let csvFile: File;
async function uploadCsv(event: Event) {
  const elem = event.target as HTMLInputElement;
  if (!elem.files || elem.files.length > 1) {
    alert("Unexpectedly not one file selected");
    return;
  }

  csvFile = elem.files[0];
}

async function submit() {
  if (!csvFile) {
    alert("Missing CSV file input");
    return;
  }

  const formData = new FormData();
  formData.append("data", JSON.stringify(inputs));
  formData.append("file", csvFile);

  mapPreviewState.value = { state: "loading" };

  const res = await fetch("http://localhost:8080/submit-coordinates", {
    method: "POST",
    body: formData,
  })

  if (!res.ok) {
    mapPreviewState.value = { state: "init" };
    alert("Failed to get map data");
    return;
  }

  const computedImageBlob = await res.blob();
  mapPreviewState.value = {
    state: "present",
    computedImageBlob,
    computedMapSrc: URL.createObjectURL(computedImageBlob),
    tag: inputs.tag
  };
}

</script>

<template>
  <div class="points-of-interest">
    <div>
      <h2>Add Points of Interest Proximity Map Dataset</h2>
      <div class="poi-inputs">
        <div class="poi-input">
          <label for="tag">Tag</label>
          <input type="text" name="tag" class="text-input" v-model="inputs.tag" />
        </div>

        <hr />

        <div class="poi-input">
          <label for="file">CSV File</label>
          <input type="file" name="file" class="file-input" accept=".csv" @change="uploadCsv" />
        </div>

        <div class="poi-input">
          <label for="lat-col">Latitude Column</label>
          <input type="text" name="lat-col" class="text-input" v-model="inputs.latCol" />
        </div>

        <div class="poi-input">
          <label for="long-col">Longitude Column</label>
          <input type="text" name="long-col" class="text-input" v-model="inputs.longCol" />
        </div>

        <div class="poi-input">
          <label for="threshold-radius">Min Threshold Radius (Miles)</label>
          <input type="number" step="any" name="threshold-radius" class="text-input" v-model="inputs.minThresholdRadiusMiles" />
        </div>

        <div class="poi-input">
          <label for="decrease-rate">Max Threshold Radius (Miles)</label>
          <input type="number" step="any" name="decrease-rate" class="text-input" v-model="inputs.maxThresholdRadiusMiles" />
        </div>

        <div>
          <input
            type="button"
            class="submit-button"
            :value="mapPreviewState.state === 'present' ? 'Update' : 'Submit'"
            @click="submit"
          />
        </div>
      </div>
    </div>

    <MapPreview :state="mapPreviewState" />
  </div>
</template>

<style lang="scss" scoped>
@use "../styles/shared.scss";

.points-of-interest {
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

.poi-inputs {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin: auto;
  width: 600px;

  hr {
    margin: 10px 0;
    border: none;
    border-top: 1px solid #bbb;
  }
}

.poi-input {
  font-size: 1.125em;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.submit-button {
  @include shared.submit-button;
  background-color: #1775b3;
}

.text-input {
  font-size: 0.875em;
  width: 250px;
  padding: 6px 10px;
  border: 1px solid #bbb;
  background-color: #fafafa;
  border-radius: 4px;
  box-shadow: 0px 2px 3px #eee;
}

.file-input {
  width: 250px;
  &::file-selector-button {
    padding: 2px 6px;
    font-size: 1.125em;
    background-color: #f0f0f0;
    border: 1px solid #bbb;
    border-radius: 4px;
    padding: 4px 8px;
    font-family: 'Inter', -apple-system, sans-serif;
    box-shadow: 0px 2px 3px #eee;

    &:hover {
      background-color: #e7e7e7;
    }
  }
}
</style>
