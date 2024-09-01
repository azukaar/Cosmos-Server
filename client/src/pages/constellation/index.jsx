import * as React from 'react';
import MainCard from '../../components/MainCard';
import { Alert, Chip, CircularProgress, Divider, LinearProgress, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../components/routeComponents';
import { getFaviconURL } from '../../utils/routes';
import * as API from '../../api';
import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, UpOutlined } from "@ant-design/icons";
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import { useClientInfos } from '../../utils/hooks';

import { ConstellationVPN } from './vpn';
import { ConstellationDNS } from './dns';
import VPNSalesPage from './free';

const ConstellationIndex = () => {
  const {role} = useClientInfos();
  const isAdmin = role === "2";
  const [coStatus, setCoStatus] = React.useState(null);
  
  const refreshStatus = () => {
    API.getStatus().then((res) => {
      setCoStatus(res.data);
    });
  }

  React.useEffect(() => {
    refreshStatus();
  }, []);

  const ConstContent = isAdmin ? <div>
    <PrettyTabbedView path="/cosmos-ui/constellation/:tab" tabs={[
        {
          title: 'VPN',
          children: <ConstellationVPN />,
          path: 'vpn'
        },
        {
          title: 'DNS',
          children: <ConstellationDNS />,
          path: 'dns'
        },
        {
          title: 'Firewall',
          children: <div>
            <Alert severity="info">
              Coming soon. This feature will allow you to open and close ports individually
              on each device and decide who can access them.
            </Alert>
          </div>,
        },
        {
          title: 'Unsafe Routes',
          children: <div>
            <Alert severity="info">
              Coming soon. This feature will allow you to tunnel your traffic through
              your devices to things outside of your constellation.
            </Alert>
          </div>,
        }
      ]}/>

  </div> : <ConstellationVPN />;

  const pickFree = (coStatus && coStatus.Licence) ? ConstContent : <VPNSalesPage />;

  return coStatus ? pickFree : <center><CircularProgress /></center>;
}

export default ConstellationIndex;