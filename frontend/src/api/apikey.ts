// API Key 管理接口封装
const API_BASE = '/api/v1';

export interface APIKey {
  id: number;
  name: string;
  token_prefix: string;
  enabled: boolean;
  description: string;
  created_at: string;
  last_used_at?: string;
}

export interface CreateAPIKeyRequest {
  name: string;
  description?: string;
}

export interface CreateAPIKeyResponse {
  id: number;
  name: string;
  token: string;
  token_prefix: string;
  enabled: boolean;
  description: string;
  created_at: string;
}

export interface UpdateAPIKeyRequest {
  name?: string;
  description?: string;
  enabled?: boolean;
}

interface ApiResponse<T> {
  code: number;
  data: T;
  msg: string;
}

export interface FullTokenResponse {
  token: string;
}

async function request<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE}${url}`, {
    headers: {
      'Content-Type': 'application/json',
    },
    ...options,
  });

  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }

  const result: ApiResponse<T> = await response.json();
  if (result.code !== 200) {
    throw new Error(result.msg);
  }

  return result.data;
}

export const apiKeyAPI = {
  // 获取所有 API Keys
  list: (): Promise<APIKey[]> => {
    return request<APIKey[]>('/apikeys');
  },

  // 创建 API Key
  create: (data: CreateAPIKeyRequest): Promise<CreateAPIKeyResponse> => {
    return request<CreateAPIKeyResponse>('/apikeys', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取完整 Token
  getFullToken: (id: number): Promise<FullTokenResponse> => {
    return request<FullTokenResponse>(`/apikeys/${id}/token`);
  },

  // 更新 API Key
  update: (id: number, data: UpdateAPIKeyRequest): Promise<APIKey> => {
    return request<APIKey>(`/apikeys/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  },

  // 删除 API Key
  delete: (id: number): Promise<void> => {
    return request<void>(`/apikeys/${id}`, {
      method: 'DELETE',
    });
  },
};
