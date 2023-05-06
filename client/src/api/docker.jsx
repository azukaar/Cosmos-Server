import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/servapps', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function get(containerName) {
  return wrap(fetch('/cosmos/api/servapps/' + containerName, {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function getContainerLogs(containerId, searchQuery, limit, lastReceivedLogs, errorOnly) {

  if(limit < 50) limit = 50;

  const queryParams = new URLSearchParams({
    search: searchQuery || "",
    limit: limit || "",
    lastReceivedLogs: lastReceivedLogs || "",
    errorOnly: errorOnly || "",
  });

  return wrap(fetch(`/cosmos/api/servapps/${containerId}/logs?${queryParams}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }));
}

function volumeList() {
  return wrap(fetch('/cosmos/api/volumes', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function volumeDelete(name) {
  return wrap(fetch(`/cosmos/api/volume/${name}`, {
    method: 'DELETE',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function networkList() {
  return wrap(fetch('/cosmos/api/networks', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function networkDelete(name) {
  return wrap(fetch(`/cosmos/api/network/${name}`, {
    method: 'DELETE',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function secure(id, res) {
  return wrap(fetch('/cosmos/api/servapps/' + id + '/secure/'+res, {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}
    
const newDB = () => {
  return wrap(fetch('/cosmos/api/newDB', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

const manageContainer = (id, action) => {
  return wrap(fetch('/cosmos/api/servapps/' + id + '/manage/' + action, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}
    
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
  getContainerLogs
};