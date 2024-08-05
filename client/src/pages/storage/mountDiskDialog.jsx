import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";
import { useTranslation } from 'react-i18next';

import * as API from '../../api';
import { DownCircleOutlined, UpCircleOutlined } from "@ant-design/icons";

const MountDiskDialogInternal = ({disk, unmount, refresh, open, setOpen }) => {
  const { t } = useTranslation();
  const formik = useFormik({
    initialValues: {
      path: '/mnt/' + disk.name.replace('/dev/', ''),
      permanent: false,
      chown: '1000:1000',
    },
    validationSchema: yup.object({
      // should start with /mnt/ or /var/mnt
      path: yup.string().required(t('global.required')).matches(/^\/(mnt|var\/mnt)\/.{1,}$/, t('mgmt.storage.pathPrefixMntValidation')),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      return (unmount ? API.storage.mounts.unmount({
        mountPoint: disk.mountpoint,
        chown: values.chown,
        permanent: values.permanent,
      }) : API.storage.mounts.mount({
        path: disk.name,
        mountPoint: values.path,
        chown: values.chown,
        permanent: values.permanent,
      })).then((res) => {
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

  return <>
    <Dialog open={open} onClose={() => setOpen(false)}>
          <FormikProvider value={formik}>
            <form onSubmit={formik.handleSubmit}>
            <DialogTitle>{t('mgmt.storage.unMountDiskButton', {unMount: unmount ? t('global.unmount') : t('global.mount')})}
            </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info">
                            {t('mgmt.storage.unMountDiskText', {unMount: unmount ? t('global.unmount') : t('global.mount'), disk: disk.name, mountpoint: disk.mountpoint ? (<> mounted at <strong>{disk.mountpoint}</strong></>) : <></>, unAvailable: unmount ? t('unavailable') : t('available')})}
                        </Alert>
                      </div>
                      {unmount ? '' : <>
                        <TextField
                          fullWidth
                          id="path"
                          name="path"
                          label={t('mgmt.storage.mountPath')}
                          value={formik.values.path}
                          onChange={formik.handleChange}
                          error={formik.touched.path && Boolean(formik.errors.path)}
                          helperText={formik.touched.path && formik.errors.path}
                        />
                        <TextField
                          fullWidth
                          id="chown"
                          name="chown"
                          label={t('mgmt.storage.chown')}
                          value={formik.values.chown}
                          onChange={formik.handleChange}
                          error={formik.touched.chown && Boolean(formik.errors.chown)}
                          helperText={formik.touched.chown && formik.errors.chown}
                        />
                      </>}
                      <div>
                        <Checkbox
                          name="permanent"
                          checked={formik.values.permanent}
                          onChange={formik.handleChange}
                        /> {t('mgmt.storage.mount.permanent')} {unmount ? t('global.unmount') : t('global.mount')}
                      </div>
                      {formik.errors.submit && (
                        <Grid item xs={12}>
                          <FormHelperText error>{formik.errors.submit}</FormHelperText>
                        </Grid>
                      )}
                    </Stack>
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setOpen(false)}>Cancel</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>{unmount ? t('global.unmount') : t('global.mount')}</LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const MountDiskDialog = ({ disk, unmount, refresh, disabled }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return <>
    {open && <MountDiskDialogInternal disk={disk} unmount={unmount} refresh={refresh} open={open} setOpen={setOpen}/>}
    
    <Button
      onClick={() => {setOpen(true);}}
      variant="outlined"
      size="small"
      disabled={disabled}
      startIcon={unmount ?  <DownCircleOutlined /> : <UpCircleOutlined />}
    >{unmount ? t('global.unmount') : t('global.mount')}</Button>
  </>
}


export default MountDiskDialog;