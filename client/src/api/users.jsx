
function list() {
  return fetch('/cosmos/api/users', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  })
  .then((res) => res.json())
}

export {
  list,
};