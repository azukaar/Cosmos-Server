// material-ui
import * as React from 'react';
import { Alert, Button, Checkbox, FormControl, FormHelperText, Grid, InputLabel, MenuItem, Select, Stack, TextField } from '@mui/material';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { LoadingButton } from '@mui/lab';
import { useFormik, FormikProvider } from 'formik';
import * as Yup from 'yup';

const PasswordModal = ({ textInfos, cb, OnClose }) => {

  const formik = useFormik({
    initialValues: {
      password: ''
    },
    validationSchema: Yup.object({
      password: Yup.string().required('Required'),
    }),
    onSubmit: async (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      cb(values.password).then((res) => {
        setStatus({ success: true });
        setSubmitting(false);
        OnClose();
      }).catch((err) => {
        setStatus({ success: false });
        setErrors({ submit: err.message });
        setSubmitting(false);
      });
    },
  });

  return (
    <>
      <Dialog open={true} onClose={() => OnClose()}>
        <FormikProvider value={formik}>
          <DialogTitle>Confirm Password</DialogTitle>
          <DialogContent>
            <DialogContentText>
              <form onSubmit={formik.handleSubmit}>
                <Stack spacing={2} width={'300px'} style={{ marginTop: '10px' }}>
                  <TextField
                    fullWidth
                    id="password"
                    name="password"
                    label="Your Password"
                    value={formik.values.password}
                    type="password"
                    onChange={formik.handleChange}
                    error={formik.touched.password && Boolean(formik.errors.password)}
                    helperText={formik.touched.password && formik.errors.password}
                    style={{ marginBottom: '16px' }}
                  />
                </Stack>
              </form>
              {formik.errors.submit && (
                <Grid item xs={12}>
                  <FormHelperText error>{formik.errors.submit}</FormHelperText>
                </Grid>
              )}
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => OnClose()}>Cancel</Button>
            <LoadingButton
              disabled={formik.errors.submit}
              onClick={formik.handleSubmit}
              loading={formik.isSubmitting}
            >
              Confirm
            </LoadingButton>
          </DialogActions>
        </FormikProvider>
      </Dialog>
    </>
  );
};

export default PasswordModal;
