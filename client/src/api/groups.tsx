import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface Group {
  name: string;
  permissions: number[];
}

export default function createGroupsAPI(apiFetch: ApiFetch) {
  function list(): Promise<ApiResponse<Record<string, Group>>> {
    return wrap(apiFetch('/cosmos/api/groups', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function create(values: { name: string; permissions?: number[] }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/groups', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function update(id: string | number, values: { name?: string; permissions?: number[] }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/groups/' + id, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function deleteGroup(id: string | number): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/groups/' + id, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  return { list, create, update, deleteGroup };
}
