import React, { useState, useEffect } from 'react';
import { Formik, Form, Field } from 'formik';
import { TextField, Select, MenuItem, FormControl, InputLabel, Button, Typography, Stack, Alert, Dialog, DialogTitle, DialogContentText, DialogContent, DialogActions, LinearProgress, Tooltip, CircularProgress } from '@mui/material';
import { ProvConfig, ProvAuth } from './rclone-providers';
import { CosmosCollapse } from '../../config/users/formShortcuts';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import { useTranslation } from 'react-i18next';
import ResponsiveButton from '../../../components/responseiveButton';
import { ArrowsAltOutlined } from '@ant-design/icons';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { simplifyNumber } from '../../dashboard/components/utils';

const RCloneTransfers = ({ transferring, refreshStats }) => {
  const { t } = useTranslation();
  const [open, setOpen] = React.useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      open && refreshStats();
    }, 3000);
    return () => clearInterval(interval);
  }, [open]);

  // sort by progress
  let transferringProcessed = transferring ? transferring.sort((a, b) => {
    return a.bytes / a.size - b.bytes / b.size;
  }) : [];
  // filter top 10
  transferringProcessed = transferringProcessed.slice(0, 30);

  let hasMore = transferring && transferring.length > 30;

  return (<>
    <ResponsiveButton variant="outlined" onClick={() => setOpen(true)} startIcon={<ArrowsAltOutlined />}>
      {t('mgmt.storage.rclone.transfers')}
    </ResponsiveButton>
    <Dialog open={open} onClose={() => setOpen(false)} fullWidth maxWidth="sm">
      <DialogTitle>{t('mgmt.storage.rclone.transfersInProgress')}</DialogTitle>
      <DialogContent>
        <DialogContentText>
          {hasMore && <Alert severity="info">{t('mgmt.storage.rclone.transfersMore', {count: transferring.length - 30})}</Alert>}
          {(transferring && transferring.length > 0) ? <PrettyTableView 
            data={transferringProcessed || []}
            getKey={(r,k) => r.name + k}
            columns={[
              {
                title: t('mgmt.storage.rclone.trans.name'),
                field: (r) => <>{r.name}</>,
              },
              {
                title: t('mgmt.storage.rclone.trans.progress'),
                field: (r) => {
                  let percent = Math.round(r.bytes / r.size * 100);
                  return <Tooltip title={`${simplifyNumber(r.bytes, "B")} / ${simplifyNumber(r.size, "B")}`}> 
                    <Stack direction="row" spacing={1} alignItems="center">
                      <CircularProgress variant="indeterminate" size={12} />
                      <LinearProgress variant="determinate" value={percent} style={{width: '100%'}} />
                    </Stack>
                  </Tooltip>;
                },
              },
              {
                title: t('mgmt.storage.rclone.trans.eta'),
                field: (r) => <>{simplifyNumber(r.eta, "s")}</>,
              },
              {
                title: t('mgmt.storage.rclone.trans.speed'),
                field: (r) => <>{simplifyNumber(r.speed, "B")}/s</>,
              },
              {
                title: t('mgmt.storage.rclone.trans.speedAvg'),
                field: (r) => <>{simplifyNumber(r.speedAvg, "B")}/s</>,
              },
            ]}
          /> : <Typography>{t('mgmt.storage.rclone.transfersEmpty')}</Typography>}
        </DialogContentText>
      </DialogContent>
      <DialogActions>

        <Button onClick={() => setOpen(false)}>{t('global.close')}</Button>
      </DialogActions>
    </Dialog>
  </>);
};

export default RCloneTransfers;