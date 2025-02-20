import React, { useEffect, useState } from 'react';
import { Link, useParams } from 'react-router-dom';
import { Stack, CircularProgress, Card, Typography, useMediaQuery, Chip, Button } from '@mui/material';
import { CloudServerOutlined, CalendarOutlined, FolderOutlined, ReloadOutlined, InfoCircleOutlined, CloudOutlined, UpCircleOutlined, DeleteOutlined } from '@ant-design/icons';
import * as API from '../../api';
import { crontabToText } from '../../utils/indexs';
import ResponsiveButton from '../../components/responseiveButton';
import MainCard from '../../components/MainCard';
import { useTranslation } from 'react-i18next';

export default function BackupOverview({backupName}) {
  const [backup, setBackup] = useState(null);
  const [loading, setLoading] = useState(true);
  const [snapshots, setSnapshots] = useState([]);
  const [isNotFound, setIsNotFound] = useState(false);
  const isMobile = useMediaQuery((theme) => theme.breakpoints.down('sm'));
  const { t } = useTranslation();
  const [taskQueued, setTaskQueued] = useState(0);

  const infoStyle = {
    backgroundColor: 'rgba(0, 0, 0, 0.1)',
    padding: '10px',
    borderRadius: '5px',
  }

  const getChip = (backup, snapshots) => {
    if (snapshots.length == 0) {
      return <Chip label="No Snapshots" color="warning" />;
    } 
    else if ((new Date(snapshots[0].time).getTime()) < Date.now() - 1000 * 60 * 60 * 24 * 60) {
      return <Chip label="Old" color="warning" />;
    }
    else {
      return <Chip label="Active" color="success" />;
    }
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
          {t("mgmt.backup.notfound")}: {backupName}
        </Typography>
      </div>
    );
  }

  function backupNow() {
    API.backups.backupNow(backupName).then(() => {
      setTaskQueued((new Date()).getTime() + 1500);
      setTimeout(() => {
        setTaskQueued(0);
      }, 1500);
    });
  }

  function forgetNow() {
    API.backups.forgetNow(backupName).then(() => {
      setTaskQueued((new Date()).getTime() + 1500);
      setTimeout(() => {
        setTaskQueued(0);
      }, 1500);
    });
  }

  return (
    <Stack spacing={3} className="p-4" style={{ maxWidth: '1000px' }}>
      <MainCard name={backup.Name} title={<div>{backup.Name}</div>}>
        <Stack spacing={2} direction={isMobile ? 'column' : 'row'} alignItems={isMobile ? 'center' : 'flex-start'}>
          {/* Left Column */}
          <Stack spacing={2} direction="column" justifyContent="center" alignItems="center">
            <div style={{ position: 'relative' }}>
              <CloudServerOutlined style={{ fontSize: '128px' }} />
            </div>
            <div>
              <strong>{getChip(backup, snapshots)}</strong>
            </div>
            <div>
              <strong>
                {t("mgmt.backup.latest")}: {snapshots.length > 0 ? new Date(snapshots[0].time).toLocaleDateString() : 'N/A'}
              </strong>
            </div>
          </Stack>

          {/* Right Column */}
          <Stack spacing={2} style={{ width: '100%' }}>
            {taskQueued < (new Date()).getTime() ?
            <Stack spacing={2} direction="row">
              <ResponsiveButton
                variant="outlined"
                startIcon={<ReloadOutlined />}
                onClick={fetchBackupDetails}
              >
                {t('global.refresh')}
              </ResponsiveButton>
              <ResponsiveButton
                variant="outlined"
                startIcon={<UpCircleOutlined />}
                onClick={backupNow}
              >
                {t("mgmt.backup.now")}
              </ResponsiveButton>
              <ResponsiveButton
                variant="outlined"
                startIcon={<DeleteOutlined />}
                onClick={forgetNow}
              >
                {t("mgmt.backup.cnow")}
              </ResponsiveButton>
            </Stack> : <Stack style={{fontSize: '120%'}}>{t("mgmt.backup.task")}</Stack>}

            <strong><FolderOutlined /> {t("mgmt.backup.sourceTitle")}</strong>
            <div style={infoStyle}>{backup.Source}</div>

            <strong><InfoCircleOutlined /> {t("mgmt.backup.repositoryTitle")}</strong>
            <div style={infoStyle}>{backup.Repository}</div>

            <strong><CalendarOutlined /> {t("mgmt.backup.backupSchedule")}</strong>
            <div style={infoStyle}>{crontabToText(backup.Crontab)}</div>

            <strong><CalendarOutlined /> {t("mgmt.backup.scheduleForget")}</strong>
            <div style={infoStyle}>{crontabToText(backup.CrontabForget)}</div>
            

            <strong><CloudServerOutlined /> {t("mgmt.backup.latestSnap")}</strong>
            <div>
              {snapshots.slice(0, 5).map((snapshot, index) => (
                <Stack key={index} direction="row" spacing={2} alignItems="center" justifyContent="space-between">
                  <Chip
                    key={index}
                    label={<><strong>{new Date(snapshot.time).toLocaleString()} </strong>{isMobile ? '' : ` - (${snapshot.short_id	 || 'Unknown'})`}</>}
                    style={{ margin: '5px', width: '100%' }}
                  />
                  <Link to={`/cosmos-ui/backups/${backupName}/restore?s=${snapshot.id}`}>
                    <Button variant="contained">{t("global.open")}</Button>
                  </Link>
                </Stack>
              ))}
              {snapshots.length === 0 && (
                <div style={infoStyle}>{t("mgmt.backup.noSnapAvail")}</div>
              )}
            </div>

          </Stack>
        </Stack>
      </MainCard>
    </Stack>
  );
}