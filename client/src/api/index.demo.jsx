
export const getStatus = () => {
  return new Promise((resolve, reject) => {
    resolve({
      "data": {
        "AVX": true,
        "CPU": "amd64",
        "HTTPSCertificateMode": "LETSENCRYPT",
        "LetsEncryptErrors": [],
        "MonitoringDisabled": false,
        "backup_status": "",
        "database": true,
        "docker": true,
        "domain": false,
        "homepage": {
          "Background": "/cosmos/api/background/avif",
          "Widgets": null,
          "Expanded": false
        },
        "hostname": "yann-server.com",
        "letsencrypt": false,
        "needsRestart": false,
        "newVersionAvailable": false,
        "resources": {},
        "theme": {
          "PrimaryColor": "rgba(191, 100, 64, 1)",
          "SecondaryColor": ""
        }
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

export const getDNS = (host) => (req) => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
        "data": "199.199.199.199"
      })},
      100
    );
  });
}

export const checkHost = (host) => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
        "data": "199.199.199.199"
      })},
      100
    );
  });
}

export const uploadImage = (file) => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
        "data": ""
      })}, 100 );
    });
  }