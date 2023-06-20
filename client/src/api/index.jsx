import * as _auth from './authentication';
import * as _users from './users';
import * as _config from './config';
import * as _docker from './docker';
import * as _market from './market';

import * as authDemo from './authentication.demo';
import * as usersDemo from './users.demo';
import * as configDemo from './config.demo';
import * as dockerDemo from './docker.demo';
import * as indexDemo from './index.demo';
import * as marketDemo from './market.demo';

import wrap from './wrap';

export let CPU_ARCH = 'amd64';
export let CPU_AVX = true;

let getStatus = () => {
  return wrap(fetch('/cosmos/api/status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
  .then(async (response) => {
    CPU_ARCH = response.data.CPU;
    CPU_AVX = response.data.AVX;
    return response
  });
}

let isOnline = () => {
  return fetch('/cosmos/api/status', {
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
  
    return fetch('/cosmos/api/newInstall', requestOptions)
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
    return wrap(fetch('/cosmos/api/newInstall', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(req)
    }))
  }
}

let checkHost = (host) => {
  return fetch('/cosmos/api/dns-check?url=' + host, {
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
  return fetch('/cosmos/api/dns?url=' + host, {
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

const isDemo = import.meta.env.MODE === 'demo';

let auth = _auth;
let users = _users;
let config = _config;
let docker = _docker;
let market = _market;

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
}

export {
  auth,
  users,
  config,
  docker,
  market,
  getStatus,
  newInstall,
  isOnline,
  checkHost,
  getDNS
};