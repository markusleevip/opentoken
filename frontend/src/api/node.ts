// 节点管理接口封装
const API_BASE = '/api/v1';

export interface NodeCredential {
  ID: number;
  Name: string;
  TokenHash: string;
  TokenPrefix: string;
  Enabled: boolean;
  CreatedAt: string;
  UpdatedAt: string;
  // Go 后端实际返回小写字段
  id?: number;
  name?: string;
  token_hash?: string;
  token_prefix?: string;
  enabled?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface CreateNodeCredentialRequest {
  name: string;
}

export interface CreateNodeCredentialResponse {
  id: number;
  name: string;
  token: string;
  token_prefix: string;
  enabled: boolean;
}

export interface FullTokenResponse {
  token: string;
}

export interface OnlineNode {
  node_id: number;
  node_name: string;
  models: string[];
  connected_at: string;
}

interface ApiResponse<T> {
  code: number;
  data: T;
  msg: string;
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

export const nodeAPI = {
  // 获取所有节点凭证
  listCredentials: (): Promise<NodeCredential[]> => {
    return request<NodeCredential[]>('/node/credentials');
  },

  // 创建节点凭证（生成 Token）
  createCredential: (data: CreateNodeCredentialRequest): Promise<CreateNodeCredentialResponse> => {
    return request<CreateNodeCredentialResponse>('/node/credentials', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  },

  // 获取完整 Token
  getFullToken: (id: number): Promise<FullTokenResponse> => {
    return request<FullTokenResponse>(`/node/credentials/${id}/token`);
  },

  // 删除节点凭证
  deleteCredential: (id: number): Promise<void> => {
    return request<void>(`/node/credentials/${id}`, {
      method: 'DELETE',
    });
  },

  // 获取在线节点列表
  listOnlineNodes: (): Promise<OnlineNode[]> => {
    return request<OnlineNode[]>('/node/online');
  },
};
