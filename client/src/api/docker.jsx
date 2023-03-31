import wrap from './wrap';

function list() {
  return wrap(fetch('/cosmos/api/servapps', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}

function secure(id, res) {
  return wrap(fetch('/cosmos/api/servapps/' + id + '/secure/'+res, {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json'
    },
  }))
}
    
const newDB = () => {
  return wrap(fetch('/cosmos/api/newDB', {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json'
    }
  }))
}
    
export {
  list,
  newDB,
  secure
};