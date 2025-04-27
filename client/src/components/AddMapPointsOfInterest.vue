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
    <div class="points-of-interest-inputs">
      <h2>Add Points of Interest Proximity Map Dataset</h2>
      <div class="form-fields">
        <div class="form-field">
          <label for="tag">Tag</label>
          <input type="text" name="tag" v-model="inputs.tag" />
        </div>

        <hr />

        <div class="form-field">
          <label for="file">CSV File</label>
          <input type="file" name="file" accept=".csv" @change="uploadCsv" />
        </div>

        <div class="form-field">
          <label for="lat-col">Latitude Column</label>
          <input type="text" name="lat-col" v-model="inputs.latCol" />
        </div>

        <div class="form-field">
          <label for="long-col">Longitude Column</label>
          <input type="text" name="long-col" v-model="inputs.longCol" />
        </div>

        <div class="form-field">
          <label for="threshold-radius">Min Threshold Radius (Miles)</label>
          <input type="number" step="any" name="threshold-radius" class="text-input" v-model="inputs.minThresholdRadiusMiles" />
        </div>

        <div class="form-field">
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
@import "../styles/shared.scss";

.points-of-interest {
  display: flex;
  flex-wrap: wrap;
  padding: 40px;
  gap: 120px;
  justify-content: center;
  align-items: center;
}

.points-of-interest-inputs {
  width: 600px;
}

h2 {
  text-align: center;
  padding-bottom: 40px;
  font-weight: bold;
}

.submit-button {
  @include submit-button;
  background-color: #1775b3;
}
</style>
