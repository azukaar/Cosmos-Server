import configDemo from './docker.demo.json';

function list() {
  return new Promise((resolve, reject) => {
    resolve(configDemo)
  });
}

function secure(id, res) {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}
    
const newDB = () => {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}
    
const manageContainer = () => {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}
    
export {
  list,
  newDB,
  secure,
  manageContainer
};