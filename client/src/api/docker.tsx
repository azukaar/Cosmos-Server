import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface ContainerState {
  Status: string;
  Running: boolean;
  Paused: boolean;
  Restarting: boolean;
  Dead: boolean;
  Pid: number;
  ExitCode: number;
  StartedAt: string;
  FinishedAt: string;
  [key: string]: any;
}

export interface Container {
  ID: string;
  Name: string;
  Image: string;
  Created: string;
  State: ContainerState;
  Config: {
    Hostname: string;
    Image: string;
    Env: string[];
    Cmd: string[];
    Labels: Record<string, string>;
    ExposedPorts: Record<string, object>;
    [key: string]: any;
  };
  HostConfig: any;
  NetworkSettings: any;
  Mounts: any[];
  [key: string]: any;
}

export interface DockerNetwork {
  ID: string;
  Name: string;
  Driver: string;
  Scope: string;
  Internal: boolean;
  Attachable: boolean;
  Labels: Record<string, string>;
  [key: string]: any;
}

export interface DockerVolume {
  Name: string;
  Driver: string;
  Mountpoint: string;
  Labels: Record<string, string>;
  Scope: string;
  Options: Record<string, string>;
  [key: string]: any;
}

export default function createDockerAPI(apiFetch: ApiFetch, createWs: (path: string) => WebSocket) {
  function list(): Promise<ApiResponse<Container[]>> {
    return wrap(apiFetch('/cosmos/api/servapps', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function get(containerName: string): Promise<ApiResponse<Container>> {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerName, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function getContainerLogs(containerId: string, searchQuery?: string, limit?: number, lastReceivedLogs?: string, errorOnly?: string): Promise<ApiResponse> {

    if(limit && limit < 50) limit = 50;

    // remove starting / from containerId
    if(containerId.startsWith('/')) {
      containerId = containerId.substring(1);
    }

    const queryParams = new URLSearchParams({
      search: searchQuery || "",
      limit: String(limit || ""),
      lastReceivedLogs: lastReceivedLogs || "",
      errorOnly: errorOnly || "",
    });

    return wrap(apiFetch(`/cosmos/api/servapps/${containerId}/logs?${queryParams}`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }));
  }

  function volumeList(): Promise<ApiResponse<DockerVolume[]>> {
    return wrap(apiFetch('/cosmos/api/volumes', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function volumeDelete(name: string): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/volume/${name}`, {
      method: 'DELETE',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function networkList(): Promise<ApiResponse<DockerNetwork[]>> {
    return wrap(apiFetch('/cosmos/api/networks', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function networkDelete(name: string): Promise<ApiResponse> {
    return wrap(apiFetch(`/cosmos/api/network/${name}`, {
      method: 'DELETE',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  function secure(id: string, res: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/servapps/' + id + '/secure/'+res, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  const newDB = (): Promise<ApiResponse> => {
    return wrap(apiFetch('/cosmos/api/newDB', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  const manageContainer = (containerId: string, action: string): Promise<ApiResponse> => {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerId + '/manage/' + action, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  function migrateHost(values: any): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/migrate-host', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function updateContainer(containerId: string, values: any): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerId + '/update', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function exportContainer(containerId: string, values: any): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerId + '/export', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function listContainerNetworks(containerId: string): Promise<ApiResponse<DockerNetwork[]>> {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerId + '/networks', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  function createNetwork(values: { Name: string; Driver?: string; [key: string]: any }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/networks', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function attachNetwork(containerId: string, networkId: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerId + '/network/' + networkId, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  function detachNetwork(containerId: string, networkId: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/servapps/' + containerId + '/network/' + networkId, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  function createVolume(values: { Name: string; Driver?: string; [key: string]: any }): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/volumes', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    }))
  }

  function attachTerminal(containerId: string, readonly?: boolean): WebSocket {
    return createWs('/cosmos/api/servapps/' + containerId + '/terminal/attach' + (readonly ? '?readonly=true' : ''));
  }

  function createTerminal(containerId: string): WebSocket {
    return createWs('/cosmos/api/servapps/' + containerId + '/terminal/new');
  }

  function createService(serviceData: any, onProgress: (line: string) => void): Promise<ReadableStream> {
    const requestOptions: RequestInit = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(serviceData)
    };

    return apiFetch('/cosmos/api/docker-service', requestOptions)
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

  function pullImage(imageName: string, onProgress: (line: string) => void, ifMissing?: boolean): Promise<ReadableStream> {
    const requestOptions: RequestInit = {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    };

    const imageNameEncoded = encodeURIComponent(imageName);

    return apiFetch(`/cosmos/api/images/${ifMissing ? 'pull-if-missing' : 'pull'}?imageName=${imageNameEncoded}`, requestOptions)
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

  function updateContainerImage(containerName: string, onProgress: (line: string) => void): Promise<ReadableStream> {
    const requestOptions: RequestInit = {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    };

    const containerNameEncoded = encodeURIComponent(containerName);

    return apiFetch(`/cosmos/api/servapps/${containerNameEncoded}/manage/update`, requestOptions)
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

  function autoUpdate(id: string, toggle: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/servapps/' + id + '/auto-update/'+toggle, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  }

  return {
    list,
    get,
    newDB,
    secure,
    manageContainer,
    volumeList,
    volumeDelete,
    networkList,
    networkDelete,
    getContainerLogs,
    updateContainer,
    listContainerNetworks,
    createNetwork,
    attachNetwork,
    detachNetwork,
    createVolume,
    attachTerminal,
    createTerminal,
    createService,
    pullImage,
    autoUpdate,
    updateContainerImage,
    exportContainer,
    migrateHost,
  };
}
