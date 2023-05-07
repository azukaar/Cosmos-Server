import React, { useState } from 'react';
import { FormikProvider, useFormik } from 'formik';
import * as Yup from 'yup';
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  FormHelperText,
} from '@mui/material';
import { PlusCircleOutlined } from '@ant-design/icons';
import { LoadingButton } from '@mui/lab';
import * as API from '../../api';

const NewVolumeButton = ({ fullWidth, refresh }) => {
  const [isOpened, setIsOpened] = useState(false);

  const formik = useFormik({
    initialValues: {
      name: '',
      driver: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required('Required'),
      driver: Yup.string().required('Required'),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      return API.docker.createVolume(values)
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
          <DialogTitle>New Volume</DialogTitle>
          <DialogContent>
            <DialogContentText></DialogContentText>
            <form onSubmit={formik.handleSubmit}>
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
                  <MenuItem value="local">Local</MenuItem>
                  {/* Add more driver options if needed */}
                </Select>
              </FormControl>
            </form>
            {formik.errors.submit && (
              <Grid item xs={12}>
                <FormHelperText error>{formik.errors.submit}</FormHelperText>
              </Grid>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setIsOpened(false)}>Cancel</Button>
            <LoadingButton
              onClick={formik.handleSubmit}
              loading={formik.isSubmitting}
            >
              Create
            </LoadingButton>
          </DialogActions>
        </FormikProvider>
      </Dialog>
      <Button
        fullWidth={fullWidth}
        onClick={() => setIsOpened(true)}
        startIcon={<PlusCircleOutlined />}
      >
        New Volume
      </Button>
    </>
  );
};

export default NewVolumeButton;