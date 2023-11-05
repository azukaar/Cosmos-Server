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

function new2FA(nickname) {
  return wrap(fetch('/cosmos/api/mfa', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

function check2FA(values) {
  return wrap(fetch('/cosmos/api/mfa', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      token: values
    }),
  }))
}

function reset2FA(values) {
  return wrap(fetch('/cosmos/api/mfa', {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      nickname: values
    }),
  }))
}

function resetPassword(values) {
  return wrap(fetch('/cosmos/api/password-reset', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function getNotifs() {
  return wrap(fetch('/cosmos/api/notifications', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

function readNotifs(notifs) {
  return wrap(fetch('/cosmos/api/notifications/read?ids=' + notifs.join(','), {
    method: 'GET',
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
  new2FA,
  check2FA,
  reset2FA,
  resetPassword,
  getNotifs,
  readNotifs,
};