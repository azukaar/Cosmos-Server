import configDemo from './demo.config.json';

interface Route {
  Name: string;
}

type Operation = 'replace' | 'move_up' | 'move_down' | 'delete' | 'add';

function get() {
  return new Promise((resolve, reject) => {
    resolve(configDemo)
  });
}

function set(values) {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function restart() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
    })
  });
}

function canSendEmail() {
  return new Promise((resolve, reject) => {
    resolve({
      "status": "ok",
      "data": {
        "canSendEmail": true,
      }
    })
  });
}

async function rawUpdateRoute(routeName: string, operation: Operation, newRoute?: Route): Promise<void> {
  return new Promise((resolve, reject) => {
    resolve()
  });
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
};