<script setup lang="ts">
import L from 'leaflet';
import 'leaflet/dist/leaflet.css'
import { onMounted } from 'vue';

onMounted(() => {
  const map = L.map('map').setView([40.741, -73.975], 12);

  const CartoDB_Positron = L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors &copy; <a href="https://carto.com/attributions">CARTO</a>',
    subdomains: 'abcd',
    maxZoom: 20
  });

  CartoDB_Positron.addTo(map);
})
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
