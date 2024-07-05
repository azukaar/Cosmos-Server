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
import { useTranslation } from 'react-i18next';


const NewNetworkButton = ({ fullWidth, refresh }) => {
  const { t } = useTranslation();
  const [isOpened, setIsOpened] = React.useState(false);

  const formik = useFormik({
    initialValues: {
      name: '',
      driver: 'bridge',
      attachCosmos: false,
      parentInterface: '',
      subnet: '',
    },
    validationSchema: Yup.object({
      name: Yup.string().required('Required'),
      driver: Yup.string().required('Required'),
      parentInterface: Yup.string().when('driver', {
        is: 'macvlan',
        then: Yup.string().required('Parent interface is required for MACVLAN')
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
          <DialogTitle>{t('NewNetwork')}</DialogTitle>
          <DialogContent>
            <DialogContentText>
              <form onSubmit={formik.handleSubmit}>
                <Stack spacing={2} width={'300px'} style={{ marginTop: '10px' }}>
                  <TextField
                    fullWidth
                    id="name"
                    name="name"
                    label={t('Name')}
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
                      label={t('Driver')}
                    >
                      <MenuItem value="">
                        <em>None</em>
                      </MenuItem>
                      <MenuItem value="bridge">{t('Bridge')}</MenuItem>
                      <MenuItem value="host">{t('Host')}</MenuItem>
                      <MenuItem value="overlay">{t('Overlay')}</MenuItem>
                      <MenuItem value="macvlan">{t('MACVLAN')}</MenuItem>
                    </Select>
                  </FormControl>
                  
                  <TextField
                    fullWidth
                    id="subnet"
                    name="subnet"
                    label={t('Subnet')}
                    value={formik.values.subnet}
                    onChange={formik.handleChange}
                    error={formik.touched.subnet && Boolean(formik.errors.subnet)}
                    helperText={formik.touched.subnet && formik.errors.subnet}
                    style={{ marginBottom: '16px' }}
                  />

                  {formik.values.driver === 'macvlan' && (
                    <TextField
                      fullWidth
                      id="parentInterface"
                      name="parentInterface"
                      label={t('ParentInterface')}
                      value={formik.values.parentInterface}
                      onChange={formik.handleChange}
                      error={formik.touched.parentInterface && Boolean(formik.errors.parentInterface)}
                      helperText={formik.touched.parentInterface && formik.errors.parentInterface}
                      style={{ marginBottom: '16px' }}
                    />
                  )}

                  <CosmosCheckbox
                    name="attachCosmos"
                    label={t('AttachToCosmos')}
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
            <Button onClick={() => setIsOpened(false)}>{t('Cancel')}</Button>
            <LoadingButton
              disabled={formik.errors.submit}
              onClick={formik.handleSubmit}
              loading={formik.isSubmitting}
            >
              {t('Create')}
            </LoadingButton>
          </DialogActions>
        </FormikProvider>
      </Dialog>
      <ResponsiveButton
        fullWidth={fullWidth}
        onClick={() => setIsOpened(true)}
        startIcon={<PlusCircleOutlined />}
      >
        {t('NewNetwork')}
      </ResponsiveButton>
    </>
  );
};

export default NewNetworkButton;
