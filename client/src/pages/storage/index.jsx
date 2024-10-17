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
import RCloneConfig from './rclone/rclone_config';
import RClonePage from './rclone/rclone_config';
import RemoteStorageSalesPage from './rclone/rclone-free';
import RCloneServePage from './rclone/rclone_serve_config';

const StorageIndex = () => {
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

  let containerized = coStatus && coStatus.containerized;

  return <div>
    <PrettyTabbedView path="/cosmos-ui/storage/:tab" tabs={[
        {
          title: t('mgmt.storage.disks'),
          children: <StorageDisks />,
          path: 'disks'
        },
        {
          title: t('mgmt.storage.mounts'),
          children: <StorageMounts />,
          path: 'mounts'
        },
        {
          title: t('mgmt.storage.externalStorage'),
          children: <div>
            {(!coStatus || !coStatus.Licence) ? <RemoteStorageSalesPage containerized={containerized}/> : <RClonePage containerized={containerized} coStatus={coStatus}/>}
          </div>,
          path: 'external'
        },
        {
          title: t('mgmt.storage.sharesTitle'),
          children: <div>
            {(!coStatus || !coStatus.Licence) ? <RemoteStorageSalesPage /> : <RCloneServePage />}
          </div>,
          path: 'shares'
        },
        {
          title: t('mgmt.storage.mergeTitle'),
          children: <StorageMerges />,
          path: 'mergerfs'
        },
        {
          title: t('mgmt.storage.parityTitle'),
          children: <Parity />,
          path: 'parity'
        },
        {
          title: t('mgmt.storage.raidTitle'),
          children:  <div>
          <Alert severity="info">
            {t('mgmt.storage.raidText')}
          </Alert>
        </div>,
          path: 'raid'
        },
      ]}/>
  </div>;
}

export default StorageIndex;