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


export {
  listen,
  list
}
