<script setup lang="ts">
import { ref } from 'vue'
const emit = defineEmits<{ upload: [file: File] }>()
const active = ref(false)
function choose(files: FileList | null) { if (files?.[0]) emit('upload', files[0]) }
</script>

<template>
  <label class="upload" :class="{ active }" @dragenter.prevent="active=true" @dragleave.prevent="active=false" @dragover.prevent @drop.prevent="active=false; choose($event.dataTransfer?.files ?? null)">
    <input type="file" accept="application/json,.json" @change="choose(($event.target as HTMLInputElement).files)" />
    <span class="upload-icon">&#8682;</span><strong>选择或拖入 JSON 配置包</strong><small>上传后先校验，错误配置不会写入</small>
  </label>
</template>

