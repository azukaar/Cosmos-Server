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
import { useTranslation } from 'react-i18next';


import { ConstellationVPN } from './vpn';
import { ConstellationDNS } from './dns';
import VPNSalesPage from './free';

const ConstellationIndex = () => {
  const { t } = useTranslation();
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
              {t('mgmt.constellation.setup.firewallInfo')}
            </Alert>
          </div>,
        },
        {
          title: t('mgmt.constellation.setup.unsafeRoutesTitle'),
          children: <div>
            <Alert severity="info">
              {t('mgmt.constellation.setup.unsafeRoutesText')}
            </Alert>
          </div>,
        }
      ]}/>

  </div> : <ConstellationVPN />;

  const pickFree = (coStatus && coStatus.Licence) ? ConstContent : <><ConstellationVPN freeVersion /></>;

  return coStatus ? pickFree : <center><CircularProgress /></center>;
}

export default ConstellationIndex;