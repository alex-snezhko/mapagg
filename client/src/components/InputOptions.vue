<script setup lang="ts">
export interface Option {
  display: string;
  value: string;
}

interface Props {
  options: Option[];
}

const props = defineProps<Props>();

const model = defineModel<string | null>({ required: true })
</script>

<template>
  <div class="input-options">
    <button
      v-for="option of props.options"
      :key="option.value"
      class="btn input-option"
      :class="{ 'active-input-option': model === option.value }"
      @click="model = option.value"
    >
      {{ option.display }}
    </button>
  </div>
</template>

<style lang="scss" scoped>
@import "../styles/shared.scss";

.input-options {
  display: flex;
  justify-content: space-between;
  box-shadow: 0 2px 3px #eee;
}

.input-option {
  width: 50%;
  padding: 10px;
  font-size: 1em;
  font-weight: 600;

  &:first-child {
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    border-right: 0;
    box-shadow: none;
  }

  &:last-child {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
    box-shadow: none;
  }
}

.active-input-option {
  background-color: gray;
}
</style>
