import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useBundleStore } from './bundles'

describe('bundle store', () => {
  beforeEach(() => setActivePinia(createPinia()))

  it('separates blocking errors from warnings', () => {
    const store = useBundleStore()
    store.bundle = {
      id: 'bundle', filename: 'bundle.json', state: 'validated', configurations: [],
      issues: [
        { code: 'MISSING_REFERENCE', severity: 'error', path: 'configurations[0]', message: 'missing' },
        { code: 'EXPIRING', severity: 'warning', path: 'configurations[1]', message: 'expiring' }
      ]
    }
    expect(store.errors.map(item => item.code)).toEqual(['MISSING_REFERENCE'])
    expect(store.warnings.map(item => item.code)).toEqual(['EXPIRING'])
  })

  it('does not publish when no bundle is selected', async () => {
    const store = useBundleStore()
    const fetch = vi.spyOn(globalThis, 'fetch')
    await store.publish([])
    expect(fetch).not.toHaveBeenCalled()
  })
})

