import metricsDemo from './metrics.demo.json';

function get() {
  return new Promise((resolve, reject) => {
    resolve(metricsDemo)
  });
}

function reset() {
  return new Promise((resolve, reject) => {
    resolve()
  });
}

// function list() {
//   return new Promise((resolve, reject) => {
//     resolve()
//   });
// }

export {
  get,
  reset,
  // list,
};