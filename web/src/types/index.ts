export interface User {
  username: string
}

export interface LoginResponse {
  success: boolean
  username: string
  token: string
}

export interface ModelStat {
  model: string
  request_count: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
}

export interface TimeSeriesPoint {
  time: string
  request_count: number
  total_tokens: number
}

export interface UsageStats {
  total_requests: number
  success_requests: number
  error_requests: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  avg_latency_ms: number
  time_series?: TimeSeriesPoint[]
  model_breakdown?: ModelStat[]
}

export interface AccountInfo {
  project_id?: string
  provider: string
  accounts_ready: number
  accounts_total: number
}

export interface StatsResponse {
  stats: UsageStats
  account: AccountInfo
  range: string
}

export interface RequestRecord {
  id: string
  timestamp: string
  endpoint: string
  model: string
  status_code: number
  latency_ms: number
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  is_stream: boolean
  error_message?: string
  client_ip?: string
}

export interface RequestFilter {
  limit?: number
  offset?: number
  model?: string
  status?: number
  search?: string
}

export interface RequestsResponse {
  requests: RequestRecord[]
  total: number
  limit: number
  offset: number
}

export interface OpenAIModel {
  id: string
  object: string
  created: number
  owned_by: string
}

export interface OpenAIModelsResponse {
  object: string
  data: OpenAIModel[]
}

export interface QuotaBucket {
  bucketId?: string
  displayName?: string
  window?: string
  remainingFraction?: number
  resetTime?: string
}

export interface QuotaGroup {
  displayName: string
  buckets: QuotaBucket[]
}

export interface AccountItem {
  id: string
  projectId: string
  removable: boolean
  coolingUntil?: string
  coolingReason?: string
  quota?: { groups: QuotaGroup[] }
  quotaError?: string
}

export interface AccountTestResult {
  id: string
  projectId: string
  success: boolean
  latencyMs: number
  error?: string
}


