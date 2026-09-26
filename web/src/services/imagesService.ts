import { request } from './apiClient'

export interface ImageData {
  b64_json?: string
  url?: string
  revised_prompt?: string
}

export interface ImageResponse {
  created: number
  data: ImageData[]
}

export interface GenerateImagePayload {
  prompt: string
  size?: string
  style?: string
  model?: string
}

export interface EditImagePayload {
  image: string
  prompt: string
  size?: string
  model?: string
}

export interface EnhancePromptPayload {
  prompt: string
  style?: string
}

export interface EnhancePromptResponse {
  enhanced_prompt: string
}

export const imagesService = {
  async generateImage(payload: GenerateImagePayload): Promise<ImageResponse> {
    return request<ImageResponse>('/api/playground/images/generate', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async editImage(payload: EditImagePayload): Promise<ImageResponse> {
    return request<ImageResponse>('/api/playground/images/edit', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },

  async enhancePrompt(payload: EnhancePromptPayload): Promise<EnhancePromptResponse> {
    return request<EnhancePromptResponse>('/api/playground/images/enhance-prompt', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
}
