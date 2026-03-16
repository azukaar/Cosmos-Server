import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/api-tokens', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

function create(request) {
  return wrap(fetch('/cosmos/api/api-tokens', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(request),
  }))
}

function remove(name) {
  return wrap(fetch('/cosmos/api/api-tokens', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ name }),
  }))
}

export {
  list,
  create,
  remove,
};
