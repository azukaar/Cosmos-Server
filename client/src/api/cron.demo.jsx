import wrap from './wrap';

function listen() {
}

function run(scheduler, name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      1000
    );
  });
}

function stop(scheduler, name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      1000
    );
  });
}

function deleteJob(name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      500
    );
  });
}

function list() {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "data": {
          "Custom": {
            "Long Job": {
              "Disabled": false,
              "Scheduler": "Custom",
              "Name": "Long Job",
              "Cancellable": true,
              "Crontab": "0 1/2 * * * *",
              "Running": true,
              "LastStarted": "2024-01-01T00:00:00Z",
              "LastRun": "0001-01-01T00:00:00Z",
              "LastRunSuccess": false,
              "Logs": [
                "2024-01-01T00:00:00Z: Started",
                "2024-01-01T20:00:00Z: Some stuff happening...",
                "2024-01-01T30:00:00Z: Some stuff happening...",
              ],
              "Container": ""
            },
            "Short job": {
              "Disabled": false,
              "Scheduler": "Custom",
              "Name": "Short job",
              "Cancellable": true,
              "Crontab": "0 1 * * * *",
              "Running": false,
              "LastStarted": "2024-01-01T00:00:00Z",
              "LastRun": "2024-01-01T00:01:10Z",
              "LastRunSuccess": true,
              "Logs": [
                "2024-01-01T00:01:10Z: Started",
                "2024-01-01T00:01:10Z: Finished"
              ],
              "Container": ""
            }
          },
          "SnapRAID": {
            "SnapRAID scrub Storage Parity": {
              "Disabled": false,
              "Scheduler": "SnapRAID",
              "Name": "SnapRAID scrub Storage Parity",
              "Cancellable": true,
              "Crontab": "* 0 4 */2 * *",
              "Running": false,
              "LastStarted": "0001-01-01T00:00:00Z",
              "LastRun": "0001-01-01T00:00:00Z",
              "LastRunSuccess": false,
              "Logs": null,
              "Container": ""
            },
            "SnapRAID sync Storage Parity": {
              "Disabled": false,
              "Scheduler": "SnapRAID",
              "Name": "SnapRAID sync Storage Parity",
              "Cancellable": true,
              "Crontab": "* 0 2 * * *",
              "Running": false,
              "LastStarted": "0001-01-01T00:00:00Z",
              "LastRun": "0001-01-01T00:00:00Z",
              "LastRunSuccess": false,
              "Logs": null,
              "Container": ""
            }
          }
        },
        "status": "OK"
      })},
      500
    );
  });
}

function get(scheduler, name) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "data": {
          "Disabled": false,
          "Scheduler": "Custom",
          "Name": "Long Job",
          "Cancellable": true,
          "Crontab": "0 1/2 * * * *",
          "Running": true,
          "LastStarted": "2024-01-01T00:00:00Z",
          "LastRun": "0001-01-01T00:00:00Z",
          "LastRunSuccess": false,
          "Logs": [
            "2024-01-01T00:00:00Z: Started",
            "2024-01-01T20:00:00Z: Some stuff happening...",
            "2024-01-01T30:00:00Z: Some stuff happening...",
          ],
          "Container": ""
        },
        "status": "ok",
      })},
      500
    );
  });
}

export {
  listen,
  list,
  run,
  stop,
  get,
  deleteJob,
}
