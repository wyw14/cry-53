import { defineStore } from 'pinia'
import { api, uploadBundle, type ApiError, type Bundle, type Change } from '../api/client'

export const useBundleStore = defineStore('bundles', {
  state: () => ({ bundle: null as Bundle | null, changes: [] as Change[], loading: false, error: null as ApiError | null }),
  getters: {
    errors: state => state.bundle?.issues.filter(item => item.severity === 'error') ?? [],
    warnings: state => state.bundle?.issues.filter(item => item.severity === 'warning') ?? []
  },
  actions: {
    async upload(file: File) {
      this.loading = true; this.error = null
      try { this.bundle = await uploadBundle(file) } catch (error) { this.error = error as ApiError } finally { this.loading = false }
    },
    async validate() {
      if (!this.bundle) return
      this.bundle = await api<Bundle>(`/bundles/${this.bundle.id}/validate`, { method: 'POST' })
    },
    async preview() {
      if (!this.bundle) return
      const result = await api<{ changes: Change[] }>(`/bundles/${this.bundle.id}/preview`)
      this.changes = result.changes
    },
    async approve() {
      if (!this.bundle) return
      this.bundle = await api<Bundle>(`/bundles/${this.bundle.id}/approve`, { method: 'POST' })
    },
    async publish(ids: string[]) {
      if (!this.bundle) return
      this.bundle = await api<Bundle>(`/bundles/${this.bundle.id}/publish`, { method: 'POST', headers: { 'Content-Type': 'application/json', 'Idempotency-Key': crypto.randomUUID() }, body: JSON.stringify({ selected_ids: ids }) })
    }
  }
})

