// 模型管理接口封装
const API_BASE = '/v1';  // 直接调用 /v1/models，不走 /api/v1

export interface Model {
  id: string;
  object: string;
  created: number;
  owned_by: string;
}

export interface ModelsResponse {
  object: string;
  data: Model[];
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

  // /v1/models 返回 OpenAI 格式，直接解析
  return response.json();
}

export const modelsAPI = {
  // 获取所有可用模型列表
  list: (): Promise<ModelsResponse> => {
    return request<ModelsResponse>('/models');
  },
};
