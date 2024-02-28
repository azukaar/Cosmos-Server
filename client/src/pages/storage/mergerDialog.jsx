import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";

import * as API from '../../api';
import { MountPicker } from "./mountPicker";

const MergerDialogInternal = ({ refresh, open, setOpen }) => {  
  const formik = useFormik({
    initialValues: {
      path: '/mnt/storage',
      permanent: false,
      branches: [],
      chown: '1000:1000',
      opts: '',
    },
    validateOnChange: false,
    validationSchema: yup.object({
      // should start with /mnt/ or /var/mnt
      branches: yup.array().min(2, 'Select at least 2 disks'),
      path: yup.string().required('Required').matches(/^\/(mnt|var\/mnt)\/.{1,}$/, 'Path should start with /mnt/ or /var/mnt'),
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
            <DialogTitle>Merge Disks</DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info">
                          You are about to merge disks together. <strong>This operation is safe and reversible</strong>.
                          It will not affect the data on the disks, but will make the content available to be viewed in the file explorer as a single disk.
                        </Alert>
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
                        label="Path to mount to"
                        value={formik.values.path}
                        onChange={formik.handleChange}
                        error={formik.touched.path && Boolean(formik.errors.path)}
                        helperText={formik.touched.path && formik.errors.path}
                      />
                      <TextField
                        fullWidth
                        id="chown"
                        name="chown"
                        label="Change mount folder owner (optional, ex. 1000:1000)"
                        value={formik.values.chown}
                        onChange={formik.handleChange}
                        error={formik.touched.chown && Boolean(formik.errors.chown)}
                        helperText={formik.touched.chown && formik.errors.chown}
                      />
                      <TextField
                        fullWidth
                        id="opts"
                        name="opts"
                        label="Addional mergerFS options (optional, comma separated)"
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
                        /> Permanent {'Mount'}
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
                  }}>Merge</LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const MergerDialog = ({ refresh }) => {
  const [open, setOpen] = useState(false);

  return <>
    {open && <MergerDialogInternal refresh={refresh} open={open} setOpen={setOpen}/>}
    
    <Button
      onClick={() => {setOpen(true);}}
      variant="outlined"
      size="small"
    >Create Merge</Button>
  </>
}

export default MergerDialog;