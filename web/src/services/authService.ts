import { apiClient } from './apiClient'
import type { User, LoginResponse } from '@/types'

export interface ChangePasswordPayload {
  current_password: string
  new_password: string
  confirm_password: string
}

export const authService = {
  async login(username: string, password: string): Promise<LoginResponse> {
    return apiClient.post<LoginResponse>('/api/auth/login', { username, password })
  },

  async logout(): Promise<{ success: boolean }> {
    return apiClient.post<{ success: boolean }>('/api/auth/logout')
  },

  async getMe(): Promise<User> {
    return apiClient.get<User>('/api/auth/me')
  },

  async changePassword(payload: ChangePasswordPayload): Promise<{ success: boolean; message?: string }> {
    return apiClient.post<{ success: boolean; message?: string }>('/api/auth/change-password', payload)
  },
}
