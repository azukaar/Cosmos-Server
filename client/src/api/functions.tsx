import wrap, { type ApiResponse, type ApiFetch } from './wrap';

// Cloud Functions (Pro): a package in a cluster registry plus run settings; the server derives a lazy fill deployment from it.

export interface FunctionSource {
  registry: string;
  access: string;
  package: string;
  // Pinned version; on create it is the version to deploy ("" = latest).
  version?: string;
  token?: string;
}

export interface FunctionLimits {
  memoryMB?: number;
  cpu?: number;
  timeoutSec?: number;
  idleTTL?: string;
}

export interface FunctionCronTrigger {
  name: string;
  enabled: boolean;
  crontab: string;
  method?: string;
  path?: string;
  body?: string;
}

export interface FunctionRelease {
  rev: number;
  version: string;
  deployedAt: string;
  deployedBy?: string;
}

export interface CloudFunction {
  name: string;
  runtime: string;
  image?: string;
  source: FunctionSource;
  handler: string;
  entry?: string;
  env?: Record<string, string>;
  tags?: string[];
  limits: FunctionLimits;
  route: any;
  triggers: { cron?: FunctionCronTrigger[] };
  rev: number;
  releases?: FunctionRelease[];
  status: 'created' | 'deployed';
  createdAt: string;
  updatedAt: string;
  // Derived names, filled by the server.
  deployment: string;
  routeName: string;
  container: string;
  imageUsed: string;
}

export interface FunctionRuntime {
  key: string;
  label: string;
  image: string;
  registryType: 'npm' | 'pypi';
  deprecated?: boolean;
}

export interface FunctionVersion {
  version: string;
  createdAt: string;
  files: number;
  latest: boolean;
  active: boolean;
}

export interface FunctionInvokeResult {
  node: string;
  status: number;
  body: string;
  durationMs: number;
}

export default function createFunctionsAPI(apiFetch: ApiFetch) {
  const base = '/cosmos/api/constellation/functions';
  const json = { 'Content-Type': 'application/json' };

  function runtimes(): Promise<ApiResponse<FunctionRuntime[]>> {
    return wrap(apiFetch('/cosmos/api/constellation/function-runtimes', { method: 'GET', headers: json }));
  }

  function list(): Promise<ApiResponse<CloudFunction[]>> {
    return wrap(apiFetch(base, { method: 'GET', headers: json }));
  }

  function get(name: string): Promise<ApiResponse<CloudFunction>> {
    return wrap(apiFetch(base + '/' + encodeURIComponent(name), { method: 'GET', headers: json }));
  }

  function create(values: Partial<CloudFunction>): Promise<ApiResponse<CloudFunction> & { warning?: string }> {
    return wrap(apiFetch(base, { method: 'POST', headers: json, body: JSON.stringify(values) }));
  }

  function update(name: string, values: Partial<CloudFunction>): Promise<ApiResponse<CloudFunction> & { warning?: string }> {
    return wrap(apiFetch(base + '/' + encodeURIComponent(name), { method: 'PUT', headers: json, body: JSON.stringify(values) }));
  }

  function remove(name: string): Promise<ApiResponse> {
    return wrap(apiFetch(base + '/' + encodeURIComponent(name), { method: 'DELETE', headers: json }));
  }

  function deploy(name: string, version: string): Promise<ApiResponse<CloudFunction>> {
    return wrap(apiFetch(base + '/' + encodeURIComponent(name) + '/deploy', {
      method: 'POST', headers: json, body: JSON.stringify({ version }),
    }));
  }

  function versions(name: string): Promise<ApiResponse<FunctionVersion[]> & { latest?: string; active?: string }> {
    return wrap(apiFetch(base + '/' + encodeURIComponent(name) + '/versions', { method: 'GET', headers: json }));
  }

  function invoke(name: string, req: { method?: string; path?: string; body?: string }): Promise<ApiResponse<FunctionInvokeResult>> {
    return wrap(apiFetch(base + '/' + encodeURIComponent(name) + '/invoke', {
      method: 'POST', headers: json, body: JSON.stringify(req),
    }));
  }

  return { runtimes, list, get, create, update, remove, deploy, versions, invoke };
}
