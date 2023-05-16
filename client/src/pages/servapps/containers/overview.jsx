import React from 'react';
import { Alert, Checkbox, Chip, CircularProgress, Stack, Typography, useMediaQuery } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { ContainerOutlined, DesktopOutlined, InfoCircleOutlined, NodeExpandOutlined, PlayCircleOutlined, PlusCircleOutlined, SafetyCertificateOutlined, SettingOutlined } from '@ant-design/icons';
import { getFaviconURL, getContainersRoutes } from '../../../utils/routes';
import HostChip from '../../../components/hostChip';
import ExposeModal from '../exposeModal';
import * as API from '../../../api';
import RestartModal from '../../config/users/restart';
import GetActions from '../actionBar';

const info = {
  backgroundColor: 'rgba(0, 0, 0, 0.1)',
  padding: '10px',
  borderRadius: '5px',
}

const ContainerOverview = ({ containerInfo, config, refresh }) => {
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

  let refreshAll = refresh && (() => refresh().then(() => {
    setIsUpdating(false);
  }));

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
      <RestartModal openModal={openRestartModal} setOpenModal={setOpenRestartModal} />

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
            <img className={isUpdating ? 'darken' : ''} src={getFaviconURL(routes && routes[0])} width="128px" />
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
          </Stack>

          <Stack spacing={2}  style={{ width: '100%' }} >
            <Stack spacing={2} direction={'row'} >
              <GetActions 
                Id={containerInfo.Name}
                image={Image}
                state={State.Status}
                refreshServeApps={() => {
                  refreshAll()
                }}
                setIsUpdatingId={() => {
                  setIsUpdating(true);
                }}
              />
            </Stack>
            {containerInfo.State.Status !== 'running' && (
            <Alert severity="warning" style={{ marginBottom: '10px' }}>
                This container is not running. Editing any settings will cause the container to start again.
              </Alert>
            )}
            <strong><ContainerOutlined /> Image</strong>
            <div style={info}>{Image}</div>
            <strong><DesktopOutlined /> Name</strong>
            <div style={info}>{Name}</div>
            <strong><InfoCircleOutlined /> IP Address</strong>
            <div style={info}>{IPAddress}</div>
            <strong>
              <SafetyCertificateOutlined/> Health
            </strong>
            <div style={info}>{healthStatus}</div>
            <strong><SettingOutlined /> Settings {State.Status !== 'running' ? '(Start container to edit)' : ''}</strong>
            <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
              <Checkbox
                checked={Config.Labels['cosmos-force-network-secured'] === 'true'}
                disabled={isUpdating}
                onChange={(e) => {
                  setIsUpdating(true);
                  API.docker.secure(Name, e.target.checked).then(() => {
                    setTimeout(() => {
                      refreshAll();
                    }, 3000);
                  })
                }}
              /> Force Secure Network
            </Stack>
            <Stack style={{ fontSize: '80%' }} direction={"row"} alignItems="center">
              <Checkbox
                checked={Config.Labels['cosmos-auto-update'] === 'true'}
                disabled={isUpdating}
                onChange={(e) => {
                  setIsUpdating(true);
                  API.docker.autoUpdate(Name, e.target.checked).then(() => {
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
          </Stack>
        </Stack>
      </MainCard>
    </div>
  );
};

export default ContainerOverview;