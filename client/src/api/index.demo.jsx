
export const getStatus = () => {
  return new Promise((resolve, reject) => {
    resolve({
      "data": {
        "HTTPSCertificateMode": "LETSENCRYPT",
        "database": true,
        "docker": true,
        "domain": false,
        "letsencrypt": false,
        "needsRestart": false,
        "newVersionAvailable": false
      },
      "status": "OK"
    });
  });
}

export const isOnline = () => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}

export const newInstall = (req) => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
      })},
      2000
    );
  });
}