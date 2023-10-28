import metricsDemo from './metrics.demo.json';

function get() {
  return new Promise((resolve, reject) => {
    resolve(metricsDemo)
  });
}

export {
  get,
};