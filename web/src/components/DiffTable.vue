<script setup lang="ts">
import type { Change } from '../api/client'
defineProps<{ changes: Change[] }>()
</script>

<template>
  <div class="changes"><article v-for="change in changes" :key="change.configuration.id">
    <header><div><strong>{{ change.configuration.name }}</strong><small>{{ change.configuration.environment }} · {{ change.configuration.type }}</small></div><span class="change-kind">{{ change.kind }}</span></header>
    <div v-for="field in change.fields" :key="field.field" class="field-diff"><code>{{ field.field }}</code><del>{{ field.before ?? '未设置' }}</del><ins>{{ field.after ?? '已删除' }}</ins></div>
  </article></div>
</template>

