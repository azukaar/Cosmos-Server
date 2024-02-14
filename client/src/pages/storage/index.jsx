import * as React from 'react';
import MainCard from '../../components/MainCard';
import { Alert, Chip, Divider, Stack, useMediaQuery } from '@mui/material';
import HostChip from '../../components/hostChip';
import { RouteMode, RouteSecurity } from '../../components/routeComponents';
import { getFaviconURL } from '../../utils/routes';
import * as API from '../../api';
import { CheckOutlined, ClockCircleOutlined, DashboardOutlined, DeleteOutlined, DownOutlined, LockOutlined, UpOutlined } from "@ant-design/icons";
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import { useClientInfos } from '../../utils/hooks';
import { StorageMounts } from './mounts';
import { StorageDisks } from './disks';

const StorageIndex = () => {
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  return <div>
    <PrettyTabbedView path="/cosmos-ui/storage/:tab" tabs={[
        {
          title: 'Files',
          children: <StorageMounts />,
          path: 'files'
        },
        {
          title: 'Disks',
          children: <StorageDisks />,
          path: 'disks'
        },
      ]}/>
  </div>;
}

export default StorageIndex;