export const FormaterForMetric = (metric) => {
  console.log(metric)

  return (num) => {
    if(!num) return 0;

    if(metric.Scale)
      num /= metric.Scale;
    
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
}