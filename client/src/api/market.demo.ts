import marketDemo from './market.demo.json';

function list() {
  return new Promise((resolve, reject) => {
      resolve(marketDemo);
  });
}

export {
  list,
};