import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface ManagedLogicalDB {
  name: string;
  role: string;
  // Blanked by the list endpoint; only /connection returns it.
  password?: string;
  createdAt?: string;
}

// The restic repository password is held server-side and never returned.
export interface ManagedDatabaseBackup {
  enabled: boolean;
  repository: string;
  crontab?: string;
  crontabForget?: string;
  retentionPolicy?: string;
  configuredAt?: string;
}

export interface ManagedDatabase {
  name: string;
  engine: string;
  image: string;
  version?: string;
  homeNode: string;
  homeNodeName?: string;
  homeIP: string;
  port: number;
  restrictToConstellation: boolean;
  superUser: string;
  superPassword?: string;
  createdAt?: string;
  databases?: Record<string, ManagedLogicalDB>;
  backup?: ManagedDatabaseBackup;
  // Merged in from node heartbeats by the list endpoint, never stored.
  live?: boolean;
  homeNodeUp?: boolean;
}

// paths are one <database>.dump per logical database plus a globals.sql.
export interface ManagedDatabaseSnapshot {
  id: string;
  shortId: string;
  time: string;
  paths?: string[];
  tags?: string[];
}

export interface ManagedDatabaseRestoreCredential {
  database: string;
  user: string;
  password: string;
  url: string;
}

// Credentials are returned exactly once: rotatable afterwards, not recoverable.
export interface ManagedDatabaseRestoreResult {
  database: ManagedDatabase;
  source: string;
  snapshot: string;
  restored: string[];
  credentials: Record<string, ManagedDatabaseRestoreCredential>;
  failures?: { database: string; error: string }[];
}

export interface ManagedDatabaseConnection {
  name: string;
  engine: string;
  host: string;
  port: number;
  homeNode: string;
  homeNodeName?: string;
  user: string;
  password: string;
  url: string;
  databases: Record<string, { database: string; user: string; password: string; url: string }>;
}

export default function createDatabasesAPI(apiFetch: ApiFetch) {
  const base = '/cosmos/api/constellation/databases';

  function list(): Promise<ApiResponse<ManagedDatabase[]>> {
    return wrap(apiFetch(base, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Send restrictToConstellation explicitly: the server defaults it to true when absent.
  function create(values: { name: string; engine?: string; image?: string; port?: number; restrictToConstellation?: boolean }): Promise<ApiResponse<ManagedDatabase>> {
    return wrap(apiFetch(base, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  // removeVolume also destroys the data directory (no replica is kept anywhere).
  function remove(name: string, removeVolume?: boolean): Promise<ApiResponse> {
    return wrap(apiFetch(base + '/' + name + (removeVolume ? '?removeVolume=true' : ''), {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function update(name: string, values: { restrictToConstellation?: boolean }): Promise<ApiResponse<ManagedDatabase>> {
    return wrap(apiFetch(base + '/' + name, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function connection(name: string): Promise<ApiResponse<ManagedDatabaseConnection>> {
    return wrap(apiFetch(base + '/' + name + '/connection', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function createLogical(name: string, values: { database: string; role?: string }): Promise<ApiResponse<any>> {
    return wrap(apiFetch(base + '/' + name + '/databases', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function removeLogical(name: string, database: string): Promise<ApiResponse> {
    return wrap(apiFetch(base + '/' + name + '/databases/' + database, {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  function rotateLogical(name: string, database: string): Promise<ApiResponse<any>> {
    return wrap(apiFetch(base + '/' + name + '/databases/' + database + '/rotate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // The repository path is resolved on the database's home node, not the answering node.
  function configureBackup(name: string, values: {
    enabled: boolean;
    repository: string;
    crontab?: string;
    crontabForget?: string;
    retentionPolicy?: string;
  }): Promise<ApiResponse<ManagedDatabase>> {
    return wrap(apiFetch(base + '/' + name + '/backup', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  // Async: the snapshot only appears in a later listing; 502 when the home node is unreachable.
  function runBackup(name: string): Promise<ApiResponse<{ status: string }>> {
    return wrap(apiFetch(base + '/' + name + '/backup/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Empty array until the first backup completes, which is not an error.
  function backupSnapshots(name: string): Promise<ApiResponse<ManagedDatabaseSnapshot[]>> {
    return wrap(apiFetch(base + '/' + name + '/backup/snapshots', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Restores into a NEW instance, leaving the source untouched; omitting databases restores all of them.
  function restore(values: {
    sourceName: string;
    snapshotId: string;
    newName: string;
    databases?: string[];
    port?: number;
    restrictToConstellation?: boolean;
  }): Promise<ApiResponse<ManagedDatabaseRestoreResult>> {
    return wrap(apiFetch(base + '/restore', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  return {
    list, create, update, remove, connection, createLogical, removeLogical, rotateLogical,
    configureBackup, runBackup, backupSnapshots, restore,
  };
}
