
function login(values) {
  return fetch('/cosmos/api/login', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(values)
  })
  .then((res) => res.json())
}

function me() {
  return fetch('/cosmos/api/me/', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  })
  .then((res) => res.json())
}

export {
  login,
  me
};