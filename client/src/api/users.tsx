import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface User {
  Nickname: string;
  Email: string;
  Role: number;
  RegisterKey: string;
  RegisterKeyExp: string;
  RegisteredAt: string;
  LastPasswordChangedAt: string;
  CreatedAt: string;
  LastLogin: string;
  MFAState: number;
  Link?: string;
}

export default function createUsersAPI(apiFetch: ApiFetch) {
  function list(): Promise<ApiResponse<User[]>> {
    return wrap(apiFetch('/cosmos/api/users', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function create(values: Partial<User> & { password?: string }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/users', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(values),
    }));
  }

  function register(values: { nickname: string; password: string; registerKey?: string }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/register', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function invite(values: { nickname: string; email?: string }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/invite', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function edit(nickname: string, values: Partial<User>): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/users/'+nickname, {
      method: 'PATCH',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function get(nickname: string): Promise<ApiResponse<User>> {
    return wrap(apiFetch('/cosmos/api/users/'+nickname, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function deleteUser(nickname: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/users/'+nickname, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function new2FA(nickname: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/mfa', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function check2FA(values: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/mfa', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        token: values
      }),
    }))
  }

  function reset2FA(values: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/mfa', {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        nickname: values
      }),
    }))
  }

  function resetPassword(values: { nickname: string; password?: string }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/password-reset', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function getNotifs(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/notifications', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function readNotifs(notifs: string[]): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/notifications/read?ids=' + notifs.join(','), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  return {
    list,
    create,
    register,
    invite,
    edit,
    get,
    deleteUser,
    new2FA,
    check2FA,
    reset2FA,
    resetPassword,
    getNotifs,
    readNotifs,
  };
}
