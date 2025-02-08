import React, { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { Stack, CircularProgress, Card, Typography, useMediaQuery, Chip, Button } from '@mui/material';
import { CloudServerOutlined, CalendarOutlined, FolderOutlined, ReloadOutlined, InfoCircleOutlined, CloudOutlined } from '@ant-design/icons';
import * as API from '../../api';
import { crontabToText } from '../../utils/indexs';
import ResponsiveButton from '../../components/responseiveButton';
import MainCard from '../../components/MainCard';
import { useTranslation } from 'react-i18next';
import BackupOverview from './overview';
import PrettyTabbedView from '../../components/tabbedView/tabbedView';
import Back from '../../components/back';
import BackupRestore from './restore';
import EventExplorerStandalone from '../dashboard/eventsExplorerStandalone';
import SingleRepoIndex from './single-repo-index';

export default function SingleBackupIndex() {
  const { t } = useTranslation();
  const { backupName } = useParams();
  
  return <div>
    <Stack spacing={1}>
      <Stack direction="row" spacing={1} alignItems="center">
        <Back />
        <div>{backupName}</div>
      </Stack>

      <PrettyTabbedView
      rootURL={`/cosmos-ui/backups/${backupName}`}
      tabs={[
        {
          title: t('mgmt.servapps.overview'),
          url: '/',
          children: <BackupOverview backupName={backupName} />
        },
        {
          title: t('mgmt.backup.restore'),
          url: '/restore',
          children: <BackupRestore backupName={backupName} />
        },
        {
          title: t('mgmt.backup.snapshots'),
          url: '/snapshots',
          children: <SingleRepoIndex singleBackup/>
        },
        {
          title: t('navigation.monitoring.eventsTitle'),
          url: '/events',
          children: <EventExplorerStandalone initSearch={`{"object":"backup@${backupName}"}`}/>
        },
      ]} />
    </Stack>
  </div>;
}