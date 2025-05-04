<script setup lang="ts">
import { initMap } from '@/map';
import type { AggregationInputs, MapResponse, OverlayBounds, OverlayBoundsResponse } from '@/types';
import { debounce } from '@/util';
import L from 'leaflet';
import 'leaflet/dist/leaflet.css'
import { onMounted, watch } from 'vue';

interface Props {
  inputs: AggregationInputs;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  loadingChange: [val: boolean];
}>();

const debouncedWatcher = debounce((map: L.Map, inputs: AggregationInputs) => {
  hydrateHeatmap(map, inputs)
}, 1000)

onMounted(async () => {
  const map = await initMap();
  if (!map) {
    return;
  }

  map.on("click", e => onMapClick(map, e));

  watch(() => props.inputs, (newTags, oldTags) => {
    if (oldTags.tags.length === 0) {
      hydrateHeatmap(map, newTags)
    } else {
      debouncedWatcher(map, newTags);
    }
  }, { deep: true })
});

function onMapClick(map: L.Map, event: L.LeafletMouseEvent) {
  const popup = L.popup()
    .setLatLng(event.latlng)
    .setContent(`
      <p><b>TODO</b></p>
    `);

  popup.addTo(map);
}

let layer: any;
async function hydrateHeatmap(map: L.Map, inputs: AggregationInputs) {
  const req: AggregationInputs = {
    ...inputs,
    tags: inputs.tags.filter(t => Number.isFinite(t.weight))
  };

  emit("loadingChange", true);

  const mapRes = await fetch("http://localhost:8080/aggregate-data", {
    method: "POST",
    body: JSON.stringify(req)
  });
  const mapResponse = await mapRes.json() as MapResponse;
  if (!mapResponse.success) {
    alert("Failed to get map " + mapResponse.error);
    return;
  }

  const { gapX, gapY, data } = mapResponse.data;
  const fs = data.map(([lat, long, val], i) => ({
    type: "Feature",
    id: i.toString(),
    properties: {
      value: val
    },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [long - gapX, lat - gapY],
          [long - gapX, lat],
          [long, lat],
          [long, lat - gapY],
        ]
      ]
    }
  }))

  const fc = {
    "type": "FeatureCollection",
    "features": fs
  };

  if (layer) {
    layer.remove();
  }

  layer = L.geoJSON(fc as any, {
    style: (feature) => ({
      stroke: false,
      fillOpacity: 0.4,
      color: color(feature?.properties.value)
    })
  });
  
  layer.addTo(map);

  emit("loadingChange", false);
}

function color(val: number) {
  const value = (1 - val) * 255;
  const green = value
  const red = 255 - value;
  return `#${componentValue(red)}${componentValue(green)}00`;
}

function componentValue(component: number) {
  let str = Math.floor(component).toString(16);
  if (str.length === 1) {
    str = '0' + str;
  }

  return str;
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
