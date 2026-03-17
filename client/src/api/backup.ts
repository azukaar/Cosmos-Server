import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface BackupConfig {
  name: string;
  source: string;
  repository: string;
  password: string;
  crontab?: string;
  tags?: string[];
  exclude?: string[];
  autoStopContainers?: boolean;
}

export interface RestoreConfig {
  snapshotId: string;
  target: string;
  include?: string[];
}

export default function createBackupsAPI(apiFetch: ApiFetch) {
  function listSnapshots(name: string) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}/snapshots`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function listSnapshotsFromRepo(name: string) {
    return wrap(apiFetch(`/cosmos/api/backups-repository/${name}/snapshots`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function listRepo() {
    return wrap(apiFetch(`/cosmos/api/backups-repository`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function listFolders(name: string, snapshot: string, path?: string) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}/${snapshot}/folders?path=${path || '/'}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function restoreBackup(name: string, config: RestoreConfig) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}/restore`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(config)
    }))
  }

  function addBackup(config: BackupConfig) {
    return wrap(apiFetch('/cosmos/api/backups', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(config)
    }))
  }

  function editBackup(config: BackupConfig) {
    return wrap(apiFetch('/cosmos/api/backups/edit', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(config)
    }))
  }

  function removeBackup(name: string, deleteRepo: boolean = false) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function forgetSnapshot(name: string, snapshot: string, deleteRepo: boolean = false) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}/${snapshot}/forget`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ deleteRepo })
    }))
  }

  function subfolderRestoreSize(name: string, snapshot: string, path: string) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}/${snapshot}/subfolder-restore-size?path=` +encodeURIComponent(path), {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function backupNow(name: string) {
    return wrap(apiFetch('/cosmos/api/jobs/run', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        scheduler: "Restic",
        name: "Restic backup " + name,
      })
    }))
  }

  function forgetNow(name: string) {
    return wrap(apiFetch('/cosmos/api/jobs/run', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        scheduler: "Restic",
        name: "Restic forget " + name,
      })
    }))
  }

  function unlockRepository(name: string) {
    return wrap(apiFetch(`/cosmos/api/backups/${name}/unlock`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  function repoStats(name: string) {
    return wrap(apiFetch(`/cosmos/api/backups-repository/${name}/stats`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  }

  return {
    listSnapshots,
    listFolders,
    restoreBackup,
    addBackup,
    removeBackup,
    listRepo,
    listSnapshotsFromRepo,
    forgetSnapshot,
    editBackup,
    backupNow,
    forgetNow,
    subfolderRestoreSize,
    unlockRepository,
    repoStats
  };
}
