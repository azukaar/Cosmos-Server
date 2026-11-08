import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { FormikProvider, useFormik } from 'formik';
import {
  Alert,
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from '@mui/material';
import { LoadingButton } from '@mui/lab';
import { PlusCircleOutlined, DeleteOutlined, SyncOutlined } from '@ant-design/icons';

import * as API from '../../../api';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { ConfirmModalDirect } from '../../../components/confirmModal';
import PermissionGuard from '../../../components/permissionGuard';
import { PERM_CONFIGURATION } from '../../../utils/permissions';
import { CosmosCheckbox, CosmosInputText } from './formShortcuts';

const isValidIPOrCIDR = (value) => {
  const v = (value || '').trim();
  if (!v) return false;
  const [ip, prefix, extra] = v.split('/');
  if (extra !== undefined) return false;
  const v4 = /^(\d{1,3})(\.\d{1,3}){3}$/.test(ip) && ip.split('.').every((n) => Number(n) <= 255);
  const v6 = ip.includes(':') && /^[0-9a-fA-F:.]+$/.test(ip);
  if (!v4 && !v6) return false;
  if (prefix === undefined) return true;
  const n = Number(prefix);
  return /^\d+$/.test(prefix) && n >= 0 && n <= (v4 ? 32 : 128);
};

const AddEntryDialog = ({ open, onClose, onAdd, t }) => {
  const formik = useFormik({
    initialValues: { IP: '', Label: '', BypassGeo: false, BypassIPRestriction: false },
    validateOnChange: false,
    validate: (values) => {
      const errors = {};
      if (!isValidIPOrCIDR(values.IP)) errors.IP = t('mgmt.shield.whitelist.invalidIP');
      return errors;
    },
    onSubmit: async (values, { setErrors, setSubmitting, resetForm }) => {
      try {
        await onAdd({ ...values, IP: values.IP.trim() });
        resetForm();
        onClose();
      } catch (err) {
        setErrors({ submit: err.message });
      }
      setSubmitting(false);
    },
  });

  return <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
    <FormikProvider value={formik}>
      <form onSubmit={formik.handleSubmit}>
        <DialogTitle>{t('mgmt.shield.whitelist.add')}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} style={{ marginTop: '10px' }}>
            <CosmosInputText name="IP" label={t('mgmt.shield.whitelist.ip')} placeholder={t('mgmt.shield.whitelist.ipHelper')} formik={formik} />
            <CosmosInputText name="Label" label={t('mgmt.shield.whitelist.label')} formik={formik} />
            <CosmosCheckbox name="BypassGeo" label={t('mgmt.shield.whitelist.bypassGeo')} formik={formik} />
            <CosmosCheckbox name="BypassIPRestriction" label={t('mgmt.shield.whitelist.bypassIP')} formik={formik} />
            {formik.errors.submit && <Alert severity="error">{formik.errors.submit}</Alert>}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose}>{t('global.cancelAction')}</Button>
          <LoadingButton type="submit" variant="contained" loading={formik.isSubmitting}>{t('global.createAction')}</LoadingButton>
        </DialogActions>
      </form>
    </FormikProvider>
  </Dialog>;
};

const ShieldWhitelist = () => {
  const { t } = useTranslation();
  const [config, setConfig] = useState(null);
  const [error, setError] = useState(null);
  const [openAdd, setOpenAdd] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(null);

  const refresh = () => {
    API.config.get().then((res) => setConfig(res.data));
  };

  useEffect(() => { refresh(); }, []);

  const entries = (config && config.ShieldWhitelist) || [];

  const save = (list) => {
    setError(null);
    return API.config.set({ ...config, ShieldWhitelist: list }).then(() => refresh()).catch((err) => {
      setError(err.message);
      throw err;
    });
  };

  return <div style={{ maxWidth: '1200px', margin: 'auto' }}>
    <AddEntryDialog open={openAdd} onClose={() => setOpenAdd(false)} t={t} onAdd={(entry) => save([...entries, entry])} />
    {confirmDelete !== null && <ConfirmModalDirect
      callback={() => save(entries.filter((_, i) => i !== confirmDelete))}
      content={t('mgmt.shield.whitelist.deleteConfirm', { ip: entries[confirmDelete] ? entries[confirmDelete].IP : '' })}
      onClose={() => setConfirmDelete(null)}
    />}

    <Typography variant="h5">{t('mgmt.shield.whitelist.title')}</Typography>
    {error && <Alert severity="error" style={{ marginBottom: '10px' }}>{error}</Alert>}

    <PrettyTableView
      isLoading={config === null}
      data={entries}
      getKey={(e, i) => i + e.IP}
      buttons={[
        <PermissionGuard key="add" permission={PERM_CONFIGURATION}>
          <Button variant="contained" startIcon={<PlusCircleOutlined />} onClick={() => setOpenAdd(true)}>{t('mgmt.shield.whitelist.add')}</Button>
        </PermissionGuard>,
        <Button key="refresh" variant="outlined" startIcon={<SyncOutlined />} onClick={refresh}>{t('global.refresh')}</Button>,
      ]}
      columns={[
        { title: t('mgmt.shield.whitelist.ip'), search: (e) => e.IP + ' ' + e.Label, field: (e) => <strong>{e.IP}</strong> },
        { title: t('mgmt.shield.whitelist.label'), screenMin: 'md', field: (e) => e.Label },
        { title: t('mgmt.shield.whitelist.bypassGeo'), field: (e) => <Checkbox disabled checked={!!e.BypassGeo} /> },
        { title: t('mgmt.shield.whitelist.bypassIP'), field: (e) => <Checkbox disabled checked={!!e.BypassIPRestriction} /> },
        {
          title: '',
          field: (e, i) => <PermissionGuard permission={PERM_CONFIGURATION}>
            <Button size="small" variant="outlined" color="error" startIcon={<DeleteOutlined />} onClick={() => setConfirmDelete(entries.indexOf(e))}>
              {t('global.delete')}
            </Button>
          </PermissionGuard>,
          style: { textAlign: 'right' },
        },
      ]}
    />
  </div>;
};

export default ShieldWhitelist;
