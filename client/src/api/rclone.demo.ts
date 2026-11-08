// rclone endpoints answer without the usual { status, data } envelope
const demoRemotes = {
  dropbox: { type: 'dropbox' },
  backblaze: { type: 'b2' },
};

function list() {
  return Promise.resolve(demoRemotes);
}

function listRemotes() {
  return Promise.resolve(demoRemotes);
}

function create(data: any) {
  return Promise.resolve({});
}

function update(data: any) {
  return Promise.resolve({});
}

function deleteRemote(remoteName: string) {
  return Promise.resolve({});
}

function pingStorage(remoteName: string) {
  return Promise.resolve({ used: 412000000000, total: 2000000000000 });
}

function stats(remoteName: string) {
  return Promise.resolve({
    diskCache: { uploadsInProgress: 0, uploadsQueued: 0, outOfSpace: false, erroredFiles: 0 },
  });
}

function coreStats() {
  return Promise.resolve({ transferring: [] });
}

function restart() {
  return Promise.resolve({ status: 'OK' });
}

export {
  create,
  list,
  listRemotes,
  deleteRemote,
  update,
  coreStats,
  stats,
  pingStorage,
  restart,
};
