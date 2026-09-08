import wrap, { type ApiResponse, type ApiFetch } from './wrap';

// Master pinned to a constellation device; the Nebula IP is baked into every -peers flag.
export interface SwfsMaster {
  device: string;
  ip: string;
}

// Crontabs are 6-field, seconds-first (the gocron parser the rest of the scheduler uses).
export interface SwfsJobsConfig {
  vacuumEnabled: boolean;
  vacuumCrontab: string;
  garbageThreshold: number;

  ecEnabled: boolean;
  ecCrontab: string;
  ecFullPercent: number;
  ecQuietForSec: number;
  ecMinFreeDiskPercent: number;

  // Repack runs iff ecEnabled: EC shards never vacuum in place, so repack is the only space-reclaim path.
  repackCrontab: string;
  repackDeleteRatio: number;

  monitorCrontab: string;
  diskAlertPercent: number;

  scrubEnabled: boolean;
  scrubCrontab: string;
}

// Same shape as the managed database backup: the restic repository password is held server-side, never returned.
export interface SeaweedFSBackup {
  enabled: boolean;
  repository: string;
  crontab?: string;
  crontabForget?: string;
  retentionPolicy?: string;
  configuredAt?: string;
}

// seaweedFSBytes is THIS instance's share of the node; disk* describe the whole filesystem under the data directory.
export interface SwfsNodeUsage {
  device: string;
  seaweedFSBytes: number;
  volumeCount: number;
  diskAllBytes: number;
  diskUsedBytes: number;
  diskFreeBytes: number;
}

// Space snapshot written by the monitor job on each pass (5 minutes by default); absent before the first pass.
export interface SwfsUsage {
  nodes: SwfsNodeUsage[];
  totalSeaweedFSBytes: number;
  totalDiskFreeBytes: number;
  collectedAt: string;
}

export interface SeaweedFSStatus {
  name: string;
  image: string;
  tags?: string[];
  masters?: SwfsMaster[];
  masterPort: number;
  volumePort: number;
  filerPort: number;
  s3Port: number;
  filerReplicas: number;
  defaultReplication: string;
  indexMode: string;
  volumeSizeLimitMB: number;
  minFreeSpace: string;
  // 0 = unlimited; enforced in whole volumes, so the real bound rounds down to a multiple of volumeSizeLimitMB.
  maxStorageGBPerNode?: number;
  restrictToConstellation: boolean;
  // Blanked by the list and get endpoints; only /status returns the real pair.
  s3AccessKey?: string;
  s3SecretKey?: string;
  // False until the S3 identity is written to the filer, asynchronously after the deployments come up.
  s3Configured?: boolean;
  // Name of the managed database auto-provisioned for the filer store.
  filerDB: string;
  jobs?: SwfsJobsConfig;
  usage?: SwfsUsage | null;
  metaBackup?: SeaweedFSBackup | null;
  status: string;
  pendingUpgradeImage?: string;
  createdAt?: string;
  // Derived from node heartbeats by the read endpoints, never stored.
  mastersLive?: string[];
  mastersDown?: string[];
  quorumOK?: boolean;
  volumeNodes?: string[];
  filerNodes?: string[];
  filerIPs?: string[];
  volumeIPs?: string[];
  // Live nodes matching the instance's tags: the set the volume tier fills and the filers are placed on.
  tagMatchedNodes?: string[];
  volumeDesired?: number;
  filerDesired?: number;
  // Derived live truth (ready | degraded | waiting-for-nodes), as opposed to `status`,
  // which is the creation lifecycle — see SwfsHealth* in src/pro/seaweedfs.go.
  health?: string;
  // Non-localised summary of what is missing; empty when health is ready.
  healthReason?: string;
}

// Unredacted view, including the S3 credentials. The hostname-based endpoint comes
// first: the only one presenting a certificate that validates on an HTTPS cluster.
export interface SeaweedFSConnection {
  instance: SeaweedFSStatus;
  endpoints: string[];
}

export interface SeaweedFSSnapshot {
  id: string;
  shortId: string;
  time: string;
  paths?: string[];
  tags?: string[];
}

export interface SeaweedFSDeleteResult {
  deleted: string;
  dataPurged: boolean;
  filerDBKept: boolean;
  teardownNotes?: string[];
}

export default function createSeaweedFSAPI(apiFetch: ApiFetch) {
  const base = '/cosmos/api/constellation/seaweedfs';

  function list(): Promise<ApiResponse<SeaweedFSStatus[]>> {
    return wrap(apiFetch(base, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Refused with SWF012 when fewer than 3 managers are online: the master tier is 3 pinned masters or nothing.
  function create(values: {
    name: string;
    tags?: string[];
    image?: string;
    filerReplicas?: number;
    indexMode?: string;
    defaultReplication?: string;
    volumeSizeLimitMB?: number;
    minFreeSpace?: string;
    maxStorageGBPerNode?: number;
    restrictToConstellation?: boolean;
  }): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  function get(name: string): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base + '/' + name, {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // purgeData deletes the volume data; keepFilerDB keeps the filer database. Both default to false.
  function remove(name: string, options?: { purgeData?: boolean; keepFilerDB?: boolean }): Promise<ApiResponse<SeaweedFSDeleteResult>> {
    const query: string[] = [];
    if (options && options.purgeData) query.push('purgeData=true');
    if (options && options.keepFilerDB) query.push('keepFilerDB=true');
    return wrap(apiFetch(base + '/' + name + (query.length ? '?' + query.join('&') : ''), {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // The one response carrying the S3 secret — fetched on demand, not part of the list payload.
  function status(name: string): Promise<ApiResponse<SeaweedFSConnection>> {
    return wrap(apiFetch(base + '/' + name + '/status', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Rewrites the filer/S3 routes, which briefly restarts the filers.
  function setRestrict(name: string, restrictToConstellation: boolean): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base + '/' + name + '/restrict', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ restrictToConstellation }),
    }));
  }

  // Full replace of the jobs block, so every field has to be sent.
  function setJobs(name: string, jobs: SwfsJobsConfig): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base + '/' + name + '/jobs', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(jobs),
    }));
  }

  // 0 = unlimited. A lowered cap only stops new volumes being placed; data is never deleted.
  function setStorageCap(name: string, maxStorageGBPerNode: number): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base + '/' + name + '/storage', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ maxStorageGBPerNode }),
    }));
  }

  // The repository path is resolved on the node running the backup job, not the answering node.
  function configureBackup(name: string, values: {
    enabled: boolean;
    repository: string;
    crontab?: string;
    crontabForget?: string;
    retentionPolicy?: string;
  }): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base + '/' + name + '/backup', {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(values),
    }));
  }

  // Snapshots already written stay in the repository but are no longer listed or restorable from here.
  function removeBackup(name: string): Promise<ApiResponse<SeaweedFSStatus>> {
    return wrap(apiFetch(base + '/' + name + '/backup', {
      method: 'DELETE',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Async: the snapshot only appears in a later listing.
  function runBackup(name: string): Promise<ApiResponse<{ status: string }>> {
    return wrap(apiFetch(base + '/' + name + '/backup/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Empty array until the first backup completes, which is not an error.
  function backupSnapshots(name: string): Promise<ApiResponse<SeaweedFSSnapshot[]>> {
    return wrap(apiFetch(base + '/' + name + '/backup/snapshots', {
      method: 'GET',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // For hardware that is GONE: re-replicates the volumes the dead node held.
  function repair(name: string): Promise<ApiResponse<any>> {
    return wrap(apiFetch(base + '/' + name + '/repair', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
    }));
  }

  // Evacuates a LIVE volume server before its tag is removed; an unreachable device is a repair, not a drain.
  function drain(name: string, device: string): Promise<ApiResponse<any>> {
    return wrap(apiFetch(base + '/' + name + '/drain', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ device }),
    }));
  }

  // Rolling upgrade: the deployments roll first, then the pinned masters.
  function upgrade(name: string, image: string): Promise<ApiResponse<any>> {
    return wrap(apiFetch(base + '/' + name + '/upgrade', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ image }),
    }));
  }

  // Omit newDevice to let the server pick an eligible manager itself.
  function replaceMaster(name: string, oldDevice: string, newDevice?: string): Promise<ApiResponse<SeaweedFSStatus>> {
    const body: { oldDevice: string; newDevice?: string } = { oldDevice };
    if (newDevice) body.newDevice = newDevice;
    return wrap(apiFetch(base + '/' + name + '/replace-master', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    }));
  }

  return {
    list, create, get, remove, status, setRestrict, setJobs, setStorageCap,
    configureBackup, removeBackup, runBackup, backupSnapshots,
    repair, drain, upgrade, replaceMaster,
  };
}
