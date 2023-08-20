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

export {
  list,
  addDevice,
};