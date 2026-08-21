<script setup lang="ts">
const nav = [
  { to: '/bundles', label: '配置包' },
  { to: '/configurations', label: '版本与影响' },
  { to: '/audits', label: '审计' }
]
</script>

<template>
  <div class="shell">
    <aside>
      <div class="brand"><span class="mark">CR</span><strong>配置发布台</strong></div>
      <nav><RouterLink v-for="item in nav" :key="item.to" :to="item.to">{{ item.label }}</RouterLink></nav>
      <div class="operator"><span>本地管理员</span><small>platform_admin</small></div>
    </aside>
    <main><RouterView /></main>
  </div>
</template>

