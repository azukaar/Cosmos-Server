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
import { Backups } from './backups';
import { Repositories } from './repositories';

export default function AllBackupsIndex() {
  const { t } = useTranslation();
  
  return <div>
    <Stack spacing={1}>
      <PrettyTabbedView
      tabs={[
        {
          title: t('mgmt.backup.backups'),
          children: <Backups />
        },
        {
          title: t('mgmt.backup.repositories'),
          children: <Repositories />
        },
      ]} />
    </Stack>
  </div>;
}