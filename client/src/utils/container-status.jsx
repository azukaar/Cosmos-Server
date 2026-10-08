// Container status helpers.
//
// We show health first when a container is RUNNING and has a healthcheck
// configured (healthy / starting / unhealthy). Health is meaningless once the
// container is gone: Docker keeps the last health value in its inspect even
// after a stop, so a stopped container must never be reported as "unhealthy".
//
// A lazy container that is not running is "dormant" - this matches upstream
// Cosmos semantics (a lazy container sleeping, stopped, or updated while
// stopped is dormant). The "dormant" state does NOT distinguish who stopped
// the container: it is derived from the cosmos-lazy label and the run state,
// so it survives a Cosmos reboot and a container update without any extra
// bookkeeping. An exited container is shown as "stopped" when it stopped
// cleanly (exit code 0) and as "exited" otherwise.
//
// The servapps list endpoint returns the summary shape (State is a plain
// string, Health/ExitCode/Dormant are extra flat fields), while the container
// detail endpoint returns the inspect shape (State.Status, State.Health.Status,
// State.ExitCode + flat Dormant). getContainerDisplayStatus accepts both.

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

// Whether the container is a lazy container that is not running (upstream
// semantics: dormant = cosmos-lazy label set AND not running). Works for both
// the summary shape (flat Labels map + string State) and the inspect shape
// (Config.Labels + State object). The backend also exposes a flat Dormant
// flag with the same meaning; checking the label + state directly keeps the
// UI correct even when the flag is absent.
export function isContainerDormant(container) {
  if (!container) return false;
  const labels = (container.Labels) || (container.Config && container.Config.Labels) || {};
  if (labels['cosmos-lazy'] !== 'true') return false;
  const state = stateFromContainer(container);
  return state !== 'running';
}

// Returns a display status string:
//   dormant  (lazy container not running - upstream semantics)
//   healthy | starting | unhealthy   (health, only while running)
//   running | paused | created | restarting | removing | dead
//   stopped (clean manual stop, exit code 0) | exited (non-zero exit)
export function getContainerDisplayStatus(container) {
  const state = stateFromContainer(container);

  // A lazy container that is not running is dormant, regardless of its raw
  // state (sleeping, stopped, or updated while stopped).
  if (isContainerDormant(container)) {
    return 'dormant';
  }

  // Health only counts while the container is actually running: a stopped
  // container must never show "unhealthy" (or "healthy"/"starting").
  if (state === 'running') {
    const health = healthFromContainer(container);
    if (health && HEALTH_STATUSES.indexOf(health) !== -1) {
      return health;
    }
    return 'running';
  }

  // Split "exited" into "exited" (failure) vs "stopped" (clean manual stop).
  if (state === 'exited') {
    const exitCode = exitCodeFromContainer(container);
    if (exitCode === 0) {
      return 'stopped';
    }
    return 'exited';
  }

  return state;
}

// Which of two display statuses should win for a stack badge.
// Mirrors the old priority list: running > paused > created > restarting >
// removing > exited > dead. Health statuses sort above plain "running" when
// they are "healthy"-ish, and below when they are not. Dormant is a sleeping
// (reachable) state, so it ranks with the healthy-ish end.
const STATUS_RANK = {
  dormant: 0,
  healthy: 0,
  running: 1,
  starting: 2,
  paused: 3,
  created: 4,
  restarting: 5,
  removing: 6,
  stopped: 7,
  exited: 8,
  dead: 9,
  unhealthy: 10,
  '': 100
};

export function rankDisplayStatus(status) {
  return Object.prototype.hasOwnProperty.call(STATUS_RANK, status) ? STATUS_RANK[status] : 100;
}

// Whether the container is actually running (raw Docker run state). Accepts
// both the summary shape (State is a string) and the inspect shape
// (State.Status). Uses the raw state, not the display status, so a healthy
// container (healthcheck reports "healthy" but state is "running") counts as
// running.
export function isContainerRunning(container) {
  return stateFromContainer(container) === 'running';
}
