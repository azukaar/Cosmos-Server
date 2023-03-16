import wrap from './wrap';

function get() {
  return wrap(fetch('/cosmos/api/config', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function set(values) {
  return wrap(fetch('/cosmos/api/config', {
    method: 'PUT',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function restart() {
  return wrap(fetch('/cosmos/api/restart', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

export {
  get,
  set,
  restart
};