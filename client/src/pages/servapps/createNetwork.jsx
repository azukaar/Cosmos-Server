// material-ui
import * as React from 'react';
import {
  Alert,
  Button,
  Checkbox,
  FormControl,
  FormHelperText,
  Grid,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  CircularProgress,
} from '@mui/material';
import { PlusCircleFilled, PlusCircleOutlined, WarningOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { LoadingButton } from '@mui/lab';
import { FormikProvider, useFormik } from 'formik';
import * as Yup from 'yup';

import * as API from '../../api';
import { CosmosCheckbox } from '../config/users/formShortcuts';
import ResponsiveButton from '../../components/responseiveButton';

const NewNetworkButton = ({ fullWidth, refresh }) => {
  const [isOpened, setIsOpened] = useState(false);
  const formik = useFormik({
    initialValues: {
      name: '',
      driver: 'bridge',
      attachCosmos: false,
    },
    validationSchema: Yup.object({
      name: Yup.string().required('Required'),
      driver: Yup.string().required('Required'),
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

  return <>
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
                    {/* Add more driver options if needed */}
                  </Select>
                </FormControl>
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
        {<DialogActions>
          <Button onClick={() => setIsOpened(false)}>Cancel</Button>
          <LoadingButton
            disabled={formik.errors.submit}
            onClick={formik.handleSubmit}
            loading={formik.isSubmitting}>
            Create
          </LoadingButton>
        </DialogActions>}
      </FormikProvider>
    </Dialog>
    <ResponsiveButton
      fullWidth={fullWidth}
      onClick={() => setIsOpened(true)}
      startIcon={<PlusCircleOutlined />}
    >
      New Network
    </ResponsiveButton>
  </>;
};

export default NewNetworkButton;
