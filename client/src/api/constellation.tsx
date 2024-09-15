import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/constellation/devices', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function ping() {
  return wrap(fetch('/cosmos/api/constellation/ping', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function addDevice(device) {
  return wrap(fetch('/cosmos/api/constellation/devices', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(device),
  }))
}

function resyncDevice(device) {
  return wrap(fetch('/cosmos/api/constellation/config-manual-sync', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify(device),
  }))
}

function restart() {
  return wrap(fetch('/cosmos/api/constellation/restart', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}


function reset() {
  return wrap(fetch('/cosmos/api/constellation/reset', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

function getConfig() {
  return wrap(fetch('/cosmos/api/constellation/config', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

function getLogs() {
  return wrap(fetch('/cosmos/api/constellation/logs', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    }
  }))
}

function connect(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    
    reader.onload = () => {
      fetch('/cosmos/api/constellation/connect', {
        method: 'POST',
        headers: {
          'Content-Type': 'text/plain',
        },
        body: reader.result,
      })
      .then(response => {
        // Add additional response handling here if needed.
        resolve(response);
      })
      .catch(error => {
        // Handle the error.
        reject(error);
      });
    };
    
    reader.onerror = () => {
      reject(new Error('Failed to read the file.'));
    };
    
    reader.readAsText(file);
  });
}

function block(nickname, devicename, block) {
  return wrap(fetch(`/cosmos/api/constellation/block`, {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      nickname, devicename, block
    }),
  }))
}

export {
  list,
  addDevice,
  resyncDevice,
  restart,
  getConfig,
  getLogs,
  reset,
  connect,
  block,
  ping,
};