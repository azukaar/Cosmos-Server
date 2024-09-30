import Folder from '../assets/images/icons/folder(1).svg';
import demoicons from './icons.demo.json';
import logogray from '../assets/images/icons/cosmos_gray.png';

import * as Yup from 'yup';
import { Alert, Grid } from '@mui/material';
import { debounce, isDomain } from './indexs';

import * as API from '../api';
import { useEffect, useState } from 'react';
import { Trans } from 'react-i18next';

export const sanitizeRoute = (_route) => {
  let route = { ..._route };

  if (!route.UseHost) {
    route.Host = "";
  }
  if (!route.UsePathPrefix) {
    route.PathPrefix = "";
  }

  route.Name = route.Name.trim();

  if (!route.SmartShield) {
    route.SmartShield = {};
  }

  if (typeof route._SmartShield_Enabled !== "undefined") {
    route.SmartShield.Enabled = route._SmartShield_Enabled;
    delete route._SmartShield_Enabled;
  }

  return route;
}

const addProtocol = (url) => {
  if (url.indexOf("http://") === 0 || url.indexOf("https://") === 0) {
    return url;
  }
  // get current protocol
  if (window.location.protocol === "http:") {
    return "http://" + url;
  }
  return "https://" + url;
}

export const getOrigin = (route) => {
  return (route.UseHost ? route.Host : window.location.origin) + (route.UsePathPrefix ? route.PathPrefix : '');
}

export const getFullOrigin = (route) => {
  return addProtocol(getOrigin(route));
}

const isDemo = import.meta.env.MODE === 'demo';

export const getFaviconURL = (route) => {
  if (isDemo) {
    if (route.Mode == "STATIC")
      return Folder;
    return demoicons[route.Name] || logogray;
  }

  if (!route) {
    return logogray;
  }

  const addRemote = (url, servapp) => {
    return '/cosmos/api/favicon?q=' + encodeURIComponent(url) + (servapp ? '&servapp=true' : '');
  }

  if (route.Mode == "SERVAPP" || route.Mode == "PROXY") {
    return addRemote(route.Target, route.Mode == "SERVAPP")
  } else if (route.Mode == "STATIC") {
    return Folder;
  } else {
    return addRemote(addProtocol(getOrigin(route)));
  }
}

export const ValidateRouteSchema = (t) => {
  return Yup.object().shape({
    Name: Yup.string().required(t('global.name.validation')),
    Mode: Yup.string().required(t('global.mode.validation')),
    Target: Yup.string().required(t('global.target.validation')).when('Mode', {
      is: 'SERVAPP',
      then: Yup.string().matches(/:[0-9]+$/, t('mgmt.config.containerPicker.targetTypeValidation.noPort')),
    }).when('Mode', {
      is: 'PROXY',
      then: Yup.string().matches(/^(https?:\/\/)/, t('mgmt.config.containerPicker.targetTypeValidation.wrongProtocol') ),
    }),

    Host: Yup.string().when('UseHost', {
      is: true,
      then: Yup.string()
        .required(t('mgmt.urls.edit.hostInput.HostRequired'))
        .matches(/[\.|\:]/, t('mgmt.urls.edit.hostInput.HostValidation'))
        .test('is-protocol', t('mgmt.urls.edit.hostInput.HostValidation.caseIsProtocol'), (value) => {
          return !value.match(/\:.*?[a-zA-Z]+/);
        })
    }),

    PathPrefix: Yup.string().when('UsePathPrefix', {
      is: true,
      then: Yup.string().required(t('mgmt.urls.edit.pathPrefixInput.pathPrefixRequired')).matches(/^\//, t('mgmt.urls.edit.pathPrefixInput.pathPrefixValidation'))
    }),

    UseHost: Yup.boolean().when('UsePathPrefix',
      {
        is: false,
        then: Yup.boolean().oneOf([true], t('mgmt.urls.edit.pathPrefixInput.pathPrefixSource'))
      }),
  })
};

export const ValidateRoute = (routeConfig, config, t) => {
  let routeNames = config.HTTPConfig.ProxyConfig.Routes.map((r) => r.Name);

  try {
    ValidateRouteSchema(t).validateSync(routeConfig);
  } catch (e) {
    return e.errors;
  }
  if (routeNames.includes(routeConfig.Name)) {
    return [t('mgmt.urls.edit.nameValidation')];
  }
  return [];
}

export const getContainersRoutes = (config, containerName) => {
  return (config && config.HTTPConfig && config.HTTPConfig.ProxyConfig.Routes && config.HTTPConfig.ProxyConfig.Routes.filter((route) => {
    let reg = new RegExp(`^(([a-z]+):\/\/)?${containerName}(:?[0-9]+)?$`, 'i');
    return route.Mode == "SERVAPP" && reg.test(route.Target)
  })) || [];
}

export const getContaienrsJobs = (config, containerName) => {
  return (config && config.CRON && Object.values(config.CRON).filter((job) => {
    return job.Container == containerName || job.Container == containerName.replace(/^\/+/, '');
  })) || [];
}

const checkHost = debounce((host, setHostError, setHostIp) => {
  if (isDomain(host)) {
    API.getDNS(host).then((data) => {
      setHostError(null)
      setHostIp(data.data)
    }).catch((err) => {
      setHostError(err.message)
      setHostIp(null)
    });
  } else {
    setHostError(null);
    setHostIp(null);
  }
}, 500)

export const HostnameChecker = ({hostname}) => {
  const [hostError, setHostError] = useState(null);
  const [hostIp, setHostIp] = useState(null);

  useEffect(() => {
    if (hostname) {
      checkHost(hostname, setHostError, setHostIp);
    }
  }, [hostname]);

  return <>{hostError && <Alert color='error'>{hostError}</Alert>}

    {hostIp && <Alert color='info'><Trans i18nKey="newInstall.hostnamePointsToInfo" values={{hostIp: hostIp}} /></Alert>}
  </>
};

const hostnameIsDomainReg = /^((?!localhost|\d+\.\d+\.\d+\.\d+)[a-zA-Z0-9\-]{1,63}\.)+[a-zA-Z]{2,63}$/

export const getHostnameFromName = (name, route, config, overrideOrigin) => {
  let origin = overrideOrigin || window.location.origin.split('://')[1];
  let protocol = overrideOrigin || window.location.origin.split('://')[0];
  let port = origin.split(':')[1];
  if (port)
    origin = origin.split(':')[0];
  let res;

  if(origin.match(hostnameIsDomainReg)) {
    res = name.replace('/', '').replace(/_/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase().replace(/\s/g, '-');
    
    if(route)
      res = (route.hostPrefix || '') + res + (route.hostSuffix || '');
    
    if(origin.endsWith('.local'))
      res = res + '.local'
    else
      res = res + '.' + origin

    if(port)
      res += ':' + port;
  } else {
    const existingRoutes = Object.values(config.HTTPConfig.ProxyConfig.Routes);
    origin = origin.split(':')[0];

    // find first port available in range 7200-7400
    let port = protocol == "https" ? 7200 : 7351;
    let endPort = protocol == "https" ? 7350 : 7500;
    while(port < endPort) {
      if(!existingRoutes.find((exiroute) => exiroute.Host == (origin + ":" + port))) {
        res = origin + ":" + port;
        return res;
      }
      else
        port++;
      }
      
    return "NO MORE PORT AVAILABLE. PLEASE CLEAN YOUR URLS!";
  }
  return res;
}