import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, { useState } from "react";
import { FormikProvider, useFormik } from "formik";
import * as yup from "yup";
import * as API from '../../api';
import ResponsiveButton from "../../components/responseiveButton";
import { PlusCircleOutlined } from "@ant-design/icons";
import { crontabToText } from "../../utils/indexs";
import { Trans, useTranslation } from 'react-i18next';
import { FilePickerButton } from '../../components/filePicker';
import { CosmosInputText } from "../config/users/formShortcuts";

const isAbsolutePath = (path) => path.startsWith('/') || path.startsWith('rclone:') || /^[a-zA-Z]:\\/.test(path); // Unix & Windows support

const BackupDialogInternal = ({ refresh, open, setOpen, preSource, preName, data = {} }) => {
  const { t } = useTranslation();
  const isEdit = Boolean(data.Name);
  const formik = useFormik({
    initialValues: {
      name: preName || data.Name || t('mgmt.backup.newBackup'),
      source: preSource || data.Source || '',
      repository: data.Repository || '/backups',
      crontab: data.Crontab || '0 0 4 * * *',
      crontabForget: data.CrontabForget || '0 0 12 * * *',
      retentionPolicy: data.RetentionPolicy || '--keep-last 3 --keep-daily 7 --keep-weekly 8 --keep-yearly 3'
    },
    validationSchema: yup.object({
      name: yup
        .string()
        .required(t('global.required'))
        .min(3, t('mgmt.backup.min3chars')),
      
      source: yup
        .string()
        .required(t('global.required'))
        .test('is-absolute', t('global.absolutePath'), isAbsolutePath),
    
      repository: yup
        .string()
        .required(t('global.required'))
        .test('is-absolute', t('global.absolutePath'), isAbsolutePath),
    
      crontab: yup.string().required(t('global.required')),
      crontabForget: yup.string().required(t('global.required')),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      (data.Name ? API.backups.editBackup({
        ...values
      }) : API.backups.addBackup({
        ...values
      })).then(() => {
        setStatus({ success: true });
        setSubmitting(false);
        setOpen(false);
        refresh && refresh();
      }).catch((err) => {
        setStatus({ success: false });
        setErrors({ submit: err.message });
        setSubmitting(false);
      });
    },
  });

  return (
    <Dialog open={open} onClose={() => setOpen(false)}>
      <FormikProvider value={formik}>
        <form onSubmit={formik.handleSubmit}>
          <DialogTitle>
            {data.Name ? (t('global.edit') + " " + data.Name) : t('mgmt.backup.createBackup')}
          </DialogTitle>
          <DialogContent>
            <DialogContentText>
              <Stack spacing={3} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                <Alert severity="info">
                  <Trans i18nKey="mgmt.backup.createBackupInfo" />
                </Alert>
                <TextField
                  fullWidth
                  name="name"
                  disabled={isEdit}
                  label={t('global.name')}
                  value={formik.values.name}
                  onChange={formik.handleChange}
                  error={formik.touched.name && Boolean(formik.errors.name)}
                  helperText={formik.touched.name && formik.errors.name}
                />
                
                <Stack direction="row" spacing={2} alignItems="flex-end">
                  {!isEdit && <FilePickerButton
                    onPick={(path) => {
                      if(path)
                        formik.setFieldValue('source', path);
                    }} 
                    size="150%" 
                    select="folder" 
                  />}
                  <CosmosInputText
                    name="source"
                    disabled={isEdit}
                    label={t('mgmt.backup.source')}
                    placeholder="/folder/to/backup"
                    formik={formik}
                  />
                </Stack>

                <Stack direction="row" spacing={2} alignItems="flex-end">
                  {!isEdit && <FilePickerButton 
                    onPick={(path) => {
                      if(path)
                        formik.setFieldValue('repository', path);
                    }}
                    restic
                    raw
                    canCreate={true}
                    size="150%" 
                    select="folder" 
                  />}
                  <CosmosInputText
                    name="repository"
                    disabled={isEdit}
                    label={t('mgmt.backup.repository')}
                    placeholder="/where/to/store/backups"
                    formik={formik}
                  />
                </Stack>

                <TextField
                  fullWidth
                  name="crontab"
                  label={t('mgmt.backup.schedule')}
                  value={formik.values.crontab}
                  onChange={formik.handleChange}
                  error={formik.touched.crontab && Boolean(formik.errors.crontab)}
                  helperText={formik.touched.crontab && formik.errors.crontab || crontabToText(formik.values.crontab, t)}
                />

                <TextField
                  fullWidth
                  name="crontabForget"
                  label={t('mgmt.backup.scheduleForget')}
                  value={formik.values.crontabForget}
                  onChange={formik.handleChange}
                  error={formik.touched.crontabForget && Boolean(formik.errors.crontabForget)}
                  helperText={formik.touched.crontabForget && formik.errors.crontabForget || crontabToText(formik.values.crontabForget, t)}
                />

                <TextField
                  fullWidth
                  name="retentionPolicy"
                  label={t('mgmt.backup.rententionPolicy')}
                  value={formik.values.retentionPolicy}
                  onChange={formik.handleChange}
                  error={formik.touched.retentionPolicy && Boolean(formik.errors.retentionPolicy)}
                  helperText={formik.touched.retentionPolicy && formik.errors.retentionPolicy}
                />

                {formik.errors.submit && (
                  <Grid item xs={12}>
                    <FormHelperText error>{formik.errors.submit}</FormHelperText>
                  </Grid>
                )}
              </Stack>
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setOpen(false)}>{t('global.cancelAction')}</Button>
            <LoadingButton 
              color="primary" 
              variant="contained" 
              type="submit" 
              loading={formik.isSubmitting}
            >
              {data.Name ? t('global.update') : t('global.createAction')}
            </LoadingButton>
          </DialogActions>
        </form>
      </FormikProvider>
    </Dialog>
  );
};

const BackupDialog = ({ refresh, data, preSource, preName }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return (
    <>
      {open && <BackupDialogInternal preSource={preSource} preName={preName} refresh={refresh} open={open} setOpen={setOpen} data={data} />}
      <div>
        {!data ? (
          <ResponsiveButton
            onClick={() => setOpen(true)}
            variant="contained"
            size="small"
            startIcon={<PlusCircleOutlined />}
          >
            {t('mgmt.backup.newBackup')}
          </ResponsiveButton>
        ) : (
          <div onClick={() => setOpen(true)}>{t('global.edit')}</div>
        )}
      </div>
    </>
  );
};

export default BackupDialog;
export { BackupDialog, BackupDialogInternal };