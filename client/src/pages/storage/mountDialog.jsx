import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";
import { useTranslation } from 'react-i18next';

import * as API from '../../api';

const MountDialogInternal = ({ unmount, refresh, open, setOpen, data }) => {
  const { t } = useTranslation();
  const formik = useFormik({
    initialValues: {
      device: data ? data.device : '',
      path: data ? data.path : '',
      permanent: false,
      chown: '1000:1000',
    },
    validationSchema: yup.object({
      // should start with /mnt/ or /var/mnt
      device: yup.string().required(t('global.required')),
      path: yup.string().required(t('global.required')).matches(/^\/(mnt|var\/mnt)\/.{1,}$/, t('mgmt.storage.pathPrefixMntValidation')),
    }),
    onSubmit: async (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      if(data && !unmount) {
        try {
          await API.storage.mounts.unmount({
            mountPoint: data.path,
            permanent: true,
          });
        } catch (err) {
          console.error(err);
        }
      }
      return (unmount ? API.storage.mounts.unmount({
        mountPoint: values.path,
        chown: values.chown,
        permanent: values.permanent,
      }) : API.storage.mounts.mount({
        path: values.device,
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
                            {t('mgmt.storage.unMountText', {unMount: unmount ? t('global.unmount') : t('global.mount'), mountpoint: data && data.mountpoint && (<> t('mgmt.storage.mountedAtText') <strong>{data && data.mountpoint}</strong></>), unAvailable: unmount ? t('mgmt.storage.unavailable') : t('mgmt.storage.available')})}
                            </Alert>
                      </div>
                      {unmount ? '' : <>
                        <TextField
                          fullWidth
                          id="device"
                          name="device"
                          label={t('mgmt.storage.mount.whatToMountLabel')}
                          value={formik.values.device}
                          onChange={formik.handleChange}
                          error={formik.touched.device && Boolean(formik.errors.device)}
                          helperText={formik.touched.device && formik.errors.device}
                        />
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
                  <Button onClick={() => setOpen(false)}>{t('global.cancelAction')}</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>{unmount ? t('global.unmount') : t('global.mount')}</LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const MountDialog = ({ disk, unmount, refresh, disabled }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return <>
    {open && <MountDialogInternal disk={disk} unmount={unmount} refresh={refresh} open={open} setOpen={setOpen}/>}
    
    <Button
      disabled={disabled}
      onClick={() => {setOpen(true);}}
      variant="outlined"
      size="small"
    >{unmount ? t('global.unmount') : t('global.mount')}</Button>
  </>
}


export default MountDialog;
export { MountDialogInternal };
