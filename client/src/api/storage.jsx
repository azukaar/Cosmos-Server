import wrap from './wrap';

const mounts = {
  list: () => {
    return wrap(fetch('/cosmos/api/mounts', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },

  mount: ({path, mountPoint, permanent, chown}) => {
    return wrap(fetch('/cosmos/api/mount', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify({path, mountPoint, permanent, chown})
    }))
  },

  unmount: ({mountPoint, permanent, chown}) => {
    return wrap(fetch('/cosmos/api/unmount', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify({mountPoint, permanent, chown})
    }))
  },

  merge: (args) => {
    return wrap(fetch('/cosmos/api/merge', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(args)
    }))
  },
};
const raid = {
  list: () => {
    return wrap(fetch('/cosmos/api/storage/raid', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      },
    }))
  },

  create: ({name, level, devices, spares, metadata}) => {
    return wrap(fetch('/cosmos/api/storage/raid', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({name, level, devices, spares, metadata})
    }))
  },

  delete: (name) => {
    return wrap(fetch(`/cosmos/api/storage/raid/${name}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  },

  status: (name) => {
    return wrap(fetch(`/cosmos/api/storage/raid/${name}/status`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  },

  addDevice: (name, device) => {
    return wrap(fetch(`/cosmos/api/storage/raid/${name}/device`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({device})
    }))
  },

  replaceDevice: (name, oldDevice, newDevice) => {
    return wrap(fetch(`/cosmos/api/storage/raid/${name}/replace`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({oldDevice, newDevice})
    }))
  },

  resize: (name) => {
    return wrap(fetch(`/cosmos/api/storage/raid/${name}/resize`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }
};

const disks = {
  list: () => {
    return wrap(fetch('/cosmos/api/disks', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },

  smartDef: () => {
    return wrap(fetch('/cosmos/api/smart-def', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },

  format({disk, format, password}, onProgress) {
    const requestOptions = {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify({disk, format, password})
    };
  
    return fetch('/cosmos/api/disks/format', requestOptions)
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
};

const snapRAID = {
  create: (args) => {
    return wrap(fetch('/cosmos/api/snapraid', {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(args)
    }))
  },
  update: (name, args) => {
    return wrap(fetch('/cosmos/api/snapraid/' + name, {
      method: 'POST',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(args)
    }))
  },
  delete: (name) => {
    return wrap(fetch('/cosmos/api/snapraid/' + name, {
      method: 'DELETE',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },
  list: (args) => {
    return wrap(fetch('/cosmos/api/snapraid', {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
      body: JSON.stringify(args)
    }))
  },
  sync: (name) => {
    return wrap(fetch(`/cosmos/api/snapraid/${name}/sync`, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },
  fix: (name) => {
    return wrap(fetch(`/cosmos/api/snapraid/${name}/fix`, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },
  enable: (name, enable) => {
    return wrap(fetch(`/cosmos/api/snapraid/${name}/` + (enable ? 'enable' : 'disable'), {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },
  scrub: (name) => {
    return wrap(fetch(`/cosmos/api/snapraid/${name}/scrub`, {
      method: 'GET',
      headers: {
          'Content-Type': 'application/json'
      },
    }))
  },
};


const listDir = (storage, path) => {
  return wrap(fetch(`/cosmos/api/list-dir?storage=${storage}&path=${path}`, {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

const newFolder = (storage, path, folder) => {
  return wrap(fetch(`/cosmos/api/new-dir?storage=${storage}&path=${path}&folder=${folder}`, {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

export {
  mounts,
  disks,
  snapRAID,
  raid,
  newFolder,
  listDir,
};