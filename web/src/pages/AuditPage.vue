<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { listAudits, type ApiError, type AuditEvent } from '../api/client'

const events = ref<AuditEvent[]>([])
const total = ref(0)
const chainValid = ref(true)
const loading = ref(false)
const error = ref('')
const filters = reactive({ operation: '', actor_id: '', target: '' })

const releaseCount = computed(() => events.value.filter(event => event.operation === 'configuration.publish').length)
const sensitiveCount = computed(() => events.value.filter(event => event.operation.includes('export')).length)

async function refresh() {
  loading.value = true
  error.value = ''
  try {
    const page = await listAudits(filters)
    events.value = page.items
    total.value = page.total
    chainValid.value = page.chain_valid
  } catch (reason) {
    error.value = (reason as ApiError).message ?? '审计事件加载失败'
  } finally {
    loading.value = false
  }
}

function operationLabel(operation: string) {
  return ({ 'configuration.publish': '配置发布', 'configuration.rollback': '版本回滚', 'configuration.transition': '状态变更', 'configuration.export': '敏感导出' } as Record<string, string>)[operation] ?? operation
}

function display(value: unknown) {
  if (value === undefined) return '未设置'
  return typeof value === 'string' ? value : JSON.stringify(value)
}

onMounted(refresh)
</script>

<template>
  <section class="page-head">
    <div><p class="eyebrow">发布追溯</p><h1>审计时间线</h1><p>核对配置发布、回滚和敏感操作的连续事件链。</p></div>
    <div class="chain-state" :class="{ broken: !chainValid }"><strong>{{ chainValid ? '链路完整' : '链路异常' }}</strong><span>{{ total }} 条事件</span></div>
  </section>
  <section class="audit-summary" aria-label="审计摘要">
    <div><strong>{{ releaseCount }}</strong><span>本页发布</span></div>
    <div><strong>{{ sensitiveCount }}</strong><span>敏感操作</span></div>
    <div><strong>{{ events.length }}</strong><span>当前载入</span></div>
  </section>
  <section class="band audit-band">
    <form class="filters" @submit.prevent="refresh">
      <select v-model="filters.operation" aria-label="操作类型"><option value="">全部操作</option><option value="configuration.publish">配置发布</option><option value="configuration.rollback">版本回滚</option><option value="configuration.export">敏感导出</option></select>
      <input v-model.trim="filters.actor_id" placeholder="操作者 ID" aria-label="操作者 ID">
      <input v-model.trim="filters.target" placeholder="资源类型:资源 ID" aria-label="目标资源">
      <button type="submit" :disabled="loading">刷新</button>
    </form>
    <p v-if="error" class="audit-error">{{ error }}</p>
    <div v-else-if="events.length" class="audit-timeline">
      <article v-for="event in events" :key="event.id">
        <time :datetime="event.occurred_at">{{ new Date(event.occurred_at).toLocaleString('zh-CN', { hour12: false }) }}</time>
        <div class="timeline-rail"><span></span></div>
        <div class="audit-event">
          <header><div><strong>{{ operationLabel(event.operation) }}</strong><code>{{ event.target.kind }}:{{ event.target.id }}</code></div><span>{{ event.actor_id }}</span></header>
          <dl v-if="event.changes?.length">
            <template v-for="change in event.changes" :key="change.field"><dt>{{ change.field }}</dt><dd><del>{{ display(change.previous) }}</del><span>→</span><ins>{{ display(change.current) }}</ins></dd></template>
          </dl>
          <footer><code>request {{ event.request_id }}</code><code>hash {{ event.hash.slice(0, 12) }}</code></footer>
        </div>
      </article>
    </div>
    <p v-else class="empty">{{ loading ? '正在读取审计链路' : '当前筛选条件下没有事件' }}</p>
  </section>
</template>
