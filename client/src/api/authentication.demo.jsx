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

function sudo(values) {
  return Promise.resolve({ "status": "OK" });
}

export {
  login,
  logout,
  me,
  sudo
};