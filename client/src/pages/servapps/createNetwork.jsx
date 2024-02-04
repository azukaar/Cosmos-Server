// material-ui
import * as React from 'react';
import { Alert, Button, Checkbox, FormControl, FormHelperText, Grid, InputLabel, MenuItem, Select, Stack, TextField } from '@mui/material';
import { PlusCircleOutlined } from '@ant-design/icons';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { LoadingButton } from '@mui/lab';
import { useFormik, FormikProvider } from 'formik';
import * as Yup from 'yup';

import * as API from '../../api';
import { CosmosCheckbox } from '../config/users/formShortcuts';
import ResponsiveButton from '../../components/responseiveButton';

const NewNetworkButton = ({ fullWidth, refresh }) => {
  const [isOpened, setIsOpened] = React.useState(false);

  const formik = useFormik({
    initialValues: {
      name: '',
      driver: 'bridge',
      attachCosmos: false,
      parentInterface: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required('Required'),
      driver: Yup.string().required('Required'),
      parentInterface: Yup.string().when('driver', {
        is: 'mcvlan',
        then: Yup.string().required('Parent interface is required for MCVLAN')
      }),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      return API.docker.createNetwork(values)
        .then((res) => {
          setStatus({ success: true });
          setSubmitting(false);
          setIsOpened(false);
          refresh && refresh();
        }).catch((err) => {
          setStatus({ success: false });
          setErrors({ submit: err.message });
          setSubmitting(false);
        });
    },
  });

  return (
    <>
      <Dialog open={isOpened} onClose={() => setIsOpened(false)}>
        <FormikProvider value={formik}>
          <DialogTitle>New Network</DialogTitle>
          <DialogContent>
            <DialogContentText>
              <form onSubmit={formik.handleSubmit}>
                <Stack spacing={2} width={'300px'} style={{ marginTop: '10px' }}>
                  <TextField
                    fullWidth
                    id="name"
                    name="name"
                    label="Name"
                    value={formik.values.name}
                    onChange={formik.handleChange}
                    error={formik.touched.name && Boolean(formik.errors.name)}
                    helperText={formik.touched.name && formik.errors.name}
                    style={{ marginBottom: '16px' }}
                  />

                  <FormControl
                    fullWidth
                    variant="outlined"
                    error={formik.touched.driver && Boolean(formik.errors.driver)}
                    style={{ marginBottom: '16px' }}
                  >
                    <InputLabel htmlFor="driver">Driver</InputLabel>
                    <Select
                      id="driver"
                      name="driver"
                      value={formik.values.driver}
                      onChange={formik.handleChange}
                      label="Driver"
                    >
                      <MenuItem value="">
                        <em>None</em>
                      </MenuItem>
                      <MenuItem value="bridge">Bridge</MenuItem>
                      <MenuItem value="host">Host</MenuItem>
                      <MenuItem value="overlay">Overlay</MenuItem>
                      <MenuItem value="mcvlan">MCVLAN</MenuItem>
                    </Select>
                  </FormControl>

                  {formik.values.driver === 'mcvlan' && (
                    <TextField
                      fullWidth
                      id="parentInterface"
                      name="parentInterface"
                      label="Parent Interface"
                      value={formik.values.parentInterface}
                      onChange={formik.handleChange}
                      error={formik.touched.parentInterface && Boolean(formik.errors.parentInterface)}
                      helperText={formik.touched.parentInterface && formik.errors.parentInterface}
                      style={{ marginBottom: '16px' }}
                    />
                  )}

                  <CosmosCheckbox
                    name="attachCosmos"
                    label="Attach to Cosmos"
                    formik={formik}
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
            <Button onClick={() => setIsOpened(false)}>Cancel</Button>
            <LoadingButton
              disabled={formik.errors.submit}
              onClick={formik.handleSubmit}
              loading={formik.isSubmitting}
            >
              Create
            </LoadingButton>
          </DialogActions>
        </FormikProvider>
      </Dialog>
      <ResponsiveButton
        fullWidth={fullWidth}
        onClick={() => setIsOpened(true)}
        startIcon={<PlusCircleOutlined />}
      >
        New Network
      </ResponsiveButton>
    </>
  );
};

export default NewNetworkButton;
