<script setup lang="ts">
import { reactive, ref } from 'vue';
import MapConfirmation from './MapConfirmation.vue';

const inputs = reactive({
  minThresholdRadiusMiles: 0.1,
  maxThresholdRadiusMiles: 0.5,
  latCol: "Latitude",
  longCol: "Longitude",
  tag: ""
})

const fileBlobSrc = ref<string | null>(null);

let csvFile: File;
async function uploadCsv(event: Event) {
  const elem = event.target as HTMLInputElement;
  if (!elem.files || elem.files.length > 1) {
    alert("Unexpectedly not one file selected");
    return;
  }

  csvFile = elem.files[0];
}

let fileBlob: Blob;
async function submit() {
  if (!csvFile) {
    alert("Missing CSV file input");
    return;
  }

  const formData = new FormData();
  formData.append("data", JSON.stringify(inputs));
  formData.append("file", csvFile);

  const res = await fetch("http://localhost:8080/submit-coordinates", {
    method: "POST",
    body: formData,
  })

  if (!res.ok) {
    alert("Failed to get map data");
    return;
  }

  fileBlob = await res.blob();
  fileBlobSrc.value = URL.createObjectURL(fileBlob);
}

</script>

<template>
  <input type="file" name="Choose CSV Sheet" accept=".csv" @change="uploadCsv" />

  <div>
    <label for="tag">Tag</label>
    <input type="text" name="tag" v-model="inputs.tag" />
  </div>

  <div>
    <label for="lat-col">Latitude Column</label>
    <input type="text" name="lat-col" v-model="inputs.latCol" />
  </div>

  <div>
    <label for="long-col">Longitude Column</label>
    <input type="text" name="long-col" v-model="inputs.longCol" />
  </div>

  <div>
    <label for="threshold-radius">Min Threshold Radius (Miles)</label>
    <input type="number" step="any" name="threshold-radius" v-model="inputs.minThresholdRadiusMiles" />
  </div>

  <div>
    <label for="decrease-rate">Max Threshold Radius (Miles)</label>
    <input type="number" step="any" name="decrease-rate" v-model="inputs.maxThresholdRadiusMiles" />
  </div>

  <div>
    <input type="button" :value="fileBlobSrc ? 'Update' : 'Submit'" @click="submit" />
  </div>

  <MapConfirmation :computed-image-blob="fileBlob" :computed-map-src="fileBlobSrc" :tag="inputs.tag" />
</template>
