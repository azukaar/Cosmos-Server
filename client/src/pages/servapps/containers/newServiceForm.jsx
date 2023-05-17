import * as React from 'react';
import MainCard from '../../../components/MainCard';
import RestartModal from '../../config/users/restart';
import { Alert, Button, Checkbox, Chip, Divider, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../../components/routeComponents';
import { getFaviconURL } from '../../../utils/routes';
import * as API from '../../../api';
import { ArrowLeftOutlined, ArrowRightOutlined, CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, UpOutlined } from "@ant-design/icons";
import IsLoggedIn from '../../../isLoggedIn';
import PrettyTabbedView from '../../../components/tabbedView/tabbedView';
import Back from '../../../components/back';
import { useParams } from 'react-router';
import ContainerOverview from './overview';
import Logs from './logs';
import DockerContainerSetup from './setup';
import NetworkContainerSetup from './network';
import VolumeContainerSetup from './volumes';
import DockerTerminal from './terminal';
import NewDockerService from './newService';
import RouteManagement from '../../config/routes/routeman';

const getHostnameFromName = (name) => {
  return name.replace('/', '').replace(/_/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase().replace(/\s/g, '-') + '.' + window.location.origin.split('://')[1]
}

const NewDockerServiceForm = () => {
  const [currentTab, setCurrentTab] = React.useState(0);
  const [maxTab, setMaxTab] = React.useState(0);
  const [containerInfo, setContainerInfo] = React.useState({
    Name: '',
    Names: [],
    Ports: [], // fake for contianerPicker
    Route: null,
    Config: {
      Env: [],
      Labels: {},
      ExposedPorts: [],
      Volumes: {},
      Name: '',
      Image: '',
      Tty: true,
      OpenStdin: true,
    },
    HostConfig: {
      PortBindings: [],
      RestartPolicy: {
        Name: 'always',
      },
      Mounts: [],
    },
    NetworkSettings: {
      Networks: {},
      Ports: [],
    },
  });

  let service = {
    services: {
      container_name : {
        container_name: containerInfo.Name,
        image: containerInfo.Config.Image,
        environment: containerInfo.Config.Env,
        labels: containerInfo.Config.Labels,
        expose: containerInfo.Config.ExposedPorts,
        tty: containerInfo.Config.Tty,
        stdin_open: containerInfo.Config.OpenStdin,
        ports: containerInfo.HostConfig.PortBindings,
        restart: containerInfo.HostConfig.RestartPolicy.Name,
        volumes: containerInfo.HostConfig.Mounts,
        networks: Object.keys(containerInfo.NetworkSettings.Networks).reduce((acc, cur) => {
          acc[cur] = {};
          return acc;
        }, {}),
        routes: containerInfo.CreateRoute ? [containerInfo.Route] : [],
      }
    },
  }

  const nav = () => <Stack
    direction="row"
    spacing={1}
    style={{
      maxWidth: '1000px',
    }}
    >
    <Button
      variant="contained"
      startIcon={<ArrowLeftOutlined />}
      disabled={currentTab === 0}  
      fullWidth
      onClick={() => {
        setCurrentTab(currentTab - 1);
      }}
    >
        Previous
    </Button>
    <Button
      variant="contained"
      fullWidth
      endIcon={<ArrowRightOutlined />}
      disabled={currentTab === 4}
      onClick={() => {
        setCurrentTab(currentTab + 1);
        setMaxTab(Math.max(currentTab + 1, maxTab));
      }}
    >
        Next
    </Button>
    </Stack>

    console.log(containerInfo)

  return <div>
    <Stack spacing={1}>
      <Stack direction="row" spacing={1} alignItems="center">
        <Back />
        <div>Start New Servapp</div>
      </Stack>
      <IsLoggedIn />

      {<PrettyTabbedView
      currentTab={currentTab}
      setCurrentTab={setCurrentTab}
      tabs={[
        {
          title: 'Docker',
          disabled: false,
          children: <Stack spacing={2}><DockerContainerSetup newContainer containerInfo={containerInfo} OnChange={(values) => {
            const newValues = {
              ...containerInfo,
              Name: values.name,
              Names: [values.name],
              Config: {
                ...containerInfo.Config,
                Image: values.image,
                Name: values.name,
                Env: values.envVars,
                Labels: values.labels,
              },
              HostConfig: {
                ...containerInfo.HostConfig,
                RestartPolicy: {
                  Name: values.restartPolicy,
                },
              },
            }
            setContainerInfo(newValues);
          }} 
          OnForceSecure={(value) => {
            const newValues = {
              ...containerInfo,
              Config: {
                ...containerInfo.Config,
                Labels: {
                  ...containerInfo.Config.Labels,
                  'cosmos-force-network-secured': ''+value,
                },
              },
            }
            setContainerInfo(newValues);
          }}/>{nav()}</Stack>
        },
        {
          title: 'URL',
          disabled: maxTab < 1,
          children:  <Stack spacing={2}>
            <MainCard  style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
              <Checkbox
                checked={containerInfo.CreateRoute}
                onChange={(e) => {
                  const newValues = {
                    ...containerInfo,
                    CreateRoute: e.target.checked,
                  }
                  setContainerInfo(newValues);
                }}
              />Create a URL to access this ServApp
            </MainCard>
            {containerInfo.CreateRoute && <RouteManagement TargetContainer={containerInfo} 
              routeConfig={{
                Target: "http://"+containerInfo.Name.replace('/', '') + ":",
                Mode: "SERVAPP",
                Name: containerInfo.Name.replace('/', ''),
                Description: "Expose " + containerInfo.Name.replace('/', '') + " to the internet",
                UseHost: true,
                Host: getHostnameFromName(containerInfo.Name),
                UsePathPrefix: false,
                PathPrefix: '',
                CORSOrigin: '',
                StripPathPrefix: false,
                AuthEnabled: false,
                Timeout: 14400000,
                ThrottlePerMinute: 10000,
                BlockCommonBots: true,
                SmartShield: {
                  Enabled: true,
                }
              }} 
              routeNames={[]}
              setRouteConfig={(newRoute) => {
                const newValues = {
                  ...containerInfo,
                  Route: newRoute,
                }
                setContainerInfo(newValues);
              }}
              up={() => {}}
              down={() => {}}
              deleteRoute={() => {}}
              noControls
              lockTarget
          />}{nav()}</Stack>
        },
        {
          title: 'Network',
          disabled: maxTab < 1,
          children: <Stack spacing={2}><NetworkContainerSetup newContainer containerInfo={containerInfo} OnChange={(values) => {
            const newValues = {
              ...containerInfo,
              Config: {
                ...containerInfo.Config,
                ExposedPorts:
                  values.ports.map((port) => {
                    return `${port.port}/${port.protocol}`;
                  }),
              },
              HostConfig: {
                ...containerInfo.HostConfig,
                PortBindings: values.ports.map((port) => {
                  return `${port.port}:${port.hostPort}/${port.protocol}`;
                }),
              },
              NetworkSettings: {
                ...containerInfo.NetworkSettings,
                Ports: values.ports && values.ports.reduce((acc, port) => {
                  acc[`${port.port}/${port.protocol}`] = [{
                    HostIp: '',
                    HostPort: `${port.hostPort}`,
                  }];
                  return acc;
                }, {}),
              },
            }
            setContainerInfo(newValues);
          }} OnConnect={(net) => {
            const newValues = {
              ...containerInfo,
              NetworkSettings: {
                ...containerInfo.NetworkSettings,
                Networks: {
                  ...containerInfo.NetworkSettings.Networks,
                  [net]: {
                    Name: net,
                  },
                },
              },
            }
            setContainerInfo(newValues);
          }} OnDisconnect={(net) => {
            let newContainerInfo = {
              ...containerInfo,
            }
            delete newContainerInfo.NetworkSettings.Networks[net];
            setContainerInfo(newContainerInfo);
          }}/>{nav()}</Stack>
        },
        {
          title: 'Storage',
          disabled: maxTab < 1,
          children: <Stack spacing={2}><VolumeContainerSetup newContainer containerInfo={containerInfo} OnChange={(values) => {
            console.log(values)
            const newValues = {
              ...containerInfo,
              HostConfig: {
                ...containerInfo.HostConfig,
                Mounts: values.volumes.map((volume) => {
                  return {
                    Type: volume.Type,
                    Source: volume.Source,
                    Target: volume.Target,
                  };
                }),
              },
            }
            setContainerInfo(newValues);
          }} />{nav()}</Stack>
        },
        {
          title: 'Review & Start',
          disabled: maxTab < 1,
          children: <Stack spacing={2}><NewDockerService service={service} />{nav()}</Stack>
        }
      ]} />}

    </Stack>
  </div>;
}

export default NewDockerServiceForm;