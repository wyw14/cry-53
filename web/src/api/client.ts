export type ApiError = { code: string; message: string; fields?: Array<{ path: string; message: string; line?: number }>; request_id: string }

const headers = { 'X-Actor-ID': 'local-admin', 'X-Actor-Roles': 'platform_admin,config_editor,approver,publisher,secret_exporter' }

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const response = await fetch(`/api/v1${path}`, { ...init, headers: { ...headers, ...init.headers } })
  if (!response.ok) throw await response.json() as ApiError
  if (response.status === 204) return undefined as T
  return response.json() as Promise<T>
}

export async function uploadBundle(file: File): Promise<Bundle> {
  const body = new FormData()
  body.append('file', file)
  return api<Bundle>('/bundles', { method: 'POST', body, headers: { 'Idempotency-Key': crypto.randomUUID() } })
}

export type Issue = { code: string; severity: 'error' | 'warning'; path: string; line?: number; message: string; suggestion?: string }
export type Bundle = { id: string; filename: string; state: string; configurations: unknown[]; issues: Issue[] }
export type Change = { configuration: { id: string; name: string; environment: string; type: string }; kind: string; fields: Array<{ field: string; before?: unknown; after?: unknown; sensitive: boolean }> }

export type AuditEvent = {
  id: string
  request_id: string
  actor_id: string
  operation: string
  target: { kind: string; id: string }
  changes?: Array<{ field: string; previous?: unknown; current?: unknown }>
  facts?: Record<string, unknown>
  occurred_at: string
  hash: string
}

export type AuditPage = { items: AuditEvent[]; total: number; page: number; size: number; chain_valid: boolean }

export function listAudits(filters: { operation?: string; actor_id?: string; target?: string } = {}): Promise<AuditPage> {
  const query = new URLSearchParams({ page: '1', size: '50' })
  for (const [key, value] of Object.entries(filters)) if (value) query.set(key, value)
  return api<AuditPage>(`/audits?${query}`)
}
