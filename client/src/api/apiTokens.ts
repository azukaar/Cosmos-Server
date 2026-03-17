import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface CreateAPITokenRequest {
  name: string;
  description?: string;
  readOnly?: boolean;
  ipWhitelist?: string[];
  restrictToConstellation?: boolean;
}

export interface CreateAPITokenResponse {
  token: string;
  name: string;
}

export interface APITokenConfig {
  Name: string;
  Description?: string;
  Owner?: string;
  TokenHash: string;
  Permissions: number[];
  IPWhitelist?: string[];
  RestrictToConstellation: boolean;
  CreatedAt: string;
}

export default function createApiTokensAPI(apiFetch: ApiFetch) {
  function list(): Promise<ApiResponse<Record<string, APITokenConfig>>> {
    return wrap(apiFetch('/cosmos/api/api-tokens', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function create(request: CreateAPITokenRequest): Promise<ApiResponse<CreateAPITokenResponse>> {
    return wrap(apiFetch('/cosmos/api/api-tokens', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(request),
    }))
  }

  function remove(name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/api-tokens', {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ name }),
    }))
  }

  return {
    list,
    create,
    remove,
  };
}
