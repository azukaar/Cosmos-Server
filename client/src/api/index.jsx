import * as auth from './authentication';
import * as users from './users';
import * as config from './config';
import * as docker from './docker';

import wrap from './wrap';

const getStatus = () => {
  return wrap(fetch('/cosmos/api/status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

const isOnline = () => {
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

const newInstall = (req) => {
  return wrap(fetch('/cosmos/api/newInstall', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(req)
  }))
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