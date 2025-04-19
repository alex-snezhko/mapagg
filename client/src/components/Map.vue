<script setup lang="ts">
import type { AggregateDataRequest, AggregateDataTag, MapResponse } from '@/types';
import { debounce } from '@/util';
import L from 'leaflet';
import 'leaflet.heat';
import 'leaflet/dist/leaflet.css'
import { onMounted, watch } from 'vue';

interface Props {
  tags: AggregateDataTag[];
}

const props = defineProps<Props>();

const debouncedWatcher = debounce((map: L.Map, tags: AggregateDataTag[]) => {
  hydrateHeatmap(map, tags)
}, 1000)

onMounted(() => {
  const map = L.map('map').setView([40.741, -73.975], 12);

  const CartoDB_Positron = L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20
  });

  CartoDB_Positron.addTo(map);

  watch(() => props.tags, (newTags, oldTags) => {
    console.log("Here", !!map, newTags, oldTags)
    if (oldTags.length === 0) {
      hydrateHeatmap(map, newTags)
    } else {
      debouncedWatcher(map, newTags);
    }
  }, { deep: true })
});

let layer: any;

async function hydrateHeatmap(map: L.Map, tags: AggregateDataTag[]) {
  const req: AggregateDataRequest = { tags: tags.filter(t => Number.isFinite(t.weight)) };

  const mapRes = await fetch("http://localhost:8080/aggregate-data", {
    method: "POST",
    body: JSON.stringify(req)
  });
  const mapResponse = await mapRes.json() as MapResponse;
  if (!mapResponse.success) {
    alert("Failed to get map " + mapResponse.error);
    return;
  }

  const dx = 0.003250541237113416;
  const dy = 0.0024469734042552854;
  const fs = mapResponse.data.map(([lat, long, val], i) => ({
    type: "Feature",
    id: i.toString(),
    properties: {
      value: val
    },
    geometry: {
      type: "Polygon",
      coordinates: [
        [
          [long - dx, lat - dy],
          [long - dx, lat],
          [long, lat],
          [long, lat - dy],
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
      fillOpacity: 0.5,
      color: color(feature?.properties.value)
    })
  });
  
  layer.addTo(map);
}

function color(val: number) {
  if (val > 1 || val < 0) {
    console.log("OOps", val)
  }
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
main {
  width: 100%;
  height: 100%;
}

#map {
  width: 100%;
  height: 100%;
  /* height: 500px;
  width: 400px; */
}
</style>
