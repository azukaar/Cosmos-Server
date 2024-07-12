import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, CircularProgress, Stack } from "@mui/material";
import PrettyTabbedView from "../../components/tabbedView/tabbedView";
import RouteManagement from "./routes/routeman";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import RouteSecurity from "./routes/routeSecurity";
import RouteOverview from "./routes/routeoverview";
import RouteMetrics from "../dashboard/routeMonitoring";
import EventExplorerStandalone from "../dashboard/eventsExplorerStandalone";
import { useTranslation } from 'react-i18next';

const RouteConfigPage = () => {
  const { t } = useTranslation();
  const { routeName } = useParams();
  const [config, setConfig] = useState(null);
  
  let currentRoute = null;
  if (config) {
    currentRoute = config.HTTPConfig.ProxyConfig.Routes.find((r) => r.Name === routeName);
  }

  const refreshConfig = () => {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  };

  useEffect(() => {
    refreshConfig();
  }, []);

  return <div>
    <h2>
      <Stack spacing={1}>
      <Stack direction="row" spacing={1} alignItems="center">
        <Back />
        <div>{routeName}</div>
      </Stack>

      {config && !currentRoute && <div>
        <Alert severity="error">{t('mgmt.servapps.routeConfig.routeNotFound')}</Alert>  
      </div>}

      {config && currentRoute && <PrettyTabbedView tabs={[
        {
          title: t('mgmt.servapps.overview'),
          children: <RouteOverview routeConfig={currentRoute} />
        },
        {
          title: t('mgmt.servapps.routeConfig.setup'),
          children: <RouteManagement
            title={t('mgmt.servapps.routeConfig.setup')}
            submitButton
            routeConfig={currentRoute}
            routeNames={config.HTTPConfig.ProxyConfig.Routes.map((r) => r.Name)}
            config={config}
          />
        },
        {
          title: t('global.securityTitle'),
          children:  <RouteSecurity
            routeConfig={currentRoute}
            config={config}
          />
        },
        {
          title: t('menu-items.navigation.monitoringTitle'),
          children:  <RouteMetrics routeName={routeName} />
        },
        {
          title: t('navigation.monitoring.eventsTitle'),
          children: <EventExplorerStandalone initLevel='info' initSearch={`{"object":"route@${routeName}"}`}/>
        },
      ]}/>}

      {!config && <div style={{textAlign: 'center'}}>
        <CircularProgress />
      </div>}
      </Stack>
    </h2>
  </div>
}

export default RouteConfigPage;