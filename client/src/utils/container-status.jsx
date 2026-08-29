// Container status helpers.
// Health (healthy/starting/unhealthy) wins over the raw run state when a
// healthcheck exists; exited containers show "completed" (clean exit 0) vs
// "exited" (non-zero). Accepts both the summary shape (State string) and the
// inspect shape (State as object with .Status/.Health/.ExitCode).

const HEALTH_STATUSES = ['healthy', 'starting', 'unhealthy'];

function stateFromContainer(container) {
  if (!container) return '';
  if (typeof container.State === 'object' && container.State !== null) { // inspect
    return container.State.Status || '';
  }
  return typeof container.State === 'string' ? container.State : ''; // summary
}

function healthFromContainer(container) {
  if (!container) return '';
  if (typeof container.State === 'object' && container.State !== null && container.State.Health) {
    return container.State.Health.Status || '';
  }
  return (container.Health && container.Health.Status) || container.Health || '';
}

function exitCodeFromContainer(container) {
  if (!container) return null;
  if (typeof container.State === 'object' && container.State !== null) {
    if (typeof container.State.ExitCode === 'number') return container.State.ExitCode;
  }
  if (typeof container.ExitCode === 'number') return container.ExitCode;
  return null;
}

// Display status: healthy/starting/unhealthy (health first), then run state,
// with exited split into exited (non-zero) vs completed (clean exit 0).
export function getContainerDisplayStatus(container) {
  const state = stateFromContainer(container);

  const health = healthFromContainer(container);
  if (health && HEALTH_STATUSES.indexOf(health) !== -1) {
    return health;
  }

  if (state === 'exited') {
    return exitCodeFromContainer(container) === 0 ? 'completed' : 'exited';
  }

  return state;
}

// Stack badge priority: healthy > running > starting > paused > created >
// restarting > removing > completed > exited > dead > unhealthy.
const STATUS_RANK = {
  healthy: 0,
  running: 1,
  starting: 2,
  paused: 3,
  created: 4,
  restarting: 5,
  removing: 6,
  completed: 7,
  exited: 8,
  dead: 9,
  unhealthy: 10,
  '': 100
};

export function rankDisplayStatus(status) {
  return Object.prototype.hasOwnProperty.call(STATUS_RANK, status) ? STATUS_RANK[status] : 100;
}
