import { apiClient } from './apiClient'
import type { APIKey } from '@/types'

export const apiKeysService = {
  getKeys: () => apiClient.get<{ keys: APIKey[] }>('/api/keys'),
  createKey: (data: { name: string; key?: string }) => apiClient.post<APIKey>('/api/keys', data),
  deleteKey: (id: number) => apiClient.delete(`/api/keys?id=${id}`),
}
