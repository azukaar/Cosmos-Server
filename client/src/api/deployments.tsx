import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface Deployment {
  name: string;
  // Exactly one replica mode: fixed count, load-based autoscale, or fill (one per eligible node).
  replicas?: number;
  minReplicas?: number;
  maxReplicas?: number;
  replicaFill?: boolean;
  // empty = bare + lazy containers: the idle replica sleeps until the next request.
  replicaFillMode?: 'full' | 'bare' | 'empty';
  strategy?: 'round-robin' | 'least-busy';
  tags?: string[];
  storage?: string[];
  compose: any;
}

export interface DeploymentHealth {
  desired: number;
  actual: number;
  minReplicas?: number;
  maxReplicas?: number;
  replicaFill?: boolean;
  replicaFillMode?: 'full' | 'bare' | 'empty';
  nodes: string[];
  broken: boolean;
  brokenReason?: string;
}

export default function createDeploymentsAPI(apiFetch: ApiFetch) {
  function list(): Promise<ApiResponse<Record<string, Deployment>>> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function get(name: string): Promise<ApiResponse<Deployment>> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments/' + name, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function create(values: Deployment): Promise<ApiResponse<Deployment>> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function update(name: string, values: Deployment): Promise<ApiResponse<Deployment>> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments/' + name, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function remove(name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments/' + name, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function health(): Promise<ApiResponse<Record<string, DeploymentHealth>> & { brokenNodes?: Record<string, string> }> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments/health', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function unbroke(name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/constellation/deployments/' + name + '/unbroke', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function unbrokeNode(name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/constellation/nodes/' + name + '/unbroke', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  return { list, get, create, update, remove, health, unbroke, unbrokeNode };
}
