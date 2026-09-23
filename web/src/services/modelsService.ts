import { apiClient } from './apiClient'
import type { OpenAIModelsResponse, OpenAIModel } from '@/types'

export const modelsService = {
  async getModels(): Promise<OpenAIModel[]> {
    const res = await apiClient.get<OpenAIModelsResponse>('/v1/models')
    return res.data || []
  },
}
