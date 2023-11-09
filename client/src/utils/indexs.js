import { Button } from "@mui/material";

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

export const debounce = (func, wait) => {
    let timeout;
    return function (...args) {
      const context = this;
      clearTimeout(timeout);
      timeout = setTimeout(() => func.apply(context, args), wait);
    };
  };

export const redirectTo = (url) => {
  window.location.href = url;
}

export const redirectToLocal = (url) => {
  let redirectUrl = new URL(url, window.location.href);
  let currentLocation = window.location;
  if (redirectUrl.origin != currentLocation.origin){
    throw new Error("URL must be local");
  }
  window.location.href = url;
}
