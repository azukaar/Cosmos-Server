import * as auth from './authentication.jsx';
import * as users from './users.jsx';
import * as config from './config.jsx';
import * as docker from './docker.jsx';

import wrap from './wrap';

const getStatus = () => {
  return wrap(fetch('/cosmos/api/status', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
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
};