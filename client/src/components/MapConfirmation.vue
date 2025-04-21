<script setup lang="ts">

interface Props {
  computedMapSrc: string | null;
  computedImageBlob: Blob | null;
  tag: string;
}

const { computedMapSrc, computedImageBlob, tag } = defineProps<Props>();
const emit = defineEmits<{
  success: [];
}>();

async function confirmMap() {
  if (!tag) {
    alert("Missing file tag");
    return;
  }

  if (!computedImageBlob) {
    alert("Missing computed image file");
    return;
  }

  const formData = new FormData();
  formData.append("file", computedImageBlob);
  formData.append("data", JSON.stringify({
    tag,
  }));

  const res = await fetch("http://localhost:8080/confirm-map", {
    method: "POST",
    body: formData,
  });

  if (res.ok) {
    emit("success");
  }
}
</script>

<template>
  <div v-if="!!computedMapSrc">
    Does this look accurate?

    <img class="ref-img" :src="computedMapSrc" />

    <button @click="confirmMap">Confirm</button>
  </div>
</template>

<style lang="scss" scoped>
.ref-img {
  height: 800px;
  margin: auto;
  border: 1px solid black;
}
</style>
