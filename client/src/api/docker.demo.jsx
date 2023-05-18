import configDemo from './docker.demo.json';
import configDemoCont from './container.demo.json';
import volumesDemo from './volumes.demo.json';
import networkDemo from './networks.demo.json';
import logDemo from './logs.demo.json';

function list() {
  return new Promise((resolve, reject) => {
    resolve(configDemo)
  });
}

function get(containerName) {
  return new Promise((resolve, reject) => {
    resolve(configDemoCont)
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
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function updateContainer(containerId, values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function listContainerNetworks(containerId) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function createNetwork(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function attachNetwork(containerId, networkId) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function detachNetwork(containerId, networkId) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function createVolume(values) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function volumeList() {
  return new Promise((resolve, reject) => {
    resolve(volumesDemo)
  });
}

function volumeDelete(name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function networkList() {
  return new Promise((resolve, reject) => {
    resolve(networkDemo)
  });
}

function networkDelete(name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

function getContainerLogs(containerId, searchQuery, limit, lastReceivedLogs, errorOnly) {
  if(limit < 50) limit = 50;

  return new Promise((resolve, reject) => {
    resolve(logDemo)
  });
}

function attachTerminal(name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      100
    );
  });
}

function createTerminal(name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      100
    );
  });
}

function autoUpdate(name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      1000
    );
  });
}

function pullImage(imageName, onProgress, ifMissing) {
  onProgress('Updating...')
    new Promise((resolve, reject) => {
      setTimeout(() => {
        onProgress('Download in progress...')
      },
        1000
      );
    });
  return new Promise((resolve, reject) => {
      setTimeout(() => {
        onProgress('[OPERATION SUCCEEDED]')
      },
      2500
    );
  });
}
const updateContainerImage = pullImage;
const createService = pullImage;

export {
  list,
  get,
  newDB,
  secure,
  manageContainer,
  volumeList,
  volumeDelete,
  networkList,
  networkDelete,
  getContainerLogs,
  updateContainer,
  listContainerNetworks,
  createNetwork,
  attachNetwork,
  detachNetwork,
  createVolume,
  attachTerminal,
  createTerminal,
  createService,
  pullImage,
  autoUpdate,
  updateContainerImage,
};