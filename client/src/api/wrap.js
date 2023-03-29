let snackit;

export default function wrap(apicall) {
  return apicall.then(async (response) => {
    let rep;
    try {
      rep = await response.json();
    } catch {
      snackit('Server error');
      throw new Error('Server error');
    }
    if (response.status == 200) {
      return rep;
    } 
    snackit(rep.message);
    const e = new Error(rep.message);
    e.status = response.status;
    throw e;
  });
}

export function setSnackit(snack) {
  snackit = snack;
}