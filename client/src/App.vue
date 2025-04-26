<script setup lang="ts">
</script>

<template>
  <nav>
    <RouterLink to="/">Home</RouterLink>
    <RouterLink to="/about">About</RouterLink>
    <RouterLink to="/add-map">Add Map</RouterLink>
  </nav>

  <main>
    <RouterView />
  </main>
</template>

<style lang="scss" scoped>
nav {
  height: 48px;
  background-color: black;
  padding: 10px 20px;
  display: flex;
  gap: 40px;

  a {
    font-weight: bold;
    font-size: 1.125em;
    color: white;
    text-decoration: none;
  }
}

main {
  width: 100%;
  height: calc(100% - 48px);
}
</style>
