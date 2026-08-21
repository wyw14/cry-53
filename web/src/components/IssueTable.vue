<script setup lang="ts">
import { computed } from 'vue'
import type { Issue } from '../api/client'
const props = defineProps<{ issues: Issue[] }>()
const grouped = computed(() => {
  const buckets = new Map<string, Issue[]>()
  for (const issue of props.issues) {
    const scope = issue.path.split('.values.')[0]
    buckets.set(scope, [...(buckets.get(scope) ?? []), issue])
  }
  return [...buckets.entries()].map(([scope, issues]) => ({ scope, issues }))
})
</script>

<template>
  <div v-if="grouped.length" class="issue-queue">
    <section v-for="group in grouped" :key="group.scope" class="issue-group">
      <header><code>{{ group.scope }}</code><span>{{ group.issues.length }} 项</span></header>
      <article v-for="issue in group.issues" :key="`${issue.path}-${issue.code}`">
        <span class="severity" :class="issue.severity">{{ issue.severity === 'error' ? '错误' : '警告' }}</span>
        <div><strong>{{ issue.message }}</strong><code>{{ issue.path }}<template v-if="issue.line"> · 第 {{ issue.line }} 行</template></code></div>
        <p>{{ issue.suggestion || '请根据字段规范修正后重新校验' }}</p>
      </article>
    </section>
  </div>
  <p v-else class="empty">暂无校验问题</p>
</template>
