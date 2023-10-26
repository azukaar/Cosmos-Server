import wrap from './wrap';

function get() {
  return wrap(fetch('/cosmos/api/metrics', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

export {
  get,
};