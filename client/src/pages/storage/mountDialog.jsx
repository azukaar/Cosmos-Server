import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";

import * as API from '../../api';

const MountDialogInternal = ({disk, unmount, refresh, open, setOpen }) => {  
  const formik = useFormik({
    initialValues: {
      path: '/mnt/' + disk.name.replace('/dev/', ''),
      permanent: false,
      chown: '1000:1000',
    },
    validationSchema: yup.object({
      // should start with /mnt/ or /var/mnt
      path: yup.string().required('Required').matches(/^\/(mnt|var\/mnt)\/.{1,}$/, 'Path should start with /mnt/ or /var/mnt'),
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
            <DialogTitle>{unmount ? 'Unmount' : 'Mount'} Disk</DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info">
                          You are about to {unmount ? 'unmount' : 'mount'} the disk <strong>{disk.name}</strong>{disk.mountpoint && (<> mounted at <strong>{disk.mountpoint}</strong></>)}. This will make the content {unmount ? 'unavailable' : 'available'} to be viewed in the file explorer.
                          Permanent {unmount ? 'unmount' : 'mount'} will persist after reboot.
                        </Alert>
                      </div>
                      {unmount ? '' : <>
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
                      </>}
                      <div>
                        <Checkbox
                          name="permanent"
                          checked={formik.values.permanent}
                          onChange={formik.handleChange}
                        /> Permanent {unmount ? 'Unmount' : 'Mount'}
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
                  }}>{unmount ? 'Unmount' : 'Mount'}</LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const MountDialog = ({ disk, unmount, refresh }) => {
  const [open, setOpen] = useState(false);

  return <>
    {open && <MountDialogInternal disk={disk} unmount={unmount} refresh={refresh} open={open} setOpen={setOpen}/>}
    
    <Button
      onClick={() => {setOpen(true);}}
      variant="outlined"
      size="small"
    >{unmount ? 'Unmount' : 'Mount'}</Button>
  </>
}


export default MountDialog;