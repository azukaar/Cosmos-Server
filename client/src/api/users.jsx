import wrap from './wrap';

function list() {
  return fetch('/cosmos/api/users', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  })
  .then((res) => res.json())
}

function create(values) {
  alert(JSON.stringify(values))
  return wrap(fetch('/cosmos/api/users', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(values),
  }));
}

function register(values) {
  return fetch('/cosmos/api/register', {
    method: 'POST',
    headers: {
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(values),
    },
  })
  .then((res) => res.json())
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
  return fetch('/cosmos/api/users/'+nickname, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  })
  .then((res) => res.json())
}

function get(nickname) {
  return fetch('/cosmos/api/users/'+nickname, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  })
  .then((res) => res.json())
}

function deleteUser(nickname) {
  return fetch('/cosmos/api/users/'+nickname, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    },
  })
  .then((res) => res.json())
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