<script setup lang="ts">
const entries = [
  { name: 'prod-metrics-primary', type: 'postgres', env: 'prod', version: 7, status: 'published', usages: 4 },
  { name: 'staging-catalog', type: 'http_api', env: 'staging', version: 3, status: 'published', usages: 2 },
  { name: 'daily-aggregate', type: 'derived', env: 'prod', version: 5, status: 'withdrawn', usages: 1 }
]
</script>

<template>
  <section class="page-head"><div><p class="eyebrow">配置资产</p><h1>版本与影响</h1><p>查看当前版本、替换关系和本地使用方。</p></div></section>
  <section class="band"><div class="filters"><input placeholder="搜索配置名称" /><select><option>全部环境</option><option>prod</option><option>staging</option></select><select><option>全部状态</option><option>published</option><option>withdrawn</option></select></div>
    <div class="table-wrap"><table><thead><tr><th>配置</th><th>类型 / 环境</th><th>版本</th><th>状态</th><th>使用影响</th></tr></thead><tbody><tr v-for="entry in entries" :key="entry.name"><td><strong>{{ entry.name }}</strong></td><td>{{ entry.type }} / {{ entry.env }}</td><td>v{{ entry.version }}</td><td><span class="state">{{ entry.status }}</span></td><td>{{ entry.usages }} 个使用方</td></tr></tbody></table></div>
  </section>
</template>

