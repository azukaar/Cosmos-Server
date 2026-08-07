import type { ApiFetch } from './wrap';

export function createApiClient(baseUrl: string = '', token: string | null = null): {
  apiFetch: ApiFetch;
  createWs: (path: string) => WebSocket;
} {
  function apiFetch(path: string, options: RequestInit = {}): Promise<Response> {
    const headers: Record<string, string> = { ...(options.headers as Record<string, string>) };
    if (token) headers['Authorization'] = `Bearer ${token}`;
    return fetch(baseUrl + path, { ...options, headers });
  }

  function createWs(path: string): WebSocket {
    let wsBase: string;
    if (baseUrl) {
      wsBase = baseUrl.replace(/^http/, 'ws');
    } else if (typeof window !== 'undefined') {
      wsBase = (window.location.protocol === 'https:' ? 'wss://' : 'ws://') + window.location.host;
    } else {
      throw new Error('baseUrl is required for WebSocket in non-browser environments');
    }
    const url = wsBase + path + (token ? (path.includes('?') ? '&' : '?') + 'token=' + token : '');
    return new WebSocket(url);
  }

  return { apiFetch, createWs };
}
