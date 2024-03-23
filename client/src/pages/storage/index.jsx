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
          children: <div>
            <Alert severity="info">
              Coming soon. This feature will allow you to mount external cloud (Dropbox, Onedrive, ...) to your server.
            </Alert>
          </div>,
          path: 'external'
        },
        {
          title: 'Shares',
          children:  <div>
          <Alert severity="info">
            Coming soon. This feature will allow you to share folders with different protocols (SMB, FTP, ...)
          </Alert>
        </div>,
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
          children:  <div>
          <Alert severity="info">
            Coming soon. This feature will allow you to create RAID arrays with your disks.
          </Alert>
        </div>,
          path: 'raid'
        },
      ]}/>
  </div>;
}

export default StorageIndex;