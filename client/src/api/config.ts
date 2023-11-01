import wrap from './wrap';
interface Route {
  Name: string;
}

type Operation = 'replace' | 'move_up' | 'move_down' | 'delete' | 'add';

function get() {
  return wrap(fetch('/cosmos/api/config', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

function set(values) {
  return wrap(fetch('/cosmos/api/config', {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(values),
  }))
}

function restart() {
  return fetch('/cosmos/api/restart', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  })
}

function canSendEmail() {
  return fetch('/cosmos/api/can-send-email', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }).then((response) => {
    return response.json();
  });
}

async function rawUpdateRoute(routeName: string, operation: Operation, newRoute?: Route): Promise<void> {
  const payload = {
    routeName,
    operation,
    newRoute,
  };

  if (operation === 'replace') {
    if (!newRoute) throw new Error('newRoute must be provided for replace operation');
  }

  return wrap(fetch('/cosmos/api/config', {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  }));
}

async function replaceRoute(routeName: string, newRoute: Route): Promise<void> {
  return rawUpdateRoute(routeName, 'replace', newRoute);
}

async function moveRouteUp(routeName: string): Promise<void> {
  return rawUpdateRoute(routeName, 'move_up');
}

async function moveRouteDown(routeName: string): Promise<void> {
  return rawUpdateRoute(routeName, 'move_down');
}

async function deleteRoute(routeName: string): Promise<void> {
  return rawUpdateRoute(routeName, 'delete');
}
async function addRoute(newRoute: Route): Promise<void> {
  return rawUpdateRoute("", 'add', newRoute);
}

function getBackup() {
  return wrap(fetch('/cosmos/api/get-backup', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    },
  }))
}

export {
  get,
  set,
  restart,
  rawUpdateRoute,
  replaceRoute,
  moveRouteUp,
  moveRouteDown,
  deleteRoute,
  addRoute,
  canSendEmail,
  getBackup,
};