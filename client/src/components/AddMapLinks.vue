<script setup lang="ts">
import AddMapLink from './AddMapLink.vue';
</script>

<template>
  <div class="overview">
    Add a map dataset to be incorporated into your aggregation.
  </div>
  <div class="add-map-links">
    <AddMapLink
      img-src="/assets/choropleth_example.png"
      link="/add-map/choropleth"
      name="Add Choropleth Map Dataset"
      description="Add a choropleth map, with distinct geographic regions marked with different colors according to a variable of interest" />

    <AddMapLink
      img-src="/assets/proximity_example.png"
      link="/add-map/points-of-interest"
      name="Add Points of Interest Proximity Map Dataset"
      description="Add a map indicating proximity to points of interest, pulled from a dataset or manually selected" />
  </div>
</template>

<style lang="scss" scoped>
.overview {
  font-size: large;
  padding-top: 20px;
  text-align: center;
}

.add-map-links {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 40px;
}
</style>