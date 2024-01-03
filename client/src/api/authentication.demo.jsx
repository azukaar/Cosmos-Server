import wrap from './wrap';

function login(values) {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "OK",
    })
  });
}

function me() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "OK",
    })
  });
}

function logout() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "OK",
    })
  });
}

export {
  login,
  logout,
  me
};