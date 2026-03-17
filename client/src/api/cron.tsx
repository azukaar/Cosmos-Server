import wrap, { type ApiResponse, type ApiFetch } from './wrap';

export interface CronJob {
  Disabled: boolean;
  Scheduler: string;
  Cancellable: boolean;
  Name: string;
  Crontab: string;
  Running: boolean;
  LastStarted: string;
  LastRun: string;
  LastRunSuccess: boolean;
  Container: string;
  Resource: string;
}

export default function createCronAPI(apiFetch: ApiFetch, createWs: (path: string) => WebSocket) {
  function listen(): WebSocket {
    return createWs('/cosmos/api/listen-jobs');
  }

  function list(): Promise<ApiResponse<Record<string, Record<string, CronJob>>>> {
    return wrap(apiFetch('/cosmos/api/jobs', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  function run(scheduler: string, name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/jobs/run', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        scheduler: scheduler,
        name: name
      })
    }))
  }

  function get(scheduler: string, name: string): Promise<ApiResponse<CronJob>> {
    return wrap(apiFetch('/cosmos/api/jobs/get', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        scheduler: scheduler,
        name: name
      })
    }))
  }

  function stop(scheduler: string, name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/jobs/stop', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        scheduler: scheduler,
        name: name
      })
    }))
  }

  function deleteJob(name: string): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/jobs/delete', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        name: name
      })
    }))
  }

  function runningJobs(): Promise<ApiResponse> {
    return wrap(apiFetch('/cosmos/api/jobs/running', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json'
      }
    }))
  }

  return {
    listen,
    list,
    run,
    stop,
    get,
    deleteJob,
    runningJobs
  };
}
