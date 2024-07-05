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
import { useTranslation } from 'react-i18next';

const StorageIndex = () => {
  const { t } = useTranslation();
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  return <div>
    <PrettyTabbedView path="/cosmos-ui/storage/:tab" tabs={[
        {
          title: t('Disks'),
          children: <StorageDisks />,
          path: 'disks'
        },
        {
          title: t('Mounts'),
          children: <StorageMounts />,
          path: 'mounts'
        },
        {
          title: t('ExternalStorages'),
          children: <div>
            <Alert severity="info">
              {t('ExternalStorageDesc')}
            </Alert>
          </div>,
          path: 'external'
        },
        {
          title: t('Shares'),
          children:  <div>
          <Alert severity="info">
            {t('SharesDesc')}
          </Alert>
        </div>,
          path: 'shares'
        },
        {
          title: t('MergeDisks'),
          children: <StorageMerges />,
          path: 'mergerfs'
        },
        {
          title: t('Parity'),
          children: <Parity />,
          path: 'parity'
        },
        {
          title: t('RAID'),
          children:  <div>
          <Alert severity="info">
            {t('RAIDDesc')}
          </Alert>
        </div>,
          path: 'raid'
        },
      ]}/>
  </div>;
}

export default StorageIndex;