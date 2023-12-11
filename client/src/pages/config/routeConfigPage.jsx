import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, CircularProgress, Stack } from "@mui/material";
import PrettyTabbedView from "../../components/tabbedView/tabbedView";
import RouteManagement from "./routes/routeman";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import RouteSecurity from "./routes/routeSecurity";
import RouteOverview from "./routes/routeoverview";
import IsLoggedIn from "../../isLoggedIn";
import RouteMetrics from "../dashboard/routeMonitoring";
import EventExplorerStandalone from "../dashboard/eventsExplorerStandalone";

const RouteConfigPage = () => {
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
        <Alert severity="error">Route not found</Alert>  
      </div>}

      {config && currentRoute && <PrettyTabbedView tabs={[
        {
          title: 'Overview',
          children: <RouteOverview routeConfig={currentRoute} />
        },
        {
          title: 'Setup',
          children: <RouteManagement
            title="Setup"
            submitButton
            routeConfig={currentRoute}
            routeNames={config.HTTPConfig.ProxyConfig.Routes.map((r) => r.Name)}
            config={config}
          />
        },
        {
          title: 'Security',
          children:  <RouteSecurity
            routeConfig={currentRoute}
            config={config}
          />
        },
        {
          title: 'Monitoring',
          children:  <RouteMetrics routeName={routeName} />
        },
        {
          title: 'Events',
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