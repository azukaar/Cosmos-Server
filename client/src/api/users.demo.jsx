import configDemo from './users.demo.json';

function list() {
  return new Promise((resolve, reject) => {
    resolve(configDemo)
  });
}

function create(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function register(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function invite(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function edit(nickname, values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function get(nickname) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function deleteUser(nickname) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function new2FA(nickname) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function check2FA(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function reset2FA(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function resetPassword(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
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
};