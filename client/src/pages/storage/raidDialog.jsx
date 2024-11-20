import React, { useState } from 'react';
import { Alert, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormControl, FormControlLabel, FormHelperText, FormLabel, Grid, InputLabel, MenuItem, Select, Stack, TextField } from '@mui/material';
import { LoadingButton } from '@mui/lab';
import { useTranslation } from 'react-i18next';
import { useFormik, FormikProvider } from 'formik';
import * as yup from 'yup';
import * as API from '../../api';
import ResponsiveButton from "../../components/responseiveButton";
import { PlusCircleOutlined } from "@ant-design/icons";
import { MountPicker } from "./mountPicker";
import { Trans } from 'react-i18next';

// ... RaidDialog component remains the same ...

export const RaidDialogInternal = ({ refresh, open, setOpen, data = {} }) => {
  const { t } = useTranslation();
  
  const formik = useFormik({
    initialValues: {
      name: data.device ? data.device.split('/').pop() : t('mgmt.storage.raid.defaultName'),
      level: data["Raid Level"] || 1,
      devices: data.devices?.map(d => d.path) || [],
      mountPoint: data.mountPoint || '',
      permanent: true
    },
    validateOnChange: false,
    validationSchema: yup.object({
      name: yup.string()
        .required(t('global.required'))
        .min(3, t('mgmt.storage.raid.min3chars'))
        .matches(/^[a-zA-Z0-9_]+$/, t('mgmt.storage.raid.notAlphanumeric')),
      level: yup.number().required(t('global.required')),
      devices: yup.array()
        .test('raid-level-devices', t('mgmt.storage.raid.deviceCountError'), function(value) {
          const level = this.parent.level;
          switch(level) {
            case 0: return value.length >= 2;
            case 1: return value.length >= 2;
            case 5: return value.length >= 3;
            case 6: return value.length >= 4;
            case 10: return value.length >= 4 && value.length % 2 === 0;
            default: return false;
          }
        }),
      mountPoint: yup.string()
        .required(t('global.required'))
        .matches(/^\/.*/, t('mgmt.storage.raid.mustStartWithSlash'))
    }),
    onSubmit: async (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      try {
        if (data.device) {
          // Handle device changes for existing array
          const currentDevices = new Set(data.devices.map(d => d.path));
          const newDevices = new Set(values.devices);
          
          // Handle removed devices
          for (const device of currentDevices) {
            if (!newDevices.has(device)) {
              const replacement = values.devices.find(d => !currentDevices.has(d));
              if (replacement) {
                await API.storage.raid.replaceDevice(data.device.split('/').pop(), device, replacement);
              }
            }
          }
          
          // Handle added devices
          for (const device of newDevices) {
            if (!currentDevices.has(device)) {
              await API.storage.raid.addDevice(data.device.split('/').pop(), device);
            }
          }

          // Update mount point if changed
          if (values.mountPoint !== data.mountPoint) {
            if (data.mountPoint) {
              await API.mounts.unmount({ mountPoint: data.mountPoint, permanent: true });
            }
            await API.mounts.mount({
              path: `/dev/md/${values.name}`,
              mountPoint: values.mountPoint,
              permanent: values.permanent
            });
          }
        } else {
          // Create new RAID array
          await API.storage.raid.create({
            name: values.name,
            level: values.level,
            devices: values.devices
          });

          // Mount the new array
          await API.storage.mounts.mount({
            path: `/dev/md/${values.name}`,
            mountPoint: values.mountPoint,
            permanent: values.permanent
          });
        }
        
        setStatus({ success: true });
        setSubmitting(false);
        setOpen(false);
        refresh && refresh();
      } catch (err) {
        setStatus({ success: false });
        setErrors({ submit: err.message });
        setSubmitting(false);
      }
    },
  });

  return (
    <Dialog open={open} onClose={() => !formik.isSubmitting && setOpen(false)} maxWidth="sm" fullWidth>
      <FormikProvider value={formik}>
        <form onSubmit={formik.handleSubmit}>
          <DialogTitle>
            {data.device ? 
              t('mgmt.storage.raid.edit', { name: data.device.split('/').pop() }) : 
              t('mgmt.storage.raid.create')}
          </DialogTitle>
          <DialogContent>
            <DialogContentText>
              <Stack spacing={2} sx={{ mt: 2 }}>
                <div>
                  <Alert severity="info">
                    <Trans i18nKey="mgmt.storage.raid.createInfo" />
                  </Alert>
                </div>

                <TextField
                  fullWidth
                  name="name"
                  label={t('mgmt.storage.raid.name')}
                  value={formik.values.name}
                  onChange={formik.handleChange}
                  error={formik.touched.name && Boolean(formik.errors.name)}
                  helperText={formik.touched.name && formik.errors.name}
                  disabled={!!data.device}
                />

                <FormControl fullWidth>
                  <InputLabel>{t('mgmt.storage.raid.level')}</InputLabel>
                  <Select
                    name="level"
                    label={t('mgmt.storage.raid.level')}
                    value={formik.values.level}
                    onChange={formik.handleChange}
                    error={formik.touched.level && Boolean(formik.errors.level)}
                    disabled={!!data.device}
                  >
                    <MenuItem value={0}>RAID 0 ({t('mgmt.storage.raid.stripe2min')})</MenuItem>
                    <MenuItem value={1}>RAID 1 ({t('mgmt.storage.raid.mirror2min')})</MenuItem>
                    <MenuItem value={5}>RAID 5 ({t('mgmt.storage.raid.raid5min3')})</MenuItem>
                    <MenuItem value={6}>RAID 6 ({t('mgmt.storage.raid.raid6min4')})</MenuItem>
                    <MenuItem value={10}>RAID 10 ({t('mgmt.storage.raid.raid10min4even')})</MenuItem>
                  </Select>
                  {formik.touched.level && formik.errors.level && (
                    <FormHelperText error>{formik.errors.level}</FormHelperText>
                  )}
                </FormControl>

                <FormLabel>
                  <strong>{t('mgmt.storage.raid.selectDevices')}</strong>
                </FormLabel>
                
                <MountPicker 
                  onChange={(value) => formik.setFieldValue('devices', value)} 
                  value={formik.values.devices}
                  diskMode={true}
                  multiple={true}
                />
                {formik.touched.devices && formik.errors.devices && (
                  <Grid item xs={12}>
                    <FormHelperText error>{formik.errors.devices}</FormHelperText>
                  </Grid>
                )}

                <TextField
                  fullWidth
                  name="mountPoint"
                  label={t('mgmt.storage.mountPoint')}
                  value={formik.values.mountPoint}
                  onChange={formik.handleChange}
                  error={formik.touched.mountPoint && Boolean(formik.errors.mountPoint)}
                  helperText={formik.touched.mountPoint && formik.errors.mountPoint}
                />

                <FormControl>
                  <FormControlLabel
                    control={
                      <Checkbox
                        name="permanent"
                        checked={formik.values.permanent}
                        onChange={formik.handleChange}
                      />
                    }
                    label={t('mgmt.storage.mountPermanent')}
                  />
                </FormControl>

                {formik.errors.submit && (
                  <Grid item xs={12}>
                    <FormHelperText error>{formik.errors.submit}</FormHelperText>
                  </Grid>
                )}
              </Stack>
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button 
              onClick={() => setOpen(false)}
              disabled={formik.isSubmitting}
            >
              {t('global.cancel')}
            </Button>
            <LoadingButton
              type="submit"
              variant="contained"
              loading={formik.isSubmitting}
            >
              {data.device ? t('global.save') : t('global.create')}
            </LoadingButton>
          </DialogActions>
        </form>
      </FormikProvider>
    </Dialog>
  );
};

export const RaidDialog = ({ refresh, disabled }) => {
  const [open, setOpen] = useState(false);
  const { t } = useTranslation();

  return (
    <>
      <ResponsiveButton
        variant="contained"
        startIcon={<PlusCircleOutlined />}
        onClick={() => setOpen(true)}
        disabled={disabled}
      >
        {t('mgmt.storage.raid.create')}
      </ResponsiveButton>
      <RaidDialogInternal
        refresh={refresh}
        open={open}
        setOpen={setOpen}
      />
    </>
  );
};