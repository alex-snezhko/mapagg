<script setup lang="ts">
import { initMap } from '@/map';
import type { AggregationInputs, ComponentData, MapAggregation, MapResponse, OverlayBounds, OverlayBoundsResponse } from '@/types';
import { onlineBinarySearch, debounce } from '@/util';
import L, { latLng } from 'leaflet';
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
  const resp = await initMap();
  if (!resp) {
    return;
  }

  const { map, overlayBounds } = resp;

  map.on("click", e => onMapClick(map, overlayBounds, e));

  watch(() => props.inputs, (newTags, oldTags) => {
    if (oldTags.tags.length === 0) {
      hydrateHeatmap(map, newTags)
    } else {
      debouncedWatcher(map, newTags);
    }
  }, { deep: true })
});

function onMapClick(map: L.Map, overlayBounds: OverlayBounds, event: L.LeafletMouseEvent) {
  const cells = findCell(event.latlng, overlayBounds);
  if (!cells) {
    return;
  }

  const htmlStr = cells.map(cell => `<p><b style="font-weight: bold;">${cell.tag}:</b> ${cell.value}</p>`).join("\n");
  const popup = L.popup()
    .setLatLng(event.latlng)
    .setContent(htmlStr);

  popup.addTo(map);
}

interface TagValue {
  tag: string;
  value: number;
}

function findCell(latlng: L.LatLng, overlayBounds: OverlayBounds): TagValue[] | undefined {
  if (!mapAggregation) {
    return;
  }

  // Each should be same width/height, doesn't matter which component is chosen
  const numPointsY = mapAggregation.componentsData[0].data.length
  const numPointsX = mapAggregation.componentsData[0].data[0].length;

  const topLat = overlayBounds.topLeft.lat;
  const leftLong = overlayBounds.topLeft.long;

  const latI = onlineBinarySearch(numPointsY, latlng.lat, mid => topLat - mid * mapAggregation.gapY, (currLat, tgt) => tgt - currLat);
  const longI = onlineBinarySearch(numPointsX, latlng.lng, mid => leftLong + mid * mapAggregation.gapX, (currLong, tgt) => currLong - tgt);

  return mapAggregation.componentsData.map(c => ({ tag: c.tag, value: c.data[latI][longI] }))
}

let mapAggregation: MapAggregation;

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

  mapAggregation = mapResponse.data;
  const { gapX, gapY, aggregateData } = mapResponse.data;
  const fs = aggregateData.map(([lat, long, val], i) => ({
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

  // layer = L.geoJSON(fc as any, {
  //   style: (feature) => ({
  //     stroke: false,
  //     fillOpacity: 0.4,
  //     color: color(feature?.properties.value)
  //   })
  // });
  
  // layer.addTo(map);

  const res = await fetch("https://nominatim.openstreetmap.org/search?q=brooklyn,+kings+county&addressdetails=1&format=jsonv2&polygon_geojson=1");
  const geojson = (await res.json())[0].geojson;

  layer = L.geoJSON(geojson);
  
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
