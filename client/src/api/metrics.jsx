import wrap from './wrap';

function get(metarr) {
  return wrap(fetch('/cosmos/api/metrics?metrics=' + metarr.join(','), {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function reset() {
  return wrap(fetch('/cosmos/api/reset-metrics', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function list() {
  return wrap(fetch('/cosmos/api/list-metrics', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

export {
  get,
  reset,
  list,
};