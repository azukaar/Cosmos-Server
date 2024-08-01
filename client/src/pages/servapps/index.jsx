import * as React from 'react';
import MainCard from '../../components/MainCard';
import RestartModal from '../config/users/restart';
import { Chip, Divider, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../components/routeComponents';
import { getFaviconURL } from '../../utils/routes';
import * as API from '../../api';
import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, UpOutlined } from "@ant-design/icons";
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import ServApps from './servapps';
import VolumeManagementList from './volumes';
import NetworkManagementList from './networks';
import { useParams } from 'react-router';

const ServappsIndex = () => {
  const { stack } = useParams();

  return <div>
    {!stack && <PrettyTabbedView path="/cosmos-ui/servapps/:tab" tabs={[
        {
          title: 'Containers',
          children: <ServApps stack={stack} />,
          path: 'containers'
        },
        {
          title: 'Volumes',
          children: <VolumeManagementList />,
          path: 'volumes'
        },
        {
          title: 'Networks',
          children:  <NetworkManagementList />,
          path: 'networks'
        },
      ]}/>
    }

    {stack && <ServApps stack={stack} />}

  </div>;
}

export default ServappsIndex;