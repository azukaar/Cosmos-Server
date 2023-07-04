
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
        "newVersionAvailable": false,
        "homepage": {
          "Background": "",
          "Widgets": null,
          "Expanded": false
        },
        "theme": {
          "PrimaryColor": "",
          "SecondaryColor": ""
        },
        "LetsEncryptErrors": [],
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

export const uploadBackground = (file) => {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve({
        "status": "ok",
        "data": ""
      })}, 100 );
    });
  }