<script setup lang="ts">
import UploadPanel from '../components/UploadPanel.vue'
import IssueTable from '../components/IssueTable.vue'
import DiffTable from '../components/DiffTable.vue'
import { useBundleStore } from '../stores/bundles'
const store = useBundleStore()
</script>

<template>
  <section class="page-head"><div><p class="eyebrow">配置发布</p><h1>配置包工作台</h1><p>先解析和校验依赖，再审批并原子发布。</p></div><span v-if="store.bundle" class="state">{{ store.bundle.state }}</span></section>
  <UploadPanel @upload="store.upload" />
  <div v-if="store.bundle" class="toolbar"><button @click="store.validate">运行校验</button><button class="secondary" @click="store.preview">查看差异</button><button class="secondary" @click="store.approve">批准</button><button :disabled="store.errors.length > 0" @click="store.publish([])">全部发布</button></div>
  <section v-if="store.bundle" class="band"><header><h2>校验结果</h2><span>{{ store.errors.length }} 个错误 · {{ store.warnings.length }} 个警告</span></header><IssueTable :issues="store.bundle.issues" /></section>
  <section v-if="store.changes.length" class="band"><header><h2>差异预览</h2><span>{{ store.changes.length }} 项配置</span></header><DiffTable :changes="store.changes" /></section>
</template>

