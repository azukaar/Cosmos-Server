import * as _auth from './authentication';
import * as _users from './users';
import * as _config from './config';
import * as _docker from './docker';

import * as authDemo from './authentication.demo';
import * as usersDemo from './users.demo';
import * as configDemo from './config.demo';
import * as dockerDemo from './docker.demo';
import * as indexDemo from './index.demo';

import wrap from './wrap';

let getStatus = () => {
  return wrap(fetch('/cosmos/api/status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
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

let newInstall = (req) => {
  return wrap(fetch('/cosmos/api/newInstall', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(req)
  }))
}

const isDemo = import.meta.env.MODE === 'demo';

let auth = _auth;
let users = _users;
let config = _config;
let docker = _docker;

if(isDemo) {
  auth = authDemo;
  users = usersDemo;
  config = configDemo;
  docker = dockerDemo;
  getStatus = indexDemo.getStatus;
  newInstall = indexDemo.newInstall;
  isOnline = indexDemo.isOnline;
}

export {
  auth,
  users,
  config,
  docker,
  getStatus,
  newInstall,
  isOnline
};