import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/users', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function create(values) {
  return wrap(fetch('/cosmos/api/users', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(values),
  }));
}

function register(values) {
  return wrap(fetch('/cosmos/api/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function invite(values) {
  return wrap(fetch('/cosmos/api/invite', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function edit(nickname, values) {
  return wrap(fetch('/cosmos/api/users/'+nickname, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function get(nickname) {
  return wrap(fetch('/cosmos/api/users/'+nickname, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

function deleteUser(nickname) {
  return wrap(fetch('/cosmos/api/users/'+nickname, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

export {
  list,
  create,
  register,
  invite,
  edit,
  get,
  deleteUser,
};