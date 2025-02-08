import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Stack, CircularProgress, Card, Typography, useMediaQuery, Chip, Button } from '@mui/material';
import { CloudServerOutlined, CalendarOutlined, FolderOutlined, ReloadOutlined, InfoCircleOutlined, CloudOutlined } from '@ant-design/icons';
import * as API from '../../api';
import { crontabToText } from '../../utils/indexs';
import ResponsiveButton from '../../components/responseiveButton';
import MainCard from '../../components/MainCard';
import { useTranslation } from 'react-i18next';
import PrettyTableView from '../../components/tableView/prettyTableView';
import { simplifyNumber } from '../dashboard/components/utils';
import { DeleteButton } from "../../components/delete";

export default function SingleRepoIndex({singleBackup}) {
  const { backupName } = useParams();
  const [backup, setBackup] = useState(null);
  const [loading, setLoading] = useState(true);
  const [snapshots, setSnapshots] = useState([]);
  const [isNotFound, setIsNotFound] = useState(false);
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const { t } = useTranslation();

  const infoStyle = {
    backgroundColor: 'rgba(0, 0, 0, 0.1)',
    padding: '10px',
    borderRadius: '5px',
  }

  const fetchBackupDetails = async () => {
    setLoading(true);
    try {
      const configResponse = await API.config.get();
      const backupData = configResponse.data.Backup.Backups[backupName];
      
      if (!backupData) {
        setIsNotFound(true);
        return;
      }
      
      setBackup(backupData);
      
      // Fetch snapshots for this backup
      const snapshotsResponse = singleBackup ? 
       await API.backups.listSnapshots(backupName) :
       await API.backups.listSnapshotsFromRepo(backupName)
       
      if(snapshotsResponse.data && snapshotsResponse.data.reverse) {
        snapshotsResponse.data.reverse();
      }
      setSnapshots(snapshotsResponse.data || []);
    } catch (error) {
      console.error('Error fetching backup details:', error);
    }
    setLoading(false);
  };

  useEffect(() => {
    fetchBackupDetails();
  }, [backupName]);

  if (loading) {
    return (
      <Stack spacing={3} alignContent={'center'} justifyContent={'center'} alignItems={'center'}>
        <CircularProgress />
      </Stack>
    );
  }

  if (!backup) {
    return (
      <div className="p-4">
        <Typography variant="h6" className="text-red-500">
          No repository found for: {backupName}
        </Typography>
      </div>
    );
  }

  return (
    <Stack spacing={3} className="p-4" style={{ maxWidth: '1000px' }}>
    <Stack direction="row" spacing={2} justifyContent="flex-start">
      <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={fetchBackupDetails}>
        {t('global.refresh')}
      </ResponsiveButton>
    </Stack>
    <PrettyTableView
      data={Object.values(snapshots)}
      getKey={(r) => r.Name}
      columns={[
        { 
          title: t('mgmt.backup.id'), 
          field: (r) => r.short_id + ' (' + r.program_version	 + ')',
        },
        { 
          title: t('mgmt.backup.backup'), 
          field: (r) => r.tags.join(', '),
        },
        { 
          title: t('mgmt.backup.date'), 
          field: (r) => (new Date(r.time)).toLocaleString(),
        },
        { 
          title: t('mgmt.backup.data'), 
          field: (r) => simplifyNumber(r.summary.total_bytes_processed, "B") + ' (' + simplifyNumber(r.summary.data_added, "B") + ' added)',
        },
        { 
          title: t('mgmt.backup.files'), 
          field: (r) => simplifyNumber(r.summary.total_files_processed, "") + ' (' + simplifyNumber(r.summary.files_new, "") + ' added)',
        },
        { 
          title: "",
          field: (r) => {
            return <Stack direction="row" spacing={1}>
            <DeleteButton onDelete={async () => {
              await API.backups.forgetSnapshot(backupName, r.id);
            }}></DeleteButton>
            </Stack>
          },
        },
      ]}
    />
    </Stack>
  );
}