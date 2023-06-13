export const randomString = (length) => {
  let text = "";
  const possible =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
  for (let i = 0; i < length; i++)
    text += possible.charAt(Math.floor(Math.random() * possible.length));
  return text;
}

export function isDomain(hostname) {
  // Regular expression to check if it's an IP address
  const ipPattern = /^(?:[0-9]{1,3}\.){3}[0-9]{1,3}$/;

  // Remove port if there is one
  hostname = hostname.replace(/:\d+$/, '');

  // Check if the hostname is an IP address
  if (ipPattern.test(hostname)) {
      return false;
  }

  // Check if the hostname is "localhost"
  if (hostname === 'localhost') {
      return false;
  }

  return true;
}

export function debounce(func, wait, immediate) {
  var timeout;

  return () => {
      var context = this, args = arguments;

      var later = () => {
          timeout = null;
          if (!immediate) func.apply(context, args);
      };

      var callNow = immediate && !timeout;

      clearTimeout(timeout);

      timeout = setTimeout(later, wait);

      if (callNow) func.apply(context, args);
  };
};