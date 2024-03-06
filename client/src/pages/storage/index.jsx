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
import { StorageMerges } from './merges';
import { Parity } from './parity';

const StorageIndex = () => {
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  return <div>
    <PrettyTabbedView path="/cosmos-ui/storage/:tab" tabs={[
        {
          title: 'Disks',
          children: <StorageDisks />,
          path: 'disks'
        },
        {
          title: 'Mounts',
          children: <StorageMounts />,
          path: 'mounts'
        },
        {
          title: 'External Storages',
          children: <StorageMounts />,
          path: 'external'
        },
        {
          title: 'Shares',
          children: <StorageMounts />,
          path: 'shares'
        },
        {
          title: 'Merge Disks',
          children: <StorageMerges />,
          path: 'mergerfs'
        },
        {
          title: 'Parity',
          children: <Parity />,
          path: 'parity'
        },
        {
          title: 'RAID',
          children: <StorageMounts />,
          path: 'raid'
        },
      ]}/>
  </div>;
}

export default StorageIndex;