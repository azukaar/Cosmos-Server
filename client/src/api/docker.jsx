import wrap from './wrap';
import yaml from 'js-yaml';

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

const manageContainer = (containerId, action) => {
  return wrap(fetch('/cosmos/api/servapps/' + containerId + '/manage/' + action, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

function migrateHost(values) {
  return wrap(fetch('/cosmos/api/migrate-host', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function updateContainer(containerId, values) {
  return wrap(fetch('/cosmos/api/servapps/' + containerId + '/update', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function exportContainer(containerId, values) {
  return wrap(fetch('/cosmos/api/servapps/' + containerId + '/export', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function listContainerNetworks(containerId) {
  return wrap(fetch('/cosmos/api/servapps/' + containerId + '/networks', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

function createNetwork(values) {
  return wrap(fetch('/cosmos/api/networks', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function attachNetwork(containerId, networkId) {
  return wrap(fetch('/cosmos/api/servapps/' + containerId + '/network/' + networkId, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

function detachNetwork(containerId, networkId) {
  return wrap(fetch('/cosmos/api/servapps/' + containerId + '/network/' + networkId, {
    method: 'DELETE',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

function createVolume(values) {
  return wrap(fetch('/cosmos/api/volumes', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function attachTerminal(containerId) {
  let protocol = 'ws://';
  if (window.location.protocol === 'https:') {
    protocol = 'wss://';
  }
  return new WebSocket(protocol + window.location.host + '/cosmos/api/servapps/' + containerId + '/terminal/attach');
}

function createTerminal(containerId) {
  let protocol = 'ws://';
  if (window.location.protocol === 'https:') {
    protocol = 'wss://';
  }
  return new WebSocket(protocol + window.location.host + '/cosmos/api/servapps/' + containerId + '/terminal/new');
}

function createService(serviceData, onProgress) {
  const requestOptions = {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(serviceData)
  };

  return fetch('/cosmos/api/docker-service', requestOptions)
    .then(response => {
      if (!response.ok) {
        throw new Error(response.statusText);
      }

      // The response body is a ReadableStream. This code reads the stream and passes chunks to the callback.
      const reader = response.body.getReader();

      // Read the stream and pass chunks to the callback as they arrive
      return new ReadableStream({
        start(controller) {
          function read() {
            return reader.read().then(({ done, value }) => {
              if (done) {
                controller.close();
                return;
              }
              // Decode the UTF-8 text
              let text = new TextDecoder().decode(value);
              // Split by lines in case there are multiple lines in one chunk
              let lines = text.split('\n');
              for (let line of lines) {
                if (line) {
                  // Call the progress callback
                  onProgress(line);
                }
              }
              controller.enqueue(value);
              return read();
            });
          }
          return read();
        }
      });
    });
}

function pullImage(imageName, onProgress, ifMissing) {
  const requestOptions = {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  };

  const imageNameEncoded = encodeURIComponent(imageName);

  return fetch(`/cosmos/api/images/${ifMissing ? 'pull-if-missing' : 'pull'}?imageName=${imageNameEncoded}`, requestOptions)
    .then(response => {
      if (!response.ok) {
        throw new Error(response.statusText);
      }

      // The response body is a ReadableStream. This code reads the stream and passes chunks to the callback.
      const reader = response.body.getReader();

      // Read the stream and pass chunks to the callback as they arrive
      return new ReadableStream({
        start(controller) {
          function read() {
            return reader.read().then(({ done, value }) => {
              if (done) {
                controller.close();
                return;
              }
              // Decode the UTF-8 text
              let text = new TextDecoder().decode(value);
              // Split by lines in case there are multiple lines in one chunk
              let lines = text.split('\n');
              for (let line of lines) {
                if (line) {
                  // Call the progress callback
                  onProgress(line);
                }
              }
              controller.enqueue(value);
              return read();
            });
          }
          return read();
        }
      });
    });
}

function updateContainerImage(containerName, onProgress) {
  const requestOptions = {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  };

  const containerNameEncoded = encodeURIComponent(containerName);

  return fetch(`/cosmos/api/servapps/${containerNameEncoded}/manage/update`, requestOptions)
    .then(response => {
      if (!response.ok) {
        throw new Error(response.statusText);
      }

      // The response body is a ReadableStream. This code reads the stream and passes chunks to the callback.
      const reader = response.body.getReader();

      // Read the stream and pass chunks to the callback as they arrive
      return new ReadableStream({
        start(controller) {
          function read() {
            return reader.read().then(({ done, value }) => {
              if (done) {
                controller.close();
                return;
              }
              // Decode the UTF-8 text
              let text = new TextDecoder().decode(value);
              // Split by lines in case there are multiple lines in one chunk
              let lines = text.split('\n');
              for (let line of lines) {
                if (line) {
                  // Call the progress callback
                  onProgress(line);
                }
              }
              controller.enqueue(value);
              return read();
            });
          }
          return read();
        }
      });
    });
}

function autoUpdate(id, toggle) {
  return wrap(fetch('/cosmos/api/servapps/' + id + '/auto-update/'+toggle, {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
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
  exportContainer,
  migrateHost,
};