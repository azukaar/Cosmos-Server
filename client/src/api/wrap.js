let snackit;

export default function wrap(apicall, noError = false) {
  return apicall.then(async (response) => {
    let rep;
    try {
      rep = await response.json();
    } catch (err) {
      if (!noError) {
        snackit('Server error');
        throw new Error('Server error');
      } else {
        const e = new Error(rep.message);
        e.status = rep.status;
        e.code = rep.code;
        throw e;
      }
    }

    if (response.status == 200) {
      return rep;
    } 

    if (!noError) {
      snackit(rep.message);
    }
    
    const e = new Error(rep.message);
    e.status = rep.status;
    e.code = rep.code;
    throw e;
  });
}

export function setSnackit(snack) {
  snackit = snack;
}

export {
  snackit
};