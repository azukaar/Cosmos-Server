import wrap from './wrap';

function listen() {
  let protocol = 'ws://';
  if (window.location.protocol === 'https:') {
    protocol = 'wss://';
  }
  return new WebSocket(protocol + window.location.host + '/cosmos/api/listen-jobs');
}

function list() {
  return wrap(fetch('/cosmos/api/jobs', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}

function run(scheduler, name) {
  return wrap(fetch('/cosmos/api/jobs/run', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      scheduler: scheduler,
      name: name
    })
  }))
}

function get(scheduler, name) {
  return wrap(fetch('/cosmos/api/jobs/get', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      scheduler: scheduler,
      name: name
    })
  }))
}

function stop(scheduler, name) {
  return wrap(fetch('/cosmos/api/jobs/stop', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      scheduler: scheduler,
      name: name
    })
  }))
}

function deleteJob(name) {
  return wrap(fetch('/cosmos/api/jobs/delete', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      name: name
    })
  }))
}

export {
  listen,
  list,
  run,
  stop,
  get,
  deleteJob,
}
