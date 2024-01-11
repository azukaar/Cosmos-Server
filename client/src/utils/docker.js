export const smartDockerLogConcat = (log, newLogRaw) => {
  const containsId = (log, id) => {
    const regex = /([a-z0-9]{12})\s([a-zA-Z]+)\s([a-zA-Z0-9\[\]]+)/;
    const match = log.match(regex);
    if (match && match[1] === id) {
      return true;
    }
    return false;  
  }

  try {
    const matches = newLogRaw.match(/([a-z0-9]{12})\s([a-zA-Z]+)\s([a-zA-Z0-9\[\]]+)/);
    const id = matches && matches[1];
    if (id) {
      if (log.find((l) => containsId(l, id))) {
        return log.map((l) => (containsId(l, id) ? newLogRaw : l));
      } else {
        return [...log, newLogRaw];
      }
    }
    return [...log, newLogRaw];
  } catch (e) {
    return [...log, newLogRaw];
  }
}

export const tryParseProgressLog = (log) => {
  return log;
  // try {
  //   const parsedLog = JSON.parse(log);
  //   if (parsedLog.status && parsedLog.progress && parsedLog.id) {
  //     return `${parsedLog.id} ${parsedLog.status} ${parsedLog.progress}`
  //   } else if (parsedLog.status && parsedLog.id && parsedLog.progressDetail && parsedLog.progressDetail.current) {
  //     return `${parsedLog.id} ${parsedLog.status} ${parsedLog.progressDetail.current}/${parsedLog.progressDetail.total}`
  //   } else if (parsedLog.status && parsedLog.id) {
  //     return `${parsedLog.id} ${parsedLog.status} ${parsedLog.sha256 || ""}`
  //   } 
  //   return log;
  // } catch (e) {
  //   return log;
  // }
}