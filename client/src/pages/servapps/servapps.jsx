// material-ui
import { AppstoreAddOutlined, CloseSquareOutlined, DeleteOutlined, PauseCircleOutlined, PlaySquareOutlined, PlusCircleOutlined, ReloadOutlined, RollbackOutlined, SearchOutlined, SettingOutlined, StopOutlined, UpCircleOutlined, UpSquareFilled } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Checkbox, Chip, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, IconButton, Input, InputAdornment, TextField, Tooltip, Typography } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2/Grid2';
import { Stack } from '@mui/system';
import { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';

import * as API from '../../api';
import IsLoggedIn from '../../isLoggedIn';
import RestartModal from '../config/users/restart';
import RouteManagement from '../config/routes/routeman';
import { ValidateRoute, getFaviconURL, sanitizeRoute, getContainersRoutes } from '../../utils/routes';
import HostChip from '../../components/hostChip';
import { Link } from 'react-router-dom';
import ExposeModal from './exposeModal';
import GetActions from './actionBar';
import ResponsiveButton from '../../components/responseiveButton';
import DockerComposeImport from './containers/docker-compose';

const Item = styled(Paper)(({ theme }) => ({
  backgroundColor: theme.palette.mode === 'dark' ? '#1A2027' : '#fff',
  ...theme.typography.body2,
  padding: theme.spacing(1),
  textAlign: 'center',
  color: theme.palette.text.secondary,
}));

const noOver = {
  overflowX: 'auto',
  width: "100%",
  maxWidth: "800px",
  height: "50px"
}

const ServeApps = () => {
  const [serveApps, setServeApps] = useState([]);
  const [isUpdating, setIsUpdating] = useState({});
  const [search, setSearch] = useState("");
  const [config, setConfig] = useState(null);
  const [updatesAvailable, setUpdatesAvailable] = useState(null);
  const [openModal, setOpenModal] = useState(false);
  const [newRoute, setNewRoute] = useState(null);
  const [submitErrors, setSubmitErrors] = useState([]);
  const [openRestartModal, setOpenRestartModal] = useState(false);

  const refreshServeApps = () => {
    API.docker.list().then((res) => {
      setServeApps(res.data);
    });
    API.config.get().then((res) => {
      setConfig(res.data);
      setUpdatesAvailable(res.updates);
    });
    setIsUpdating({});
  };

  const setIsUpdatingId = (id, value) => {
    setIsUpdating({
      ...isUpdating,
      [id]: value
    });
  }

  useEffect(() => {
    refreshServeApps();
  }, []);
  
  function updateRoutes(newRoute) {
    let con = {
      ...config,
      HTTPConfig: {
        ...config.HTTPConfig,
        ProxyConfig: {
          ...config.HTTPConfig.ProxyConfig,
          Routes: [
            newRoute,
            ...config.HTTPConfig.ProxyConfig.Routes,
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

  const selectable = {
    cursor: 'pointer',
    "&:hover": {
      textDecoration: 'underline',
    }
  }

  const getFirstRouteFavIcon = (app) => {
    let routes = getContainersRoutes(config, app.Names[0].replace('/', ''));
    if(routes.length > 0) {
      let url = getFaviconURL(routes[0]);
      return url;
    } else {
      return getFaviconURL('');
    }
  }

  return <div>
    <RestartModal openModal={openRestartModal} setOpenModal={setOpenRestartModal} />
    <ExposeModal
      openModal={openModal} 
      setOpenModal={setOpenModal}
      container={serveApps.find((app) => {
        return app.Names[0].replace('/', '') === (openModal && openModal.Names[0].replace('/', ''));
      })}
      config={config}
      updateRoutes={
        (_newRoute) => {
          updateRoutes(_newRoute);
        }
      }
    />

    <Stack spacing={{xs: 1, sm: 1, md: 2 }}>
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
        <ResponsiveButton variant="contained" startIcon={<ReloadOutlined />} onClick={() => {
          refreshServeApps();
        }}>Refresh</ResponsiveButton>
        <Link to="/ui/servapps/new-service">
          <ResponsiveButton
            variant="contained" 
            startIcon={<AppstoreAddOutlined />}
            >Start ServApp</ResponsiveButton>
        </Link>
        <DockerComposeImport refresh={refreshServeApps}/>
      </Stack>

      <Grid2 container spacing={{xs: 1, sm: 1, md: 2 }}>
        {serveApps && serveApps.filter(app => search.length < 2 || app.Names[0].toLowerCase().includes(search.toLowerCase())).map((app) => {
          return <Grid2 style={gridAnim} xs={12} sm={6} md={6} lg={6} xl={4} key={app.Id} item>
            <Item>
            <Stack justifyContent='space-around' direction="column" spacing={2} padding={2} divider={<Divider orientation="horizontal" flexItem />}>
              <Stack direction="column" spacing={0} alignItems="flex-start">
                <Stack style={{position: 'relative', overflowX: 'hidden', width: '100%'}} direction="row" spacing={2} alignItems="center">
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
                  <Stack direction="row" spacing={2} alignItems="center">
                    <img src={getFirstRouteFavIcon(app)} width="40px" />
                    <Stack direction="column" spacing={0} alignItems="flex-start" style={{height: '40px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'no-wrap'}}>
                      <Typography  variant="h5" color="text.secondary">
                        {app.Names[0].replace('/', '')}&nbsp;
                      </Typography>
                      <Typography color="text.secondary" style={{fontSize: '80%', whiteSpace: 'nowrap', overflow: 'hidden', maxWidth: '100%', textOverflow: 'ellipsis'}}>
                        {app.Image}
                      </Typography>
                    </Stack>
                  </Stack>
                </Stack>
                <Stack direction="row" spacing={2}>
                  {/* <Button variant="contained" size="small" onClick={() => {}}>
                    Update
                  </Button> */}
                  <GetActions 
                    Id={app.Names[0].replace('/', '')}
                    state={app.State}
                    setIsUpdatingId={setIsUpdatingId}
                    refreshServeApps={refreshServeApps}
                    updateAvailable={updatesAvailable[app.Names[0]]}
                  />
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Ports
                </Typography> 
                <Stack style={noOver} margin={1} direction="row" spacing={1}>
                  {app.Ports.filter(p => p.IP != '::').map((port) => {
                    return <Tooltip title={port.PublicPort ? 'Warning, this port is publicly accessible' : ''}>
                      <Chip style={{ fontSize: '80%' }} label={port.PrivatePort + (port.PublicPort ? (":" + port.PublicPort) : '')} color={port.PublicPort ? 'warning' : 'default'} />
                    </Tooltip>
                  })}
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Networks
                </Typography> 
                <Stack style={noOver} margin={1} direction="row" spacing={1}>
                  {app.NetworkSettings.Networks && Object.keys(app.NetworkSettings.Networks).map((network) => {
                    return <Chip style={{ fontSize: '80%' }} label={network} color={network === 'bridge' ? 'warning' : 'default'} />
                  })}
                </Stack>
              </Stack>
              {isUpdating[app.Id] ? <div>
                  <CircularProgress color="inherit" />
                </div>
              :
                <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                  <Typography  variant="h6" color="text.secondary">
                    Settings {app.State !== 'running' ? '(Start container to edit)' : ''}
                  </Typography> 
                  <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
                    <Checkbox
                      checked={app.Labels['cosmos-force-network-secured'] === 'true'}
                      disabled={app.State !== 'running'}
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
                  </Stack>
                </Stack>
              }
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  URLs
                </Typography>
                <Stack style={noOver} spacing={2} direction="row">
                  {getContainersRoutes(config, app.Names[0].replace('/', '')).map((route) => {
                    return <HostChip route={route} settings/>
                  })}
                  {/* {getContainersRoutes(config, app.Names[0].replace('/', '')).length == 0 && */}
                    <Chip 
                      label="New"
                      color="primary"
                      style={{paddingRight: '4px'}}
                      deleteIcon={<PlusCircleOutlined />}
                      onClick={() => {
                        setOpenModal(app);
                      }}
                      onDelete={() => {
                        setOpenModal(app);
                      }}
                    />
                    {/* } */}
                </Stack>
              </Stack>
              <div>
                <Link to={`/ui/servapps/containers/${app.Names[0].replace('/', '')}`}>
                  <Button variant="outlined" color="primary" fullWidth>Details</Button>
                </Link>
              </div>
              {/* <Stack>
                <Button variant="contained" color="primary" onClick={() => {
                  setOpenModal(app);
                }}>Connect</Button>
              </Stack> */}
            </Stack>
          </Item></Grid2>
        })
        }
      </Grid2>
    </Stack>
  </div>
}

export default ServeApps;
