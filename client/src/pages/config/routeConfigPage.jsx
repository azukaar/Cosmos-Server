import { useParams } from "react-router";
import Back from "../../components/back";
import { Alert, CircularProgress, Stack } from "@mui/material";
import PrettyTabbedView from "../../components/tabbedView/tabbedView";
import RouteManagement from "./routes/routeman";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import RouteSecurity from "./routes/routeSecurity";
import RouteOverview from "./routes/routeoverview";

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
          />
        },
        {
          title: 'Security',
          children:  <RouteSecurity
            routeConfig={currentRoute}
          />
        },
        {
          title: 'Permissions',
          children: <div>WIP</div>
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