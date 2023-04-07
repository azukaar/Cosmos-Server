// material-ui
import { AppstoreAddOutlined, ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Checkbox, Chip, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, Input, InputAdornment, TextField, Tooltip, Typography } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2/Grid2';
import { Stack } from '@mui/system';
import { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';

import * as API from '../../api';
import isLoggedIn from '../../isLoggedIn';
import RestartModal from '../config/users/restart';
import RouteManagement, { ValidateRoute } from '../config/users/routeman';

const Item = styled(Paper)(({ theme }) => ({
  backgroundColor: theme.palette.mode === 'dark' ? '#1A2027' : '#fff',
  ...theme.typography.body2,
  padding: theme.spacing(1),
  textAlign: 'center',
  color: theme.palette.text.secondary,
}));

const ServeApps = () => {
  isLoggedIn();

  const [serveApps, setServeApps] = useState([]);
  const [isUpdating, setIsUpdating] = useState({});
  const [search, setSearch] = useState("");
  const [config, setConfig] = useState(null);
  const [openModal, setOpenModal] = useState(false);
  const [newRoute, setNewRoute] = useState(null);
  const [submitErrors, setSubmitErrors] = useState([]);
  const [openRestartModal, setOpenRestartModal] = useState(false);

  const hasCosmosNetwork = (containerName) => {
    const container = serveApps.find((app) => {
      return app.Names[0].replace('/', '') === containerName.replace('/', '');
    });
    return container && container.NetworkSettings.Networks && Object.keys(container.NetworkSettings.Networks).some((network) => {
      if(network.startsWith('cosmos-network'))
        return true;
    })
  }

  const testRoute = (route) => {
    console.log(newRoute)
    try {
      ValidateRoute.validateSync(route);
    } catch (e) {
      return e.errors;
    }
  }

  const refreshServeApps = () => {
    API.docker.list().then((res) => {
      setServeApps(res.data);
    });
    API.config.get().then((res) => {
      setConfig(res.data);
    });
    setIsUpdating({});
  };

  const setIsUpdatingId = (id, value) => {
    setIsUpdating({
      ...isUpdating,
      [id]: value
    });
  }

  const getContainersRoutes = (containerName) => {
    return (config && config.HTTPConfig && config.HTTPConfig.ProxyConfig.Routes.filter((route) => {
      return route.Mode == "SERVAPP" && (
        route.Target.startsWith(containerName) ||
        route.Target.split('://')[1].startsWith(containerName)
      )
    })) || [];
  }

  useEffect(() => {
    refreshServeApps();
  }, []);
  
  function updateRoutes() {
    let con = {
      ...config,
      HTTPConfig: {
        ...config.HTTPConfig,
        ProxyConfig: {
          ...config.HTTPConfig.ProxyConfig,
          Routes: [
            ...config.HTTPConfig.ProxyConfig.Routes,
            newRoute,
          ]
        },
      },
    };
    
    API.config.set(con).then((res) => {
      setOpenModal(false);
      setOpenRestartModal(true);
    });
  }
  const gridAnim = {
    transition: 'all 0.2s ease',
    opacity: 1,
    transform: 'translateY(0px)',
    '&.MuiGrid2-item--hidden': {
      opacity: 0,
      transform: 'translateY(-20px)',
    },
  };

  return <div>
    <RestartModal openModal={openRestartModal} setOpenModal={setOpenRestartModal} />
    <Dialog open={openModal} onClose={() => setOpenModal(false)}>
        <DialogTitle>Connect ServApp</DialogTitle>
            {openModal && <>
            <DialogContent>
                <DialogContentText>
                  <Stack spacing={2}>
                    <div>
                      Welcome to the Connect Wizard. This interface will help you expose your ServApp securely to the internet.
                    </div>
                    <div>
                        {openModal && !hasCosmosNetwork(openModal.Names[0]) && <Alert severity="warning">This ServApp does not appear to be connected to a Cosmos Network, so the hostname might not be accessible. The easiest way to fix this is to check the box "Force Secure Network" or manually create a sub-network in Docker.</Alert>}
                    </div>
                    <div>
                        <RouteManagement TargetContainer={openModal} 
                          routeConfig={{
                            Target: "http://"+openModal.Names[0].replace('/', '') + ":",
                            Mode: "SERVAPP",
                            Name: openModal.Names[0].replace('/', ''),
                            Description: "Expose " + openModal.Names[0].replace('/', '') + " to the internet",
                            UseHost: false,
                            Host: '',
                            UsePathPrefix: false,
                            PathPrefix: '',
                            Timeout: 30000,
                            ThrottlePerMinute: 0,
                            CORSOrigin: '',
                            StripPathPrefix: false,
                            AuthEnabled: false,
                          }} 
                          setRouteConfig={(_newRoute) => {
                            setNewRoute(_newRoute);
                          }}
                          up={() => {}}
                          down={() => {}}
                          deleteRoute={() => {}}
                          noControls
                          lockTarget
                        />
                    </div>
                  </Stack>
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                {submitErrors && submitErrors.length > 0 && <Stack spacing={2} direction={"column"}>
                  <Alert severity="error">{submitErrors.map((err) => {
                      return <div>{err}</div>
                    })}</Alert>
                </Stack>}
                <Button onClick={() => setOpenModal(false)}>Cancel</Button>
                <Button onClick={() => {
                  let errors = testRoute(newRoute);
                  if (errors && errors.length > 0) {
                    errors = errors.map((err) => {
                      return `${err}`;
                    });
                    setSubmitErrors(errors);
                    return true;
                  } else {
                    setSubmitErrors([]);
                    updateRoutes();
                  }
                  
                }}>Connect</Button>
            </DialogActions>
        </>}
    </Dialog>

    <Stack spacing={2}>
      <Stack direction="row" spacing={2}>
        <Input placeholder="Search"
          value={search}
          startAdornment={
            <InputAdornment position="start">
              <SearchOutlined />
            </InputAdornment>
          }
          onChange={(e) => {
            setSearch(e.target.value);
          }}
        />
        <Button variant="contained" startIcon={<ReloadOutlined />} onClick={() => {
          refreshServeApps();
        }}>Refresh</Button>
        <Tooltip title="This is not implemented yet.">
          <span style={{ cursor: 'not-allowed' }}>
            <Button variant="contained" startIcon={<AppstoreAddOutlined />} disabled>Start ServApp</Button>
          </span>
        </Tooltip>
      </Stack>

      <Grid2 container  spacing={2}>
        {serveApps && serveApps.filter(app => search.length < 2 || app.Names[0].includes(search)).map((app) => {
          return <Grid2 style={gridAnim} xs={12} sm={6} md={6} lg={6} xl={4}>
            <Item>
            <Stack justifyContent='space-around' direction="column" spacing={2} padding={2} divider={<Divider orientation="horizontal" flexItem />}>
              <Stack direction="row" spacing={2} alignItems="center">
                <Typography variant="body2" color="text.secondary">
                  {
                    ({
                      "created": <Chip label="Created" color="warning" />,
                      "restarting": <Chip label="Restarting" color="warning" />,
                      "running": <Chip label="Running" color="success" />,
                      "removing": <Chip label="Removing" color="error" />,
                      "paused": <Chip label="Paused" color="info" />,
                      "exited": <Chip label="Exited" color="error" />,
                      "dead": <Chip label="Dead" color="error" />,
                    })[app.State]
                  }
                </Typography>
                <Stack direction="column" spacing={0} alignItems="flex-start">
                  <Typography  variant="h5" color="text.secondary">
                    {app.Names[0].replace('/', '')}&nbsp;
                  </Typography>
                  <Typography style={{ fontSize: '80%' }} color="text.secondary">
                    {app.Image}
                  </Typography>
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Ports
                </Typography> 
                <Stack margin={1} direction="row" spacing={1}>
                  {app.Ports.map((port) => {
                    return <Tooltip title={port.PublicPort ? 'Warning, this port is publicly accessible' : ''}>
                      <Chip style={{ fontSize: '80%' }} label={":" + port.PrivatePort} color={port.PublicPort ? 'warning' : 'default'} />
                    </Tooltip>
                  })}
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Networks
                </Typography> 
                <Stack margin={1} direction="row" spacing={1}>
                  {app.NetworkSettings.Networks && Object.keys(app.NetworkSettings.Networks).map((network) => {
                    return <Chip style={{ fontSize: '80%' }} label={network} color={network === 'bridge' ? 'warning' : 'default'} />
                  })}
                </Stack>
              </Stack>
              {isUpdating[app.Id] ? <div>
                <CircularProgress color="inherit" />
              </div> :
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Settings
                </Typography> 
                <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
                  <Checkbox
                    checked={app.Labels['cosmos-force-network-secured'] === 'true'}
                    onChange={(e) => {
                      setIsUpdatingId(app.Id, true);
                      API.docker.secure(app.Id, e.target.checked).then(() => {
                        setTimeout(() => {
                          setIsUpdatingId(app.Id, false);
                          refreshServeApps();
                        }, 3000);
                      })
                    }}
                  /> Force Secure Network
                </Stack></Stack>}
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Proxies
                </Typography>
                <Stack spacing={2} direction="row">
                  {getContainersRoutes(app.Names[0].replace('/', '')).map((route) => {
                    return <Chip label={route.Host + route.PathPrefix} color="info" />
                  })}
                  {getContainersRoutes(app.Names[0].replace('/', '')).length == 0 &&
                    <Chip label="No Proxy Setup" />}
                </Stack>
              </Stack>
              <Stack>
                <Button variant="contained" color="primary" onClick={() => {
                  setOpenModal(app);
                }}>Connect</Button>
              </Stack>
            </Stack>
          </Item></Grid2>
        })
        }
      </Grid2>
    </Stack>
  </div>
}

export default ServeApps;
