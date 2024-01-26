import wrap from './wrap';

function login(values) {
  return wrap(fetch('/cosmos/api/login', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(values)
  }))
}

function me() {
  return fetch('/cosmos/api/me', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  })
  .then((res) => res.json())
}

function logout() {
  return wrap(fetch('/cosmos/api/logout', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

export {
  login,
  logout,
  me
};