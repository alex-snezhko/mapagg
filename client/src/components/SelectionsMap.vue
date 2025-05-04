<script setup lang="ts">
import { initMap } from '@/map';
import type { OverlayBoundsResponse, PointOfInterestWithId } from '@/types';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css'
import { onMounted, onUnmounted } from 'vue';

const model = defineModel<PointOfInterestWithId[]>({ required: true });

interface SaveWeightEventData {
  id: number;
  value: number;
}

const popupsById: Record<number, L.Popup> = {};
const markersById: Record<number, L.Marker<any>> = {};

function savePointOfInterestWeight(event: CustomEvent<SaveWeightEventData>) {
  popupsById[event.detail.id].remove();
  delete popupsById[event.detail.id];
  model.value = model.value.map(poi => poi.id === event.detail.id ? { ...poi, weight: event.detail.value } : poi);
}

function removePointOfInterest(event: CustomEvent<number>) {
  popupsById[event.detail].remove();
  delete popupsById[event.detail];
  markersById[event.detail].remove();
  delete markersById[event.detail];
  model.value = model.value.filter(poi => poi.id !== event.detail);
}

onMounted(async () => {
  const map = await initMap();
  if (!map) {
    return;
  }
  
  const script = document.createElement("script");
  script.textContent = `
  function saveWeight(id) {
    const inputElem = document.getElementById("weight-input-" + id);
    const value = inputElem.valueAsNumber;
    window.dispatchEvent(new CustomEvent("save-weight", { detail: { id, value } }))
  }
  function removeItem(id) {
    window.dispatchEvent(new CustomEvent("remove-point-of-interest", { detail: id }))
  }
  `;
  document.head.appendChild(script);

  window.addEventListener("save-weight", savePointOfInterestWeight as EventListener);
  window.addEventListener("remove-point-of-interest", removePointOfInterest as EventListener);

  map.on("click", e => onMapClick(map, e));
});

onUnmounted(() => {
  window.removeEventListener("save-weight", savePointOfInterestWeight as EventListener);
  window.removeEventListener("remove-point-of-interest", removePointOfInterest as EventListener);
})

let id = 0;
function onMapClick(map: L.Map, event: L.LeafletMouseEvent) {
  const itemId = id++;
  model.value.push({
    id: itemId,
    latLong: { lat: event.latlng.lat, long: event.latlng.lng },
    weight: 1
  });

  const marker = L.marker(event.latlng, { title: "Click to edit" });
  markersById[itemId] = marker;
  marker.addTo(map);
  marker.on("click", () => {
    const popup = L.popup()
      .setLatLng(event.latlng)
      .setContent(`
        <label>Weight: </label><input id="weight-input-${itemId}" type="number" value="${model.value.find(x => x.id === itemId)!.weight}" />

        <button onclick="saveWeight(${itemId})">Save</button>
        <button onclick="removeItem(${itemId})">Remove</button>
      `);
    popupsById[itemId] = popup;
    popup.addTo(map);
  });
}
</script>

<template>
  <div id="map"></div>
</template>

<style scoped>
#map {
  width: 100%;
  height: 100%;
}
</style>
