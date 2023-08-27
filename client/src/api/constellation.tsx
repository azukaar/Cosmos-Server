import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/constellation/devices', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function addDevice(device) {
  return wrap(fetch('/cosmos/api/constellation/devices', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(device),
  }))
}

function restart() {
  return wrap(fetch('/cosmos/api/constellation/restart', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

function getConfig() {
  return wrap(fetch('/cosmos/api/constellation/config', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

function getLogs() {
  return wrap(fetch('/cosmos/api/constellation/logs', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

export {
  list,
  addDevice,
  restart,
  getConfig,
  getLogs,
};