<script setup lang="ts">
interface Props {
  state: MapPreviewState;
}

export type MapPreviewState = {
  state: "init"
} | {
  state: "loading"
} | {
  state: "present";
  computedMapSrc: string;
  computedImageBlob: Blob;
  tag: string;
}

const { state } = defineProps<Props>();
const emit = defineEmits<{
  success: [];
}>();

async function confirmMap() {
  if (state.state !== 'present') {
    alert("State unexpectedly not present");
    return;
  }

  if (!state.tag) {
    alert("Missing file tag");
    return;
  }

  if (!state.computedImageBlob) {
    alert("Missing computed image file");
    return;
  }

  const formData = new FormData();
  formData.append("file", state.computedImageBlob);
  formData.append("data", JSON.stringify({
    tag: state.tag,
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
  <div class="confirm-map">
    <div class="preview-img-container">
      <div v-if="state.state === 'init'">Preview Will Show Here</div>
      <img v-else-if="state.state === 'loading'" src="/assets/loading_spinner.svg" alt="Loading..." />
      <img v-else :src="state.computedMapSrc" />
    </div>

    <div v-if="state.state === 'present'">
      <p class="confirmation-desc">Does this look accurate?</p>
      <button @click="confirmMap" class="confirm-button">Confirm</button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use "../styles/shared.scss";

.confirm-map {
  h2 {
    font-weight: bold;
    text-align: center;
    padding-bottom: 40px;
  }
}

.preview-img-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 640px;
  width: 640px;
  margin: auto;
  border: 1px solid #bbb;
  border-radius: 4px;

  img {
    max-width: 640px;
    max-height: 640px;
    opacity: 0.75;
  }
}

.confirmation-desc {
  text-align: center;
  margin-top: 20px;
}

.confirm-button {
  background-color: #18a118;
}
</style>
