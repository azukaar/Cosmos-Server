export const simplifyNumber = (num) => {
    if(!num) return 0;
    
    if (Math.abs(num) >= 1e12) {
      return (num / 1e12).toFixed(1) + 'T'; // Convert to Millions
    } else if (Math.abs(num) >= 1e9) {
      return (num / 1e9).toFixed(1) + 'G'; // Convert to Millions
    } else if (Math.abs(num) >= 1e6) {
        return (num / 1e6).toFixed(1) + 'M'; // Convert to Millions
    } else if (Math.abs(num) >= 1e3) {
        return (num / 1e3).toFixed(1) + 'K'; // Convert to Thousands
    } else {
        return num.toString();
    }
}

export const FormaterForMetric = (metric, displayMax) => {
  return (num) => {
    if(!num) return 0;

    if(metric.Scale)
      num /= metric.Scale;
    
    num = simplifyNumber(num);

    if(displayMax && metric.Max) {
      num += ` / ${simplifyNumber(metric.Max)}`
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
