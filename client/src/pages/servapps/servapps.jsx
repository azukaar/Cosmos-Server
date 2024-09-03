// material-ui
import { AlertFilled, AppstoreAddOutlined, CloseSquareOutlined, DeleteOutlined, PauseCircleOutlined, PlaySquareOutlined, PlusCircleOutlined, ReloadOutlined, RollbackOutlined, SearchOutlined, SettingOutlined, StopOutlined, UpCircleOutlined, UpSquareFilled, WarningFilled } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Checkbox, Chip, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, IconButton, Input, InputAdornment, TextField, Tooltip, Typography } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2/Grid2';
import { Stack } from '@mui/system';
import { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';

import * as API from '../../api';
import RestartModal from '../config/users/restart';
import RouteManagement from '../config/routes/routeman';
import { ValidateRoute, getFaviconURL, sanitizeRoute, getContainersRoutes } from '../../utils/routes';
import HostChip from '../../components/hostChip';
import { Link } from 'react-router-dom';
import ExposeModal from './exposeModal';
import GetActions from './actionBar';
import ResponsiveButton from '../../components/responseiveButton';
import DockerComposeImport from './containers/docker-compose';
import { ContainerNetworkWarning } from '../../components/containers';
import { ServAppIcon } from '../../utils/servapp-icon';
import MiniPlotComponent from '../dashboard/components/mini-plot';
import { DownloadFile } from '../../api/downloadButton';
import { useTheme } from '@mui/material/styles';
import { useTranslation } from 'react-i18next';

const Item = styled(Paper)(({ theme }) => ({
  backgroundColor: theme.palette.mode === 'dark' ? '#1A2027' : '#fff',
  ...theme.typography.body2,
  padding: theme.spacing(1),
  textAlign: 'center',
  color: theme.palette.text.secondary,
}));

const StyledBadge = styled(Badge)(({ theme }) => ({
  '& .MuiBadge-badge': {
    right: 5,
    top: -10,
    border: `2px solid ${theme.palette.background.paper}`,
    padding: '0 4px',
  },
}));

const noOver = {
  overflowX: 'auto',
  width: "100%",
  maxWidth: "800px",
  height: "50px"
}

const ServApps = ({stack}) => {
  const { t } = useTranslation();
  const [servApps, setServApps] = useState([]);
  const [isUpdating, setIsUpdating] = useState({});
  const [search, setSearch] = useState("");
  const [config, setConfig] = useState(null);
  const [updatesAvailable, setUpdatesAvailable] = useState(null);
  const [openModal, setOpenModal] = useState(false);
  const [newRoute, setNewRoute] = useState(null);
  const [submitErrors, setSubmitErrors] = useState([]);
  const [openRestartModal, setOpenRestartModal] = useState(false);
  const [selfName, setSelfName] = useState("");

  const refreshServApps = () => {
    API.docker.list().then((res) => {
      setServApps(res.data);
    });
    API.config.get().then((res) => {
      setConfig(res.data);
      setUpdatesAvailable(res.updates);
      setSelfName(res.hostname);
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
    refreshServApps();
  }, []);
  
  function updateRoutes(newRoute) {
    return API.config.addRoute(newRoute).then((res) => {
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

  const getFirstRoute = (app) => {
    let routes = getContainersRoutes(config, app.Names[0].replace('/', ''));
    if(routes.length > 0) {
      return routes[0];
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

  const statusPriority = [
    "running",
    "paused",
    "created",
    "restarting",
    "removing",
    "exited",
    "dead"
  ]

  const servAppsStacked = servApps && servApps.reduce((acc, app) => {
    // if has label cosmos-stack, add to stack
    if(!stack && (app.Labels['cosmos-stack'] || app.Labels['cosmos.stack'] || app.Labels['com.docker.compose.project'])) {
      let stackName = app.Labels['cosmos-stack'] || app.Labels['cosmos.stack'] || app.Labels['com.docker.compose.project'];
      let stackMain = app.Labels['cosmos-stack-main'] ||  app.Labels['cosmos.stack.main'] || (app.Labels['com.docker.compose.container-number'] == '1' && app.Names[0].replace('/', ''));
      
      if(!acc[stackName]) {
        acc[stackName] = {
          type: 'stack',
          name: stackName,
          state: -1,
          app: app,
          apps: [],
          ports: [],
          isUpdating: false,
          updateAvailable: false,
          labels: {
            'cosmos-force-network-secured': 'true',
            'cosmos-auto-update': 'true',
          },
          networkSettings: {
            Networks: {}
          },
        };
      }
      
      acc[stackName].apps.push(app);
      if(statusPriority.indexOf(app.State) > statusPriority.indexOf(acc[stackName].state)) {
        acc[stackName].state = app.State;
      }
      acc[stackName].ports = acc[stackName].ports.concat(app.Ports);
      
      if(!app.Labels['cosmos-force-network-secured']) {
        acc[stackName].labels['cosmos-force-network-secured'] = 'false';
      }

      if(isUpdating[app.Names[0].replace('/', '')]) {
        acc[stackName].isUpdating = true;
      }
      
      if(!app.Labels['cosmos-auto-update']) {
        acc[stackName].labels['cosmos-auto-update'] = 'false';
      }

      acc[stackName].networkSettings = {
        ...acc[stackName].networkSettings,
        ...app.NetworkSettings
      };

      if(updatesAvailable && updatesAvailable[app.Names[0]]) {
        acc[stackName].updateAvailable = true;
      }

      if(stackMain == app.Names[0].replace('/', '') || !acc[stackName].app) {
        acc[stackName].app = app;
      }
    } else if (!stack || (stack && (app.Labels['cosmos-stack'] === stack || app.Labels['cosmos.stack'] === stack || app.Labels['com.docker.compose.project'] === stack))){
      // else add to default stack
      acc[app.Names[0]] = {
        type: 'app',
        name: app.Names[0],
        state: app.State,
        app: app,
        apps: [app],
        isUpdating: isUpdating[app.Names[0].replace('/', '')],
        ports: app.Ports,
        networkSettings: app.NetworkSettings,
        labels: app.Labels,
        updateAvailable: updatesAvailable && updatesAvailable[app.Names[0]],
      };
    }
    return acc;
  }, {});

  // flatten stacks with single app
  Object.keys(servAppsStacked).forEach((key) => {
    if(servAppsStacked[key].type === 'stack' && servAppsStacked[key].apps.length === 1) {
      servAppsStacked[key] = {
        ...servAppsStacked[key],
        type: 'app',
        name: servAppsStacked[key].apps[0].Names[0],
        app: servAppsStacked[key].app,
        apps: [servAppsStacked[key].app],
        isUpdating: isUpdating[servAppsStacked[key].apps[0].Names[0].replace('/', '')],
        ports: servAppsStacked[key].apps[0].Ports,
        networkSettings: servAppsStacked[key].apps[0].NetworkSettings,
        labels: servAppsStacked[key].apps[0].Labels,
        updateAvailable: updatesAvailable && updatesAvailable[servAppsStacked[key].apps[0].Names[0]],
      };
    }
  });

  return <div>
    <RestartModal openModal={openRestartModal} setOpenModal={setOpenRestartModal} config={config} newRoute />
    <ExposeModal
      openModal={openModal} 
      setOpenModal={setOpenModal}
      container={servApps.find((app) => {
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
        {stack && <Link to="/cosmos-ui/servapps">
          <ResponsiveButton variant="secondary" startIcon={<RollbackOutlined />}>Back</ResponsiveButton>
        </Link>}
        <Input placeholder={t('global.searchPlaceholder')}
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
          refreshServApps();
        }}>{t('global.refresh')}</ResponsiveButton>
        {!stack && <>
        <Link to="/cosmos-ui/servapps/new-service">
          <ResponsiveButton
            variant="contained" 
            startIcon={<AppstoreAddOutlined />}
            >{t('navigation.market.startServAppButton')}</ResponsiveButton>
        </Link>
        <DockerComposeImport refresh={refreshServApps}/>
        <DownloadFile
          filename={'backup.cosmos-compose.json'}
          label={t('mgmt.servApps.exportDockerBackupButton.exportDockerBackupLabel')}
          contentGetter={API.config.getBackup}
        />
        </>}
      </Stack>
      
      <Grid2 container spacing={{xs: 1, sm: 1, md: 2 }}>
        {updatesAvailable && updatesAvailable.length && <Grid2 style={gridAnim} xs={12} item>
          <Item>
            <Alert severity="info">{t('mgmt.servapps.updatesAvailableFor')} {Object.keys(updatesAvailable).join(', ')}</Alert>
          </Item>
        </Grid2>}
        {servApps && Object.values(servAppsStacked)
          .filter(app => {
            if (search.length < 2) return true;
            if (app.name.toLowerCase().includes(search.toLowerCase())) return true;
            if (app.app.Image.toLowerCase().includes(search.toLowerCase())) return true;
            if (app.app.Id.toLowerCase().includes(search.toLowerCase())) return true;
            if (app.apps.find((app) => app.Names[0].toLowerCase().includes(search.toLowerCase()))) return true;
          })
          .map((app) => {
          return <Grid2 sx={{...gridAnim}} xs={12} sm={6} md={6} lg={6} xl={4} key={app.app.Id} item>
            <Item>
            <Stack justifyContent='space-around' direction="column" spacing={2} padding={2} divider={<Divider orientation="horizontal" flexItem />}>
              <Stack direction="column" spacing={0} alignItems="flex-start">
                <Stack style={{position: 'relative', overflowX: 'hidden', width: '100%'}} direction="row" spacing={2} alignItems="center">
                  <Typography variant="body2" color="text.secondary">
                    {
                      ({
                        "created": <Chip label={t('mgmt.servApps.createdChip.createdLabel')} color="warning" />,
                        "restarting": <Chip label={t('mgmt.servApps.restartingChip.restartingLabel')} color="warning" />,
                        "running": <Chip label={t('mgmt.servApps.runningChip.runningLabel')} color="success" />,
                        "removing": <Chip label={t('mgmt.servApps.removingChip.removingLabel')} color="error" />,
                        "paused": <Chip label={t('mgmt.servApps.pausedChip.pausedLabel')} color="info" />,
                        "exited": <Chip label={t('mgmt.servApps.exitedChip.exitedLabel')} color="error" />,
                        "dead": <Chip label={t('mgmt.servApps.deadChip.deadLabel')} color="error" />,
                      })[app.state]
                    }
                  </Typography>
                  <Stack direction="row" spacing={2} alignItems="center">
                    {app.type === 'app' && <ServAppIcon container={app.app} route={getFirstRoute(app.app)} className="loading-image" width="40px"/>}
                    {app.type === 'stack' && <StyledBadge  overlap="circular" 
                      anchorOrigin={{
                        vertical: 'bottom',
                        horizontal: 'right',
                      }}color="primary"
                      badgeContent={app.apps.length} >
                      <ServAppIcon container={app.app} route={getFirstRoute(app.app)} className="loading-image" width="40px"/>
                    </StyledBadge>}

                    <Stack direction="column" spacing={0} alignItems="flex-start" style={{height: '40px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'no-wrap'}}>
                      <Typography  variant="h5" color="text.secondary">
                      {app.name.replace('/', '')}&nbsp; {app.type === 'stack' && <Chip label="Stack" color="primary" size="small" style={{fontSize: '55%'}}/>}
                      </Typography>
                      <Typography color="text.secondary" style={{fontSize: '80%', whiteSpace: 'nowrap', overflow: 'hidden', maxWidth: '100%', textOverflow: 'ellipsis'}}>
                        {app.app.Image}
                      </Typography>
                    </Stack>
                  </Stack>
                </Stack>
                <Stack direction="row" spacing={1} width='100%'>
                  <GetActions 
                    Id={app.app.Names[0].replace('/', '')}
                    Ids={app.apps.map((app) => {
                      return app.Names[0].replace('/', '');
                    })}
                    image={app.app.Image}
                    state={app.app.State}
                    setIsUpdatingId={setIsUpdatingId}
                    refreshServApps={refreshServApps}
                    updateAvailable={app.updateAvailable}
                    isStack={app.type === 'stack'}
                    containers={app.apps}
                    config={config}
                  />
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  Ports
                </Typography> 
                <Stack style={noOver} margin={1} direction="row" spacing={1}>
                  {app.ports.filter(p => p.IP != '::').map((port) => {
                    return <Tooltip title={port.PublicPort ? 'Warning, this port is publicly accessible' : ''}>
                      <Chip style={{ fontSize: '80%' }} label={(port.PublicPort ? (port.PublicPort + ":") : '') + port.PrivatePort} color={port.PublicPort ? 'warning' : 'default'} />
                    </Tooltip>
                  })}
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  {t('global.networks')}
                </Typography> 
                <Stack style={noOver} margin={1} direction="row" spacing={1}>
                  {app.networkSettings.Networks && Object.keys(app.networkSettings.Networks).map((network) => {
                    return <Chip style={{ fontSize: '80%' }} label={network} color={network === 'bridge' ? 'warning' : 'default'} />
                  })}
                </Stack>
              </Stack>
              <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                <Typography  variant="h6" color="text.secondary">
                  {t('menu-items.management.urls')}
                </Typography>
                <Stack style={noOver} spacing={2} direction="row">
                  {getContainersRoutes(config, app.name.replace('/', '')).map((route) => {
                    return <HostChip route={route} settings/>
                  })}
                  {/* {getContainersRoutes(config, app.Names[0].replace('/', '')).length == 0 && */}
                    <Chip 
                      label={t('mgmt.servApps.newChip.newLabel')}
                      color="primary"
                      style={{paddingRight: '4px'}}
                      deleteIcon={<PlusCircleOutlined />}
                      onClick={() => {
                        setOpenModal(app.app);
                      }}
                      onDelete={() => {
                        setOpenModal(app.app);
                      }}
                    />
                    {/* } */}
                </Stack>
              </Stack>
              {app.isUpdating ? <div>
                  <CircularProgress color="inherit" />
                </div>
              :
                <Stack margin={1} direction="column" spacing={1} alignItems="flex-start">
                  {/* <Typography  variant="h6" color="text.secondary">
                    Settings {app.type == "app" && (app.state !== 'running' ? '(Start container to edit)' : '')}
                  </Typography> 
                  <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
                    <Checkbox
                      checked={app.labels['cosmos-force-network-secured'] === 'true'}
                      disabled={app.type == "stack" || app.state !== 'running'}
                      onChange={(e) => {
                        const name = app.name.replace('/', '');
                        setIsUpdatingId(name, true);
                        API.docker.secure(name, e.target.checked).then(() => {
                          setTimeout(() => {
                            setIsUpdatingId(name, false);
                            refreshServApps();
                          }, 3000);
                        }).catch(() => {
                          setIsUpdatingId(name, false);
                          refreshServApps();
                        })
                      }}
                    /> Isolate Container Network <ContainerNetworkWarning container={app.app} />
                  </Stack> */}
                  <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
                    <Checkbox
                      checked={app.labels['cosmos-auto-update'] === 'true' ||
                        (selfName && app.name.replace('/', '') == selfName && config.AutoUpdate)}
                      disabled={app.type == "stack" || app.state !== 'running'}
                      onChange={(e) => {
                        const name = app.name.replace('/', '');
                        setIsUpdatingId(name, true);
                        API.docker.autoUpdate(name, e.target.checked).then(() => {
                          setTimeout(() => {
                            setIsUpdatingId(name, false);
                            refreshServApps();
                          }, 3000);
                        }).catch(() => {
                          setIsUpdatingId(name, false);
                          refreshServApps();
                        })
                      }}
                    /> {t('mgmt.servApps.autoUpdateCheckbox')}
                  </Stack>
                </Stack>
              }
              <div>
                <MiniPlotComponent agglo metrics={[
                  "cosmos.system.docker.cpu." + app.name.replace('/', ''),
                  "cosmos.system.docker.ram." + app.name.replace('/', ''),
                ]} labels={{
                  ["cosmos.system.docker.cpu." + app.name.replace('/', '')]: t('global.CPU'), 
                  ["cosmos.system.docker.ram." + app.name.replace('/', '')]: t('global.RAM')
                }}/>
              </div>
 
              <div>
                <Link to={app.type === 'stack' ? 
                  `/cosmos-ui/servapps/stack/${app.name}` : 
                  `/cosmos-ui/servapps/containers/${app.name.replace('/', '')}`
                  }>
                  <Button variant="outlined" color="primary" fullWidth>
                    {app.type === 'stack' ? t('mgmt.servapps.viewStackButton') : t('mgmt.servapps.viewDetailsButton')}
                  </Button>
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

export default ServApps;
