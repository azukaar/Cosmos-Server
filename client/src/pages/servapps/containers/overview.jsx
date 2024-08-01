import React from 'react';
import { Alert, Checkbox, Chip, CircularProgress, Stack, Typography, useMediaQuery } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { ContainerOutlined, DashboardOutlined, DesktopOutlined, InfoCircleOutlined, NodeExpandOutlined, PlayCircleOutlined, PlusCircleOutlined, SafetyCertificateOutlined, SettingOutlined } from '@ant-design/icons';
import { getFaviconURL, getContainersRoutes } from '../../../utils/routes';
import HostChip from '../../../components/hostChip';
import ExposeModal from '../exposeModal';
import * as API from '../../../api';
import RestartModal from '../../config/users/restart';
import GetActions from '../actionBar';
import { ServAppIcon } from '../../../utils/servapp-icon';
import MiniPlotComponent from '../../dashboard/components/mini-plot';
import UploadButtons from '../../../components/fileUpload';

const info = {
  backgroundColor: 'rgba(0, 0, 0, 0.1)',
  padding: '10px',
  borderRadius: '5px',
}

const ContainerOverview = ({ containerInfo, config, refresh, updatesAvailable, selfName }) => {
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const [openModal, setOpenModal] = React.useState(false);
  const [openRestartModal, setOpenRestartModal] = React.useState(false);
  const [isUpdating, setIsUpdating] = React.useState(false);
  
  const { Name, Config, NetworkSettings, State } = containerInfo;
  const Image = Config.Image;
  const IPAddress = NetworkSettings.Networks?.[Object.keys(NetworkSettings.Networks)[0]]?.IPAddress;
  const Health = State.Health;
  const healthStatus = Health ? Health.Status : 'Healthy';
  const healthIconColor = healthStatus === 'Healthy' ? 'green' : 'red';
  const routes = getContainersRoutes(config, Name.replace('/', ''));

  let refreshAll = refresh ? (() => refresh().then(() => {
    setIsUpdating(false);
  })) : (() => {setIsUpdating(false);});

  const updateRoutes = (newRoute) => {
    API.config.addRoute(newRoute).then(() => {
      refreshAll();
    });
  }

  const addNewRoute = async () => {
    const apps = (await API.docker.list()).data;
    const app = apps.find((a) => a.Names[0] === Name);
    setOpenModal(app);
  }

  return (
    <div style={{ maxWidth: '1000px', width: '100%' }}>
      <RestartModal openModal={openRestartModal} setOpenModal={setOpenRestartModal} config={config} />

      <ExposeModal
        openModal={openModal} 
        setOpenModal={setOpenModal}
        container={Config}
        config={config}
        updateRoutes={
          (_newRoute) => {
            updateRoutes(_newRoute);
            setOpenModal(false);
            setOpenRestartModal(true);
          }
        }
      />
      <MainCard name={Name} title={<div>{Name}</div>}>
        <Stack spacing={2} direction={isMobile ? 'column' : 'row'} alignItems={isMobile ? 'center' : 'flex-start'}>
          <Stack spacing={2} direction={'column'} justifyContent={'center'} alignItems={'center'}>
          <div style={{ position: 'relative' }}>
            <ServAppIcon container={Config} route={routes && routes[0]} className={"loading-image " + (isUpdating ? 'darken' : '')} width="128px" />
            {isUpdating ? (
              <CircularProgress
                style={{ position: 'absolute', top: 'calc(50% - 22.5px)', left: 'calc(50% - 22.5px)' }}
                width="128px"
              />
            ) : null}
          </div>
            <div>
              {({
               "created": <Chip label="Created" color="warning" />,
               "restarting": <Chip label="Restarting" color="warning" />,
               "running": <Chip label="Running" color="success" />,
               "removing": <Chip label="Removing" color="error" />,
               "paused": <Chip label="Paused" color="info" />,
               "exited": <Chip label="Exited" color="error" />,
               "dead": <Chip label="Dead" color="error" />,
             })[State.Status]}
            </div>
            <UploadButtons
              accept='.jpg, .png, .gif, .jpeg, .webp, .bmp, .avif, .tiff, .svg'
              label="icon"
              OnChange={(e) => {
                const file = e.target.files[0];
                setIsUpdating(true);
                API.uploadImage(file, "servapp-" + Name.replace('/', '')).then((data) => {
                  API.docker.updateContainer(Name.replace('/', ''), {
                    labels: {
                      ...Config.Labels,
                      "cosmos-icon": data.data.path,
                    }
                  })
                  .then(() => {
                    refreshAll();
                  });
                });
              }}
            />
          </Stack>

          <Stack spacing={2}  style={{ width: '100%' }} >
            <Stack spacing={2} direction={'row'} >
              <GetActions 
                Id={containerInfo.Name}
                Ids={[containerInfo.Name]}
                image={Image}
                state={State.Status}
                refreshServApps={() => {
                  setIsUpdating(false);
                  refreshAll()
                }}
                setIsUpdatingId={() => {
                  setIsUpdating(true);
                }}
                updateAvailable={updatesAvailable && updatesAvailable[Name]}
                isStack={false}
                containers={[containerInfo]}
                config={config}
              />
            </Stack>
            {containerInfo.State.Status !== 'running' && (
            <Alert severity="warning" style={{ marginBottom: '10px' }}>
                This container is not running. Editing any settings will cause the container to start again.
              </Alert>
            )}
            <strong><ContainerOutlined /> Image</strong>
            <div style={info}>{Image}</div>
            <strong><DesktopOutlined /> ID</strong>
            <div style={info}>{containerInfo.Id}</div>
            <strong><InfoCircleOutlined /> IP Address</strong>
            <div style={info}>{IPAddress}</div>
            <strong>
              <SafetyCertificateOutlined/> Health
            </strong>
            <div style={info}>{healthStatus}</div>
            <strong><SettingOutlined /> Settings {State.Status !== 'running' ? '(Start container to edit)' : ''}</strong>
            {/* <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
              <Checkbox
                checked={Config.Labels['cosmos-force-network-secured'] === 'true'}
                disabled={isUpdating}
                onChange={(e) => {
                  setIsUpdating(true);
                  API.docker.secure(Name.replace('/', ''), e.target.checked).then(() => {
                    setTimeout(() => {
                      refreshAll();
                    }, 3000);
                  })
                }}
              /> Isolate Container Network
            </Stack> */}
            <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
              <Checkbox
                checked={Config.Labels['cosmos-auto-update'] === 'true'  ||
                  (selfName && Name.replace('/', '') == selfName && config.AutoUpdate)}
                disabled={isUpdating}
                onChange={(e) => {
                  setIsUpdating(true);
                  API.docker.autoUpdate(Name.replace('/', ''), e.target.checked).then(() => {
                    setTimeout(() => {
                      refreshAll();
                    }, 3000);
                  })
                }}
              /> Auto Update Container
            </Stack>
            <strong><NodeExpandOutlined /> URLs</strong>
            <div>
              {routes.map((route) => {
                return <HostChip route={route} settings style={{margin: '5px'}}/>
              })}
              <br />
              <Chip 
                label="New"
                color="primary"
                style={{paddingRight: '4px', margin: '5px'}}
                deleteIcon={<PlusCircleOutlined />}
                onClick={() => {
                  addNewRoute();
                }}
                onDelete={() => {
                  addNewRoute();
                }}
              />
            </div>
            <strong><DashboardOutlined /> Monitoring</strong>
              <div style={{ width: '96%' }}>
                <MiniPlotComponent agglo metrics={[
                  "cosmos.system.docker.cpu." + Name.replace('/', ''),
                  "cosmos.system.docker.ram." + Name.replace('/', ''),
                ]} labels={{
                  ["cosmos.system.docker.cpu." + Name.replace('/', '')]: "CPU", 
                  ["cosmos.system.docker.ram." + Name.replace('/', '')]: "RAM"
                }}/>
                <MiniPlotComponent agglo metrics={[
                  "cosmos.system.docker.netTx." + Name.replace('/', ''),
                  "cosmos.system.docker.netRx." + Name.replace('/', ''),
                ]} labels={{
                  ["cosmos.system.docker.netTx." + Name.replace('/', '')]: "NTX", 
                  ["cosmos.system.docker.netRx." + Name.replace('/', '')]: "NRX"
                }}/>
              </div>
          </Stack>
        </Stack>
      </MainCard>
    </div>
  );
};

export default ContainerOverview;