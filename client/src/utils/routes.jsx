import Folder from '../assets/images/icons/folder(1).svg';

export const sanitizeRoute = (route) => {
  if (!route.UseHost) {
    route.Host = "";
  }
  if (!route.UsePathPrefix) {
    route.PathPrefix = "";
  }
  return route;
}

const addProtocol = (url) => {
  if (url.indexOf("http://") === 0 || url.indexOf("https://") === 0) {
    return url;
  }
  return "https://" + url;
}

export const getOrigin = (route) => {
  return (route.UseHost ? route.Host : '') + (route.UsePathPrefix ? route.PathPrefix : '');
}

export const getFullOrigin = (route) => {
  return addProtocol(getOrigin(route));
}

export const getFaviconURL = (route) => {
  const addRemote = (url) => {
    return '/cosmos/api/favicon?q=' + encodeURIComponent(url)
  }
  if(route.Mode == "SERVAP") {
    return addRemote(addProtocol(route.Target))
  } else if (route.Mode == "STATIC") {
    return Folder;
  } else {
    return addRemote(addProtocol(getOrigin(route)));
  }
}