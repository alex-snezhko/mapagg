<script setup lang="ts">
import Map from "./components/Map.vue";
</script>

<template>
  <nav>
    <RouterLink to="/">Home</RouterLink>
    <RouterLink to="/add-map/choropleth">Add Map</RouterLink>
    <RouterLink to="/add-map/points-of-interest">Add Map Points of Interest</RouterLink>
  </nav>

  <main>
    <RouterView />
  </main>
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
