// material-ui
import * as React from 'react';
import { Alert, Button, Checkbox, FormControl, FormHelperText, Grid, InputLabel, MenuItem, Select, Stack, TextField } from '@mui/material';
import { PlusCircleFilled, PlusCircleOutlined, WarningOutlined } from '@ant-design/icons';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import CircularProgress from '@mui/material/CircularProgress';
import { useEffect, useState } from 'react';
import { LoadingButton } from '@mui/lab';
import { FormikProvider, useFormik } from 'formik';
import * as Yup from 'yup';

import * as API from '../../api';
import { CosmosCheckbox } from '../config/users/formShortcuts';
import ResponsiveButton from '../../components/responseiveButton';
import { CosmosContainerPicker } from '../config/users/containerPicker';
import { randomString } from '../../utils/indexs';

const LinkContainersButton = ({ fullWidth, refresh, originContainer, newContainer, OnConnect }) => {
  const [isOpened, setIsOpened] = useState(false);
  const formik = useFormik({
    initialValues: {
      container: '',
    },
    validationSchema: Yup.object({
      container: Yup.string().required('Required'),
    }),
    validate: (values) => {
      const errors = {};
      if (values.container === originContainer) {
        errors.container = 'Cannot link to itself';
      }
      return errors;
    },
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      let name = 'link-' + originContainer + '-' + values.container + '-' + randomString(3);
      return API.docker.createNetwork({
        name,
        driver : 'bridge',
        attachCosmos: false,
      })
        .then((res) => {
          return Promise.all([
            newContainer ? OnConnect(name) : API.docker.attachNetwork(originContainer, name),
            API.docker.attachNetwork(values.container, name),
          ])
          .then(() => {
            setStatus({ success: true });
            setSubmitting(false);
            setIsOpened(false);
            refresh && refresh();
          })
          .catch((err) => {
            setStatus({ success: false });
            setErrors({ submit: err.message });
            setSubmitting(false);
          });
        })
    },
  });

  return (
    <>
      <Dialog open={isOpened} onClose={() => setIsOpened(false)}>
        <FormikProvider value={formik}>
          <DialogTitle>Link with container</DialogTitle>
          <DialogContent>
            <DialogContentText>
              <form onSubmit={formik.handleSubmit}>
                <Stack spacing={2} width={'300px'} style={{ marginTop: '10px' }}>
                  <CosmosContainerPicker
                    formik={formik}
                    onTargetChange={(_, name) => {
                      formik.setFieldValue('container', name);
                    }}
                    nameOnly
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
              loading={formik.isSubmitting}>
              Link Containers
            </LoadingButton>
          </DialogActions>
        </FormikProvider>
      </Dialog>
      <ResponsiveButton
        fullWidth={fullWidth}
        onClick={() => setIsOpened(true)}
        startIcon={<PlusCircleOutlined />}
      >
        Link Containers
      </ResponsiveButton>
    </>
  );
};

export default LinkContainersButton;