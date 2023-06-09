import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/markets', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

export {
  list,
};