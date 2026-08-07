import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface DiskInfo {
  Path: string;
  Name: string;
  Size: number;
  Used: number;
  [key: string]: any;
}

export interface MountRequest {
  path: string;
  mountPoint: string;
  permanent: boolean;
  netDisk?: boolean;
  chown?: string;
}

export interface UnmountRequest {
  mountPoint: string;
  permanent: boolean;
  chown?: string;
}

export interface MergeRequest {
  Branches: string[];
  MountPoint: string;
  Permanent: boolean;
  Chown?: string;
  Opts?: string;
}

export interface SnapRAIDConfig {
  Name: string;
  Enabled: boolean;
  Data: Record<string, string>;
  Parity: string[];
  SyncCrontab: string;
  ScrubCrontab: string;
  CheckOnFix: boolean;
}

export interface RaidCreateRequest {
  name: string;
  level: string;
  devices: string[];
  spares?: string[];
  metadata?: string;
}

export default function createStorageAPI(apiFetch: ApiFetch) {
  const mounts = {
    list: (): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/mounts', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },

    mount: ({path, mountPoint, permanent, netDisk, chown}: MountRequest): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/mount', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({path, mountPoint, permanent, netDisk, chown})
      }))
    },

    unmount: ({mountPoint, permanent, chown}: UnmountRequest): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/unmount', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({mountPoint, permanent, chown})
      }))
    },

    merge: (args: MergeRequest): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/merge', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(args)
      }))
    },
  };
  const raid = {
    list: (): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/storage/raid', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        },
      }))
    },

    create: ({name, level, devices, spares, metadata}: RaidCreateRequest): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/storage/raid', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({name, level, devices, spares, metadata})
      }))
    },

    delete: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/storage/raid/${name}`, {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json'
        }
      }))
    },

    status: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/storage/raid/${name}/status`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }))
    },

    addDevice: (name: string, device: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/storage/raid/${name}/device`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({device})
      }))
    },

    replaceDevice: (name: string, oldDevice: string, newDevice: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/storage/raid/${name}/replace`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({oldDevice, newDevice})
      }))
    },

    resize: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/storage/raid/${name}/resize`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      }))
    }
  };

  const disks = {
    list: (): Promise<ApiResponse<DiskInfo[]>> => {
      return wrap(apiFetch('/cosmos/api/disks', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },

    smartDef: (): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/smart-def', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },

    format({disk, format, password}: { disk: string; format: string; password?: string }, onProgress: (line: string) => void): Promise<ReadableStream> {
      const requestOptions: RequestInit = {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({disk, format, password})
      };

      return apiFetch('/cosmos/api/disks/format', requestOptions)
        .then(response => {
          if (!response.ok) {
            throw new Error(response.statusText);
          }

          const reader = response.body!.getReader();

          return new ReadableStream({
            start(controller) {
              function read(): any {
                return reader.read().then(({ done, value }) => {
                  if (done) {
                    controller.close();
                    return;
                  }
                  let text = new TextDecoder().decode(value);
                  let lines = text.split('\n');
                  for (let line of lines) {
                    if (line) {
                      onProgress(line);
                    }
                  }
                  controller.enqueue(value);
                  return read();
                });
              }
              return read();
            }
          });
        });
    }
  };

  const snapRAID = {
    create: (args: Partial<SnapRAIDConfig>): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/snapraid', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(args)
      }))
    },
    update: (name: string, args: Partial<SnapRAIDConfig>): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/snapraid/' + name, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(args)
      }))
    },
    delete: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch('/cosmos/api/snapraid/' + name, {
        method: 'DELETE',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },
    list: (args?: any): Promise<ApiResponse<SnapRAIDConfig[]>> => {
      return wrap(apiFetch('/cosmos/api/snapraid', {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(args)
      }))
    },
    sync: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/snapraid/${name}/sync`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },
    fix: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/snapraid/${name}/fix`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },
    enable: (name: string, enable: boolean): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/snapraid/${name}/` + (enable ? 'enable' : 'disable'), {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },
    scrub: (name: string): Promise<ApiResponse> => {
      return wrap(apiFetch(`/cosmos/api/snapraid/${name}/scrub`, {
        method: 'GET',
        headers: {
            'Content-Type': 'application/json'
        },
      }))
    },
  };

  const listDir = (storage: string, path: string): Promise<ApiResponse> => {
    return wrap(apiFetch(`/cosmos/api/list-dir?storage=${storage}&path=${path}`, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  const newFolder = (storage: string, path: string, folder: string): Promise<ApiResponse> => {
    return wrap(apiFetch(`/cosmos/api/new-dir?storage=${storage}&path=${path}&folder=${folder}`, {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  return {
    mounts,
    disks,
    snapRAID,
    raid,
    newFolder,
    listDir,
  };
}
