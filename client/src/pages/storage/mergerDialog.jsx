import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";

import * as API from '../../api';

const MergerDialog = ({ refresh }) => {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  
  const formik = useFormik({
    initialValues: {
      path: '/mnt/storage',
      permanent: false,
    },
    validationSchema: yup.object({
      // should start with /mnt/ or /var/mnt
      path: yup.string().required('Required').matches(/^\/(mnt|var\/mnt)\/.{1,}$/, 'Path should start with /mnt/ or /var/mnt'),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      return API.storage.mounts.merge({
        branches: ['/mnt/sda1'],
        mountPoint: values.path,
        chown: '',
        permanent: values.permanent,
        opts: '',
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
              {open && <>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info">
                          You are about to merge disks together. <strong>This operation is safe and reversible</strong>.
                          It will not affect the data on the disks, but will make the content available to be viewed in the file explorer as a single disk.
                        </Alert>
                      </div>
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
                      <div>
                        <Checkbox
                          name="permanent"
                          checked={formik.values.permanent}
                          onChange={formik.handleChange}
                        /> Permanent {'Mount'}
                      </div>
                    </Stack>
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  {formik.errors.submit && (
                    <Grid item xs={12}>
                      <FormHelperText error>{formik.errors.submit}</FormHelperText>
                    </Grid>
                  )}
                  <Button onClick={() => setOpen(false)}>Cancel</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>Merge</LoadingButton>
                </DialogActions>
              </>}
            </form>
        </FormikProvider>
    </Dialog>
    <LoadingButton
      loading={loading}
      onClick={() => {setOpen(true);}}
      variant="outlined"
      size="small"
    >Create Merge</LoadingButton>
  </>
}

export default MergerDialog;