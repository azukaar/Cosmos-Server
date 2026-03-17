import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface UserMe {
  Nickname: string;
  Email: string;
  Role: number;
  MFAState: number;
}

export default function createAuthAPI(apiFetch: ApiFetch) {
  function login(values: LoginRequest): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/login', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(values)
    }))
  }

  function sudo(values: { password: string }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/sudo', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(values)
    }))
  }

  function me(): Promise<UserMe> {
    return apiFetch('/cosmos/api/me', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      }
    })
    .then((res) => res.json())
  }

  function logout(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/logout', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      }
    }))
  }

  return {
    login,
    logout,
    me,
    sudo
  };
}
