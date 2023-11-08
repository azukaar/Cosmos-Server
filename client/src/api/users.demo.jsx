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

function getNotifs() {
  return new Promise((resolve, reject) => {
    resolve({"data":[{"ID":"654b0fccce74bf6f8c8ccc61","Title":"Container Update","Message":"Container Prowlarr updated to the latest version!","Icon":"","Link":"/cosmos-ui/servapps/containers/Prowlarr","Date":"2023-11-08T04:33:39.041Z","Level":"info","Read":false,"Recipient":"admin","Actions":null},{"ID":"654b0fccce74bf6f8c8ccc60","Title":"Container Update","Message":"Container Lidarr updated to the latest version!","Icon":"","Link":"/cosmos-ui/servapps/containers/Lidarr","Date":"2023-11-08T04:33:29.779Z","Level":"info","Read":false,"Recipient":"admin","Actions":null},{"ID":"654a589ff54b04d499103b19","Title":"Container Update","Message":"Container transmission updated to the latest version!","Icon":"","Link":"/cosmos-ui/servapps/containers/transmission","Date":"2023-11-07T15:31:56.385Z","Level":"info","Read":true,"Recipient":"admin","Actions":null},{"ID":"6547ff777b87120576934c61","Title":"Container Update","Message":"Container Jellyfin updated to the latest version!","Icon":"","Link":"/cosmos-ui/servapps/containers/Jellyfin","Date":"2023-11-05T20:47:04.658Z","Level":"info","Read":true,"Recipient":"admin","Actions":null},{"ID":"6547ff777b87120576934c60","Title":"Container Update","Message":"Container logs updated to the latest version!","Icon":"","Link":"/cosmos-ui/servapps/containers/logs","Date":"2023-11-05T20:46:56.503Z","Level":"info","Read":true,"Recipient":"admin","Actions":null}],"status":"OK"})
  });
}

function readNotifs(notifs) {
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
  getNotifs,
  readNotifs
};