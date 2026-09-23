import { apiClient } from './apiClient'
import type { StatsResponse, RequestsResponse, RequestFilter } from '@/types'

export const usageService = {
  async getStats(range = '24h'): Promise<StatsResponse> {
    return apiClient.get<StatsResponse>('/api/usage/stats', { range })
  },

  async getRequests(filter: RequestFilter = {}): Promise<RequestsResponse> {
    const params: Record<string, string | number | undefined> = {
      limit: filter.limit ?? 50,
      offset: filter.offset ?? 0,
      model: filter.model || undefined,
      status: filter.status !== undefined && filter.status !== 0 ? filter.status : undefined,
      search: filter.search || undefined,
    }
    return apiClient.get<RequestsResponse>('/api/usage/requests', params)
  },
}
