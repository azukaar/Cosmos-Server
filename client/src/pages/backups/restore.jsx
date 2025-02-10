import React, { useEffect, useState } from 'react';
import { useParams, useSearchParams } from 'react-router-dom';
import { Stack, CircularProgress, Card, Typography, useMediaQuery, Chip, Button, Select, MenuItem } from '@mui/material';
import { CloudServerOutlined, CalendarOutlined, FolderOutlined, ReloadOutlined, InfoCircleOutlined, CloudOutlined } from '@ant-design/icons';
import * as API from '../../api';
import { crontabToText } from '../../utils/indexs';
import ResponsiveButton from '../../components/responseiveButton';
import MainCard from '../../components/MainCard';
import { useTranslation } from 'react-i18next';
import FileExplorer from './fileExplorer';
import BackupFileExplorer from './fileExplorer';

export default function BackupRestore({backupName}) {
  const [sp] = useSearchParams();

  const [backup, setBackup] = useState(null);
  const [loading, setLoading] = useState(true);
  const [snapshots, setSnapshots] = useState([]);
  const [selectedSnapshot, setSelectedSnapshot] = useState(sp.get('s'));
  const [files, setFiles] = useState([]);
  const [isNotFound, setIsNotFound] = useState(false);
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const { t } = useTranslation();

  const getSubfolderRestoreSize = () => {
    API.backups.subfolderRestoreSize(backupName, selectedSnapshot, '/').then((response) => {
      console.log(response)
    });
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
      const snapshotsResponse = await API.backups.listSnapshots(backupName);
      if(snapshotsResponse.data && snapshotsResponse.data.reverse) {
        snapshotsResponse.data.reverse();
      }

      setSnapshots(snapshotsResponse.data || []);

      if(!selectedSnapshot && snapshotsResponse.data.length > 0) {
        setSelectedSnapshot(snapshotsResponse.data[0].id);
      }

      // getSubfolderRestoreSize();
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
      <Stack spacing={2} justifyContent={'center'} alignItems={'center'}>
        <CircularProgress />
      </Stack>
    );
  }

  if (!backup) {
    return (
      <div className="p-4">
        <Typography variant="h6">
          {t("mgmt.backup.notfound")}: {backupName}
        </Typography>
      </div>
    );
  }

  return (
    <Stack spacing={2} style={{ maxWidth: '1000px' }}>
      <div>
      <Select value={selectedSnapshot} onChange={(event) => setSelectedSnapshot(event.target.value)} sx={{ minWidth: 120, marginBottom: '15px' }}>
          {snapshots && snapshots.map((snap, index) => (
            <MenuItem key={snap.id} value={snap.id}>
              {new Date(snap.time).toLocaleString()} - ({snap.short_id})
            </MenuItem>
          ))}
        </Select>
      </div>
      <MainCard name={t("mgmt.backup.files")} title={<div>{t("mgmt.backup.files")}</div>}>
          {selectedSnapshot && <BackupFileExplorer backup={backup} backupName={backupName} selectedSnapshot={selectedSnapshot} getFile={(path) => API.backups.listFolders(backupName, selectedSnapshot, path)} />}
      </MainCard>
    </Stack>
  );
}