import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authService, type ChangePasswordPayload } from '@/services/authService'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const initialized = ref(false)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!user.value)

  async function initAuth(): Promise<boolean> {
    if (initialized.value) return isAuthenticated.value
    loading.value = true
    error.value = null
    try {
      user.value = await authService.getMe()
      return true
    } catch {
      user.value = null
      return false
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  async function login(username: string, pass: string): Promise<boolean> {
    loading.value = true
    error.value = null
    try {
      const res = await authService.login(username, pass)
      user.value = { username: res.username }
      return true
    } catch (err: unknown) {
      if (err instanceof Error) {
        error.value = err.message
      } else {
        error.value = 'Failed to login'
      }
      throw err
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    loading.value = true
    try {
      await authService.logout()
    } catch {
      // ignore
    } finally {
      user.value = null
      loading.value = false
    }
  }

  async function changePassword(payload: ChangePasswordPayload): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await authService.changePassword(payload)
    } finally {
      loading.value = false
    }
  }

  return {
    user,
    initialized,
    loading,
    error,
    isAuthenticated,
    initAuth,
    login,
    logout,
    changePassword,
  }
})
