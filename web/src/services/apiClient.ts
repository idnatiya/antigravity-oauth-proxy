export class ApiError extends Error {
  status: number
  data?: unknown

  constructor(message: string, status: number, data?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

interface RequestOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined | null>
}

export async function request<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { params, headers, ...customConfig } = options

  let url = endpoint
  if (params) {
    const searchParams = new URLSearchParams()
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined && value !== null && value !== '') {
        searchParams.append(key, String(value))
      }
    }
    const queryString = searchParams.toString()
    if (queryString) {
      url += (url.includes('?') ? '&' : '?') + queryString
    }
  }

  const defaultHeaders: Record<string, string> = {
    'Accept': 'application/json',
  }

  if (options.body && typeof options.body === 'string') {
    defaultHeaders['Content-Type'] = 'application/json'
  }

  const config: RequestInit = {
    method: options.method || 'GET',
    headers: {
      ...defaultHeaders,
      ...headers,
    },
    credentials: 'include', // Include cookie session
    ...customConfig,
  }

  const response = await fetch(url, config)

  if (!response.ok) {
    let errorMsg = `HTTP Error ${response.status}: ${response.statusText}`
    let errorData: unknown

    const text = await response.text()
    try {
      errorData = JSON.parse(text)
      if (typeof errorData === 'object' && errorData !== null && 'error' in errorData) {
        errorMsg = String((errorData as { error: unknown }).error)
      }
    } catch {
      // plain-text error body (e.g. from http.Error)
      if (text.trim()) errorMsg = text.trim()
    }

    throw new ApiError(errorMsg, response.status, errorData)
  }

  if (response.status === 204) {
    return {} as T
  }

  try {
    return (await response.json()) as T
  } catch {
    return {} as T
  }
}

export const apiClient = {
  get: <T>(url: string, params?: RequestOptions['params'], options?: RequestOptions) =>
    request<T>(url, { ...options, method: 'GET', params }),

  post: <T>(url: string, body?: unknown, options?: RequestOptions) =>
    request<T>(url, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    }),

  put: <T>(url: string, body?: unknown, options?: RequestOptions) =>
    request<T>(url, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    }),

  delete: <T>(url: string, options?: RequestOptions) =>
    request<T>(url, { ...options, method: 'DELETE' }),
}
