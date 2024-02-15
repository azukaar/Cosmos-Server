export const simplifyNumber = (num, unit) => {
    if(typeof num == 'undefined' || num === null) return 0;

    num = Math.round(num * 100) / 100;
    

    if(unit.toLowerCase() === "b") {
      if (Math.abs(num) >= 1e12) {
        return parseFloat((num / 1e12).toFixed(1)) + ' TB'; // Convert to Millions
      } else if (Math.abs(num) >= 1e9) {
        return parseFloat((num / 1e9).toFixed(1)) + ' GB'; // Convert to Millions
      } else if (Math.abs(num) >= 1e6) {
          return parseFloat((num / 1e6).toFixed(1)) + ' MB'; // Convert to Millions
      } else if (Math.abs(num) >= 1e3) {
          return parseFloat((num / 1e3).toFixed(1)) + ' KB'; // Convert to Thousands
      } else {
          return num.toString();
      }
    } else if (unit.toLowerCase() === "ms") {
      if (Math.abs(num) >= 1e3) {
          return parseFloat((num / 1e3).toFixed(1)) + ' s'; // Convert to Seconds
      } else {
          return num.toString() + ' ms';
      }
    } else {
      if (Math.abs(num) >= 1e6) {
          return parseFloat((num / 1e6).toFixed(1)) + ' M' + unit; // Convert to Millions
      } else if (Math.abs(num) >= 1e3) {
          return parseFloat((num / 1e3).toFixed(1)) + ' K' + unit; // Convert to Thousands
      } else {
          return num.toString() + ' ' + unit;
      }
    }
}

export const FormaterForMetric = (metric, displayMax) => {
  return (num) => {
    if(!metric) return num;
    
    if(metric.Scale)
      num /= metric.Scale;
    
    num = simplifyNumber(num, metric.Unit);

    if(displayMax && metric.Max) {
      num += ` / ${simplifyNumber(metric.Max, metric.Unit)}`
    }

    return num;
  }
}

export const formatDate = (now, time) => {
  // use as UTC
  // now.setMinutes(now.getMinutes() - now.getTimezoneOffset());

  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0'); // Months are 0-based
  const day = String(now.getDate()).padStart(2, '0');
  const hours = String(now.getHours()).padStart(2, '0');
  const minutes = String(now.getMinutes()).padStart(2, '0');
  const seconds = String(now.getSeconds()).padStart(2, '0');

  return time ? `${year}-${month}-${day} ${hours}:${minutes}:${seconds}` : `${year}-${month}-${day}`;
}

export const toUTC = (date, time) => {
  let now = new Date(date);
  now.setMinutes(now.getMinutes() + now.getTimezoneOffset());
  return formatDate(now, time);
}
