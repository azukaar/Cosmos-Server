import { createApiClient } from '../../client/src/api/client';
import createAuthAPI from '../../client/src/api/authentication';
import createUsersAPI from '../../client/src/api/users';
import createConfigAPI from '../../client/src/api/config';
import createDockerAPI from '../../client/src/api/docker';
import createMarketAPI from '../../client/src/api/market';
import createConstellationAPI from '../../client/src/api/constellation';
import createMetricsAPI from '../../client/src/api/metrics';
import createStorageAPI from '../../client/src/api/storage';
import createCronAPI from '../../client/src/api/cron';
import createRcloneAPI from '../../client/src/api/rclone';
import createBackupsAPI from '../../client/src/api/backup';
import createApiTokensAPI from '../../client/src/api/apiTokens';
import createOpenIDAPI from '../../client/src/api/openid';
import wrap from '../../client/src/api/wrap';

export function createClient({ baseUrl, token }: { baseUrl: string; token: string }) {
  const { apiFetch, createWs } = createApiClient(baseUrl, token);

  return {
    auth: createAuthAPI(apiFetch),
    users: createUsersAPI(apiFetch),
    config: createConfigAPI(apiFetch),
    docker: createDockerAPI(apiFetch, createWs),
    market: createMarketAPI(apiFetch),
    constellation: createConstellationAPI(apiFetch),
    metrics: createMetricsAPI(apiFetch),
    storage: createStorageAPI(apiFetch),
    cron: createCronAPI(apiFetch, createWs),
    rclone: createRcloneAPI(apiFetch),
    backups: createBackupsAPI(apiFetch),
    apiTokens: createApiTokensAPI(apiFetch),
    openid: createOpenIDAPI(apiFetch),

    getStatus: () => {
      return wrap(apiFetch('/cosmos/api/status', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      }));
    },

    isOnline: () => {
      return apiFetch('/cosmos/api/status', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      }).then(async (response) => {
        const rep = await response.json();
        if (response.status === 200) return rep;
        const e: any = new Error(rep.message);
        e.status = response.status;
        throw e;
      });
    },

    uploadImage: (file: any, name: string) => {
      const formData = new FormData();
      formData.append('image', file);
      return wrap(apiFetch('/cosmos/api/upload/' + name, {
        method: 'POST',
        body: formData
      }));
    },

    restartServer: () => {
      return wrap(apiFetch('/cosmos/api/restart-server', {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' },
      }));
    },

    terminal: (startCmd: string = "bash") => {
      return createWs('/cosmos/api/terminal/' + startCmd);
    },

    forceAutoUpdate: () => {
      return wrap(apiFetch('/cosmos/api/force-server-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      }));
    },

    checkHost: (host: string) => {
      return apiFetch('/cosmos/api/dns-check?url=' + host, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      }).then(async (response) => {
        const rep = await response.json();
        if (response.status === 200) return rep;
        const e: any = new Error(rep.message);
        e.status = response.status;
        throw e;
      });
    },

    getDNS: (host: string) => {
      return apiFetch('/cosmos/api/dns?url=' + host, {
        method: 'GET',
        headers: { 'Content-Type': 'application/json' }
      }).then(async (response) => {
        const rep = await response.json();
        if (response.status === 200) return rep;
        const e: any = new Error(rep.message);
        e.status = response.status;
        throw e;
      });
    },
  };
}

export { createApiClient };

// Re-export types for SDK consumers
export type { ApiResponse, ApiFetch } from '../../client/src/api/wrap';
export type { LoginRequest, UserMe } from '../../client/src/api/authentication';
export type { User } from '../../client/src/api/users';
export type { Route, Operation } from '../../client/src/api/config';
export type { Container, ContainerState, DockerNetwork, DockerVolume } from '../../client/src/api/docker';
export type { MarketApp, MarketResult } from '../../client/src/api/market';
export type { CreateAPITokenRequest, CreateAPITokenResponse, UpdateAPITokenRequest, APITokenConfig } from '../../client/src/api/apiTokens';
export type { OpenIDClient } from '../../client/src/api/openid';
export type { ConstellationDevice } from '../../client/src/api/constellation';
export type { CronJob } from '../../client/src/api/cron';
export type { DiskInfo, MountRequest, UnmountRequest, MergeRequest, SnapRAIDConfig, RaidCreateRequest } from '../../client/src/api/storage';
export type { BackupConfig, RestoreConfig } from '../../client/src/api/backup';
