import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export default function createMetricsAPI(apiFetch: ApiFetch) {
  function get(metarr: string[]): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/metrics?metrics=' + metarr.join(','), {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function reset(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/reset-metrics', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function list(): Promise<ApiResponse<Record<string, string>>> {
    return wrap(apiFetch('/cosmos/api/list-metrics', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function events(from: string, to: string, search: string = '', query: string = '', page: string = '', logLevel?: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/events?from=' + from + '&to=' + to + '&search=' + search + '&query=' + query + '&page=' + page + '&logLevel=' + logLevel, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  // --- Alert CRUD ---
  function listAlerts(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/alerts', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }))
  }

  function getAlert(name: string): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/alerts/${encodeURIComponent(name)}`, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }))
  }

  function createAlert(alert: any): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/alerts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(alert),
    }))
  }

  function updateAlert(name: string, alert: any): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/alerts/${encodeURIComponent(name)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(alert),
    }))
  }

  function deleteAlert(name: string): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/alerts/${encodeURIComponent(name)}`, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }))
  }

  return {
    get,
    reset,
    list,
    events,
    listAlerts,
    getAlert,
    createAlert,
    updateAlert,
    deleteAlert,
  };
}
