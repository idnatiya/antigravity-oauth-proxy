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
  project_id: string
  provider: string
  token_valid_seconds?: number
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
