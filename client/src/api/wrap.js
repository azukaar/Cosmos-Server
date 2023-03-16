let snackit;

export default function wrap(apicall) {
  return apicall.then(async (response) => {
    const rep = await response.json();
    if (response.status == 200) {
      return rep;
    } 
    snackit(rep.message);
    throw new Error(rep.message);
  });
}

export function setSnackit(snack) {
  snackit = snack;
}