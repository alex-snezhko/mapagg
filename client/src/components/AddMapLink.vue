<script setup lang="ts">
interface Props {
  imgSrc: string;
  link: string;
  name: string;
  description: string;
}

const props = defineProps<Props>();
</script>

<template>
  <div class="add-map-container">
    <div class="example-image-container">
      <img class="example-image" :src="props.imgSrc" :alt="props.description" />
    </div>
    <RouterLink class="btn add-map-link" :to="props.link">{{ name }}</RouterLink>
    <div class="add-map-description">
      {{ description }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
@import "../styles/shared.scss";

.add-map-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 500px;
  gap: 20px;
  margin: 80px 0;
}

.add-map-link {
  text-decoration: none;
  color: black;
  font-weight: 600;
  padding: 8px 12px;
}

.example-image-container {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 400px;
  height: 300px;
}

.example-image {
  max-width: 400px;
}

.add-map-description {
  font-size: 14px;
  text-align: center;
}
</style>