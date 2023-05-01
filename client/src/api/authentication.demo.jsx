import wrap from './wrap';

function login(values) {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function me() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function logout() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

export {
  login,
  logout,
  me
};