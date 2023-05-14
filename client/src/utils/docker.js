export const smartDockerLogConcat = (log, newLogRaw) => {
  const containsId = (log, id) => {
    try {
      const parsedLog = JSON.parse(log);
      return parsedLog.id === id;
    }
    catch (e) {
      return false;
    }
 }

  try {
    const newLog = JSON.parse(newLogRaw);
    if (newLog.id) {
      if (log.find((l) => containsId(l, newLog.id))) {
        return log.map((l) => (containsId(l, newLog.id) ? newLogRaw : l));
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
  try {
    const parsedLog = JSON.parse(log);
    if (parsedLog.status && parsedLog.progress && parsedLog.id) {
      return `${parsedLog.id} ${parsedLog.status} ${parsedLog.progress}`
    } else if (parsedLog.status && parsedLog.id && parsedLog.progressDetail && parsedLog.progressDetail.current) {
      return `${parsedLog.id} ${parsedLog.status} ${parsedLog.progressDetail.current}/${parsedLog.progressDetail.total}`
    } else if (parsedLog.status && parsedLog.id) {
      return `${parsedLog.id} ${parsedLog.status} ${parsedLog.sha256 || ""}`
    } 
    return log;
  } catch (e) {
    return log;
  }
}
