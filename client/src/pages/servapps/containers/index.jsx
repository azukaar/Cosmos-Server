import * as React from 'react';
import MainCard from '../../../components/MainCard';
import RestartModal from '../../config/users/restart';
import { Chip, Divider, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../../components/routeComponents';
import { getFaviconURL } from '../../../utils/routes';
import * as API from '../../../api';
import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, UpOutlined } from "@ant-design/icons";
import IsLoggedIn from '../../../isLoggedIn';
import PrettyTabbedView from '../../../components/tabbedView/tabbedView';
import Back from '../../../components/back';
import { useParams } from 'react-router';
import ContainerOverview from './overview';
import Logs from './logs';

const ContainerIndex = () => {
  const { containerName } = useParams();
  const [container, setContainer] = React.useState(null);
  const [config, setConfig] = React.useState(null);
  
  const refreshContainer = () => {
    return Promise.all([API.docker.get(containerName).then((res) => {
      setContainer(res.data);
    }),
    API.config.get().then((res) => {
      setConfig(res.data);
    })]);
  };

  React.useEffect(() => {
    refreshContainer();
  }, []);

  return <div>
    <Stack spacing={1}>
      <Stack direction="row" spacing={1} alignItems="center">
        <Back />
        <div>{containerName}</div>
      </Stack>
      <IsLoggedIn />

      <PrettyTabbedView 
      isLoading={!container || !config}
      tabs={[
        {
          title: 'Overview',
          children: <ContainerOverview refresh={refreshContainer} containerInfo={container} config={config}/>
        },
        {
          title: 'Logs',
          children: <Logs containerInfo={container} config={config}/>
        },
        {
          title: 'Terminal',
          children: <Logs containerInfo={container} config={config}/>
        },
        {
          title: 'Links',
          children: <div>Links</div>
        },
        // {
        //   title: 'Advanced'
        // },
        {
          title: 'Setup',
          children: <div>Image, Restart Policy, Environment Variables, Labels, etc...</div>
        },
        {
          title: 'Network',
          children: <div>Urls, Networks, Ports, etc...</div>
        },
        {
          title: 'Volumes',
          children: <div>Volumes</div>
        },
        {
          title: 'Resources',
          children: <div>Runtime Resources, Capabilities...</div>
        },
      ]} />
    </Stack>
  </div>;
}

export default ContainerIndex;