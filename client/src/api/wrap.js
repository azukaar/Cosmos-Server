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
    throw new Error(rep.message);
  });
}

export function setSnackit(snack) {
  snackit = snack;
}