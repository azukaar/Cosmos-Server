// Container status helpers.
//
// We show health first when a container has a healthcheck configured
// (healthy / starting / unhealthy), and only fall back to the raw Docker
// run state otherwise. An exited container is shown as "completed" when it
// stopped cleanly (exit code 0) and as "exited" otherwise.
//
// The servapps list endpoint returns the summary shape (State is a plain
// string, Health/ExitCode are extra flat fields), while the container detail
// endpoint returns the inspect shape (State.Status, State.Health.Status,
// State.ExitCode). getContainerDisplayStatus accepts both.

const HEALTH_STATUSES = ['healthy', 'starting', 'unhealthy'];

function stateFromContainer(container) {
  if (!container) return '';
  // inspect shape: container.State is an object
  if (typeof container.State === 'object' && container.State !== null) {
    return container.State.Status || '';
  }
  // summary shape: container.State is a string
  return typeof container.State === 'string' ? container.State : '';
}

function healthFromContainer(container) {
  if (!container) return '';
  // inspect shape
  if (typeof container.State === 'object' && container.State !== null && container.State.Health) {
    return container.State.Health.Status || '';
  }
  // summary shape: flat Health field
  return (container.Health && container.Health.Status) || container.Health || '';
}

function exitCodeFromContainer(container) {
  if (!container) return null;
  // inspect shape
  if (typeof container.State === 'object' && container.State !== null) {
    if (typeof container.State.ExitCode === 'number') return container.State.ExitCode;
  }
  // summary shape: flat ExitCode field
  if (typeof container.ExitCode === 'number') return container.ExitCode;
  return null;
}

// Returns a display status string:
//   healthy | starting | unhealthy   (health takes priority when present)
//   running | paused | created | restarting | removing | dead
//   exited  (non-zero exit) | completed (clean exit, code 0)
export function getContainerDisplayStatus(container) {
  const state = stateFromContainer(container);

  // Health takes priority over the run state when a healthcheck exists.
  const health = healthFromContainer(container);
  if (health && HEALTH_STATUSES.indexOf(health) !== -1) {
    return health;
  }

  // Split "exited" into "exited" (failure) vs "completed" (clean stop).
  if (state === 'exited') {
    const exitCode = exitCodeFromContainer(container);
    if (exitCode === 0) {
      return 'completed';
    }
    return 'exited';
  }

  return state;
}

// Which of two display statuses should win for a stack badge.
// Mirrors the old priority list: running > paused > created > restarting >
// removing > exited > dead. Health statuses sort above plain "running" when
// they are "healthy"-ish, and below when they are not.
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