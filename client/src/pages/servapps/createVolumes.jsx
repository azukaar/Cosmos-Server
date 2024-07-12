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
import { useTranslation } from 'react-i18next';

const NewVolumeButton = ({ fullWidth, refresh }) => {
  const { t } = useTranslation();
  const [isOpened, setIsOpened] = useState(false);
  const formik = useFormik({
    initialValues: {
      name: '',
      driver: 'local',
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
          <DialogTitle>{t('mgmt.servApps.volumes.newVolumeTitle')}</DialogTitle>
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
                <InputLabel htmlFor="driver">{t('global.driver')}</InputLabel>
                <Select
                  id="driver"
                  name="driver"
                  value={formik.values.driver}
                  onChange={formik.handleChange}
                  label={t('global.driver')}
                >
                  <MenuItem value="">
                    <em>{t('mgmt.servApps.driver.none')}</em>
                  </MenuItem>
                  <MenuItem value="local">{t('mgmt.servApps.volumes.newVolume.driverSelection.localChoice')}</MenuItem>
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
            <Button onClick={() => setIsOpened(false)}>{t('global.cancelAction')}</Button>
            <LoadingButton
              onClick={formik.handleSubmit}
              loading={formik.isSubmitting}
            >
              {t('global.createAction')}
            </LoadingButton>
          </DialogActions>
        </FormikProvider>
      </Dialog>
      <Button
        fullWidth={fullWidth}
        onClick={() => setIsOpened(true)}
        startIcon={<PlusCircleOutlined />}
      >
        {t('mgmt.servApps.volumes.newVolumeTitle')}
      </Button>
    </>
  );
};

export default NewVolumeButton;