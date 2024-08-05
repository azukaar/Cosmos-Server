import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";

import * as API from '../../api';
import { MountPicker } from "./mountPicker";
import ResponsiveButton from "../../components/responseiveButton";
import { PlusCircleOutlined } from "@ant-design/icons";
import { Trans, useTranslation } from 'react-i18next';

const MergerDialogInternal = ({ refresh, open, setOpen, data }) => {  
  const { t } = useTranslation();
  const formik = useFormik({
    initialValues: {
      path: data ? data.path : '/mnt/storage',
      permanent: data ? data.permanent :  false,
      branches: [],
      chown: data ? data.chown : '1000:1000',
      opts: data ? data.opts.join(',') : '',
    },
    validateOnChange: false,
    validationSchema: yup.object({
      // should start with /mnt/ or /var/mnt
      branches: yup.array().min(2, t('mgmt.storage.selectMin2')),
      path: yup.string().required(t('global.required')).matches(/^\/(mnt|var\/mnt)\/.{1,}$/, t('mgmt.storage.pathPrefixMntValidation')),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      return API.storage.mounts.merge({
        branches: values.branches,
        mountPoint: values.path,
        chown: values.chown,
        permanent: values.permanent,
        opts: values.opts,
      }).then((res) => {
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
            <DialogTitle>{t('mgmt.storage.mergeTitle')}</DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info"><Trans i18nKey="mgmt.storage.mergeText" /></Alert>
                      </div>
                      <MountPicker onChange={(value) => formik.setFieldValue('branches', value)} />
                      {formik.errors.branches && (
                          <Grid item xs={12}>
                              <FormHelperText error>{formik.errors.branches}</FormHelperText>
                          </Grid>
                      )}

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
                      <TextField
                        fullWidth
                        id="opts"
                        name="opts"
                        label={t('mgmt.storage.merge.fsOptions.fsOptionsLabel')}
                        value={formik.values.opts}
                        onChange={formik.handleChange}
                        error={formik.touched.opts && Boolean(formik.errors.opts)}
                        helperText={formik.touched.opts && formik.errors.opts}
                      />
                      <div>
                        <Checkbox
                          name="permanent"
                          checked={formik.values.permanent}
                          onChange={formik.handleChange}
                        /> {t('mgmt.storage.mount.permanent')} {t('global.mount')}
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
                  }}>{t('mgmt.storage.mergeButton')}</LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const MergerDialog = ({ refresh, disabled }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return <>
    {open && <MergerDialogInternal refresh={refresh} open={open} setOpen={setOpen}/>}
    
    <ResponsiveButton
      disabled={disabled}
      onClick={() => {setOpen(true);}}
      variant="contained"
      startIcon={<PlusCircleOutlined />}
      size="small"
    >{t('mgmt.storage.newMerge.newMergeButton')}</ResponsiveButton>
  </>
}

export default MergerDialog;
export { MergerDialogInternal };