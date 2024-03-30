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

export {
  mounts,
  disks,
  snapRAID,
};