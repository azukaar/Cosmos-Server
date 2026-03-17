export interface ApiResponse<T = any> {
  status: string;
  data: T;
  message?: string;
  code?: string;
}

export type ApiFetch = (path: string, options?: RequestInit) => Promise<Response>;

interface CosmosError extends Error {
  status?: number;
  code?: string | number;
}

let snackit: ((msg: string) => void) | undefined;

export function wrapRClone<T = any>(apicall: Promise<Response>): Promise<T> {
  return apicall.then(async (response) => {
    let rep = await response.json();
    if (response.status >= 400) {
      if (snackit) snackit(rep.error);
      const e: CosmosError = new Error(rep.error);
      e.status = response.status;
      e.code = response.status;
      throw e;
    }
    return rep;
  });
}

export default function wrap<T = any>(apicall: Promise<Response>, noError = false): Promise<ApiResponse<T>> {
  return apicall.then(async (response) => {
    let rep: any;
    try {
      rep = await response.text();

      try {
        rep = JSON.parse(rep);
      } catch (err) {
        rep = {
          message: rep,
          status: response.status,
          code: response.status
        };
      }
    } catch (err) {
      if (!noError) {
        if (snackit) snackit('Server error');
        throw new Error('Server error');
      } else {
        const e: CosmosError = new Error(rep.message);
        e.status = rep.status;
        e.code = rep.code;
        throw e;
      }
    }

    if (response.status == 200) {
      return rep;
    }

    if (!noError && rep.message) {
      if (snackit) snackit(rep.message);
    }

    const e: CosmosError = new Error(rep.message);
    e.status = rep.status;
    e.code = rep.code;
    throw e;
  });
}

export function setSnackit(snack: (msg: string) => void) {
  snackit = snack;
}

export {
  snackit
};
