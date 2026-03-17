import createAuthAPI from './authentication';
import createUsersAPI from './users';
import createConfigAPI from './config';
import createDockerAPI from './docker';
import createMarketAPI from './market';
import createConstellationAPI from './constellation';
import createMetricsAPI from './metrics';
import createStorageAPI from './storage';
import createCronAPI from './cron';
import createRcloneAPI from './rclone';
import createBackupsAPI from './backup';
import createApiTokensAPI from './apiTokens';

import * as authDemo from './authentication.demo';
import * as usersDemo from './users.demo';
import * as configDemo from './config.demo';
import * as dockerDemo from './docker.demo';
import * as indexDemo from './index.demo';
import * as marketDemo from './market.demo';
import * as constellationDemo from './constellation.demo';
import * as metricsDemo from './metrics.demo';
import * as storageDemo from './storage.demo';
import * as cronDemo from './cron.demo';
import * as backupDemo from './backup.demo';

import wrap from './wrap';
import { createApiClient } from './client';
import { redirectToLocal } from '../utils/indexs';

export let CPU_ARCH = 'amd64';
export let CPU_AVX = true;

export let HOME_BACKGROUND;
export let PRIMARY_COLOR;
export let SECONDARY_COLOR;

export let FIRST_LOAD = false;

// --- SDK entry point ---
export function createClient({ baseUrl, token }) {
  const { apiFetch, createWs } = createApiClient(baseUrl, token);

  const client = {
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

    getStatus: () => {
      return wrap(apiFetch('/cosmos/api/status', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }));
    },

    isOnline: () => {
      return apiFetch('/cosmos/api/status', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }).then(async (response) => {
        let rep;
        try {
          rep = await response.json();
        } catch {
          throw new Error('Server error');
        }
        if (response.status == 200) {
          return rep;
        }
        const e = new Error(rep.message);
        e.status = response.status;
        throw e;
      });
    },

    checkHost: (host) => {
      return apiFetch('/cosmos/api/dns-check?url=' + host, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }).then(async (response) => {
        let rep;
        try {
          rep = await response.json();
        } catch {
          throw new Error('Server error');
        }
        if (response.status == 200) {
          return rep;
        }
        const e = new Error(rep.message);
        e.status = response.status;
        e.message = rep.message;
        throw e;
      });
    },

    getDNS: (host) => {
      return apiFetch('/cosmos/api/dns?url=' + host, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      }).then(async (response) => {
        let rep;
        try {
          rep = await response.json();
        } catch {
          throw new Error('Server error');
        }
        if (response.status == 200) {
          return rep;
        }
        const e = new Error(rep.message);
        e.status = response.status;
        e.message = rep.message;
        throw e;
      });
    },

    uploadImage: (file, name) => {
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
        headers: {
          'Content-Type': 'application/json'
        },
      }));
    },

    terminal: (startCmd = "bash") => {
      return createWs('/cosmos/api/terminal/' + startCmd);
    },

    forceAutoUpdate: () => {
      return wrap(apiFetch('/cosmos/api/force-server-update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
      }));
    },
  };

  return client;
}

// --- Default browser client (backward compat) ---
const { apiFetch: defaultFetch, createWs: defaultWs } = createApiClient();

let getStatus = (initial) => {
  return wrap(defaultFetch('/cosmos/api/status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }), initial)
  .then(async (response) => {
    CPU_ARCH = response.data.CPU;
    CPU_AVX = response.data.AVX;
    HOME_BACKGROUND = response.data.homepage.Background;
    PRIMARY_COLOR = response.data.theme.PrimaryColor;
    SECONDARY_COLOR = response.data.theme.SecondaryColor;
    FIRST_LOAD = true;
    return response
  }).catch((response) => {
    const urlSearch = encodeURIComponent(window.location.search);
    const redirectToURL = (window.location.pathname + urlSearch);

    if(response.status != 'OK') {
      if(
          window.location.href.indexOf('/cosmos-ui/oauth/callback') == -1 &&
          window.location.href.indexOf('/cosmos-ui/newInstall') == -1 &&
          window.location.href.indexOf('/cosmos-ui/login') == -1 &&
          window.location.href.indexOf('/cosmos-ui/loginmfa') == -1 &&
          window.location.href.indexOf('/cosmos-ui/newmfa') == -1 &&
          window.location.href.indexOf('/cosmos-ui/register') == -1 &&
          window.location.href.indexOf('/cosmos-ui/forgot-password') == -1) {
        if(response.status == 'NEW_INSTALL') {
            redirectToLocal('/cosmos-ui/newInstall');
        } else if (response.status == 'error' && response.code == "HTTP004") {
            redirectToLocal('/cosmos-ui/login?redirect=' + redirectToURL);
        } else if (response.status == 'error' && response.code == "HTTP006") {
            redirectToLocal('/cosmos-ui/loginmfa?redirect=' + redirectToURL);
        } else if (response.status == 'error' && response.code == "HTTP007") {
            redirectToLocal('/cosmos-ui/newmfa?redirect=' + redirectToURL);
        }
      } else {
        return "nothing";
      }
    }

    return "NOT_AVAILABLE";
  });
}

let isOnline = () => {
  return defaultFetch('/cosmos/api/status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }).then(async (response) => {
    let rep;
    try {
      rep = await response.json();
    } catch {
      throw new Error('Server error');
    }
    if (response.status == 200) {
      return rep;
    }
    const e = new Error(rep.message);
    e.status = response.status;
    throw e;
  });
}

let newInstall = (req, onProgress) => {
  if(req.step == '2') {
    const requestOptions = {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(req)
    };

    return defaultFetch('/cosmos/api/newInstall', requestOptions)
      .then(response => {
        if (!response.ok) {
          throw new Error(response.statusText);
        }

        // The response body is a ReadableStream. This code reads the stream and passes chunks to the callback.
        const reader = response.body.getReader();

        // Read the stream and pass chunks to the callback as they arrive
        return new ReadableStream({
          start(controller) {
            function read() {
              return reader.read().then(({ done, value }) => {
                if (done) {
                  controller.close();
                  return;
                }
                // Decode the UTF-8 text
                let text = new TextDecoder().decode(value);
                // Split by lines in case there are multiple lines in one chunk
                let lines = text.split('\n');
                for (let line of lines) {
                  if (line) {
                    // Call the progress callback
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
      }).catch((e) => {
        console.error(e);
      });
  } else {
    return wrap(defaultFetch('/cosmos/api/newInstall', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(req)
    }))
  }
}

let checkHost = (host) => {
  return defaultFetch('/cosmos/api/dns-check?url=' + host, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }).then(async (response) => {
    let rep;
    try {
      rep = await response.json();
    } catch {
      throw new Error('Server error');
    }
    if (response.status == 200) {
      return rep;
    }
    const e = new Error(rep.message);
    e.status = response.status;
    e.message = rep.message;
    throw e;
  });
}

let getDNS = (host) => {
  return defaultFetch('/cosmos/api/dns?url=' + host, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }).then(async (response) => {
    let rep;
    try {
      rep = await response.json();
    } catch {
      throw new Error('Server error');
    }
    if (response.status == 200) {
      return rep;
    }
    const e = new Error(rep.message);
    e.status = response.status;
    e.message = rep.message;
    throw e;
  });
}

let uploadImage = (file, name) => {
  const formData = new FormData();
  formData.append('image', file);
  return wrap(defaultFetch('/cosmos/api/upload/' + name, {
    method: 'POST',
    body: formData
  }));
};

let restartServer = () => {
  return wrap(defaultFetch('/cosmos/api/restart-server', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }));
}

function terminal(startCmd = "bash") {
  return defaultWs('/cosmos/api/terminal/' + startCmd);
}

function forceAutoUpdate() {
  return wrap(defaultFetch('/cosmos/api/force-server-update', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
  }));
}

const isDemo = import.meta.env.MODE === 'demo';

let auth = createAuthAPI(defaultFetch);
let users = createUsersAPI(defaultFetch);
let config = createConfigAPI(defaultFetch);
let docker = createDockerAPI(defaultFetch, defaultWs);
let market = createMarketAPI(defaultFetch);
let constellation = createConstellationAPI(defaultFetch);
let metrics = createMetricsAPI(defaultFetch);
let storage = createStorageAPI(defaultFetch);
let cron = createCronAPI(defaultFetch, defaultWs);
let rclone = createRcloneAPI(defaultFetch);
let backups = createBackupsAPI(defaultFetch);
let apiTokens = createApiTokensAPI(defaultFetch);

if(isDemo) {
  auth = authDemo;
  users = usersDemo;
  config = configDemo;
  docker = dockerDemo;
  market = marketDemo;
  getStatus = indexDemo.getStatus;
  newInstall = indexDemo.newInstall;
  isOnline = indexDemo.isOnline;
  checkHost = indexDemo.checkHost;
  getDNS = indexDemo.getDNS;
  uploadImage = indexDemo.uploadImage;
  restartServer = indexDemo.restartServer;
  terminal = indexDemo.terminal;
  constellation = constellationDemo;
  metrics = metricsDemo;
  storage = storageDemo;
  cron = cronDemo;
  backups = backupDemo;
}

export {
  auth,
  users,
  config,
  docker,
  market,
  constellation,
  getStatus,
  newInstall,
  isOnline,
  checkHost,
  getDNS,
  metrics,
  uploadImage,
  storage,
  terminal,
  cron,
  rclone,
  restartServer,
  forceAutoUpdate,
  backups,
  apiTokens,
};
