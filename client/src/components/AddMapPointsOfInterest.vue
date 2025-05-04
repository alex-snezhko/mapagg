<script setup lang="ts">
import { reactive, ref } from 'vue';
import MapPreview, { type MapPreviewState } from './MapPreview.vue';
import SelectionsMap from './SelectionsMap.vue';
import type { PointOfInterestWithId, SubmitPointsOfInterestData, SubmitPointsOfInterestFromCsvData } from '@/types';

type InputMode = "csv-import" | "select-map";

const inputs = reactive({
  tag: "",
  minThresholdRadiusMiles: 0.1,
  maxThresholdRadiusMiles: 0.5,
});

const csvInputs = reactive({
  latCol: "Latitude",
  longCol: "Longitude",
  weightCol: null,
});

const mapInputs = ref<PointOfInterestWithId[]>([]);

const mapPreviewState = ref<MapPreviewState>({ state: "init" });

const inputMode = ref<InputMode | null>(null);

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
  let res: Response;
  switch (inputMode.value) {
    case "csv-import":
      if (!csvFile) {
        alert("Missing CSV file input");
        return;
      }

      const formData = new FormData();
      formData.append("data", JSON.stringify({ ...inputs, ...csvInputs } satisfies SubmitPointsOfInterestFromCsvData));
      formData.append("file", csvFile);

      mapPreviewState.value = { state: "loading" };

      res = await fetch("http://localhost:8080/submit-coordinates-from-csv", {
        method: "POST",
        body: formData,
      });
      break;
    case "select-map":
      if (mapInputs.value.length === 0) {
        alert("No lat longs specified");
        return;
      }

      const data: SubmitPointsOfInterestData = { ...inputs, pointsOfInterest: mapInputs.value };

      mapPreviewState.value = { state: "loading" };

      res = await fetch("http://localhost:8080/submit-coordinates", {
        method: "POST",
        body: JSON.stringify(data),
      });
      break;
    default:
      alert("Unexpectedly encountered submit without mode selected");
      return;
  }

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

        <div class="input-options">
          <button
            class="btn input-option"
            :class="{ 'active-input-option': inputMode === 'csv-import' }"
            @click="inputMode = 'csv-import'"
          >
            Import CSV File
          </button>
          <button
            class="btn input-option"
            :class="{ 'active-input-option': inputMode === 'select-map' }"
            @click="inputMode = 'select-map'"
          >
            Select on Map
          </button>
        </div>

        <template v-if="inputMode === 'csv-import'">
          <div class="form-field">
            <label for="file">CSV File</label>
            <input type="file" name="file" accept=".csv" @change="uploadCsv" />
          </div>

          <div class="form-field">
            <label for="lat-col">Latitude Column</label>
            <input type="text" name="lat-col" v-model="csvInputs.latCol" />
          </div>

          <div class="form-field">
            <label for="long-col">Longitude Column</label>
            <input type="text" name="long-col" v-model="csvInputs.longCol" />
          </div>

          <div class="form-field">
            <label for="weight-col">Weight Column (Optional)</label>
            <input type="text" name="weight-col" v-model="csvInputs.weightCol" />
          </div>
        </template>
        <template v-else-if="inputMode === 'select-map'">
          <div class="selections-map">
            <SelectionsMap v-model="mapInputs" />
          </div>
        </template>

        <template v-if="!!inputMode">
          <div class="form-field">
            <label for="threshold-radius">Base Min Threshold Radius (Miles)</label>
            <input type="number" step="any" name="threshold-radius" class="text-input" v-model="inputs.minThresholdRadiusMiles" />
          </div>

          <div class="form-field">
            <label for="decrease-rate">Base Max Threshold Radius (Miles)</label>
            <input type="number" step="any" name="decrease-rate" class="text-input" v-model="inputs.maxThresholdRadiusMiles" />
          </div>
        </template>

        <div>
          <input
            type="button"
            class="btn submit-btn"
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

.input-options {
  display: flex;
  justify-content: space-between;
  box-shadow: 0 2px 3px #eee;
}

.input-option {
  width: 50%;
  padding: 10px;
  font-size: 1em;
  font-weight: 600;

  &:first-child {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    border-right: 0;
    box-shadow: none;
  }

  &:last-child {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
    box-shadow: none;
  }
}

.active-input-option {
  background-color: gray;
}

.selections-map {
  height: 640px;
  width: 640px;
  border: 1px solid #bbb;
  border-radius: 4px;
}

.points-of-interest {
  display: flex;
  flex-wrap: wrap;
  padding: 40px;
  gap: 120px;
  justify-content: center;
  align-items: center;
}

.points-of-interest-inputs {
  width: 640px;
}

h2 {
  text-align: center;
  padding-bottom: 40px;
  font-weight: bold;
}
</style>
