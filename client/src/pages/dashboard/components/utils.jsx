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
  console.log(metric)

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