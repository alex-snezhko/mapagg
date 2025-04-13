<script setup lang="ts">
import Map from "./Map.vue";
</script>

<template>
  <main>
    <RouterLink to="/add-map">Add Map</RouterLink>

    <Map />
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
