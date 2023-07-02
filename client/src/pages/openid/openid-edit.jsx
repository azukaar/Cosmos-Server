// material-ui
import { AppstoreAddOutlined, PlusCircleOutlined, ReloadOutlined, SearchOutlined, SettingOutlined } from '@ant-design/icons';
import { Alert, Badge, Button, Card, Checkbox, Chip, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, Divider, Input, InputAdornment, TextField, Tooltip, Typography } from '@mui/material';
import Grid2 from '@mui/material/Unstable_Grid2/Grid2';
import { Stack } from '@mui/system';
import { useEffect, useState } from 'react';
import Paper from '@mui/material/Paper';
import { styled } from '@mui/material/styles';

import * as API from '../../api';
import IsLoggedIn from '../../isLoggedIn';
import RestartModal from '../config/users/restart';
import RouteManagement from '../config/routes/routeman';
import { ValidateRoute, getFaviconURL, sanitizeRoute } from '../../utils/routes';
import HostChip from '../../components/hostChip';
import { Formik, useFormik } from 'formik';
import * as yup from 'yup';

const OpenIdEditModal = ({ clientId, openNewModal, setOpenNewModal, config, onSubmit }) => {
  const [submitErrors, setSubmitErrors] = useState([]);
  const [newRoute, setNewRoute] = useState(null);

  const clientConfig = config.OpenIDClients && Object.values(config.OpenIDClients).find((c) => c.id === clientId);

  return <>
    <Dialog open={openNewModal} onClose={() => setOpenNewModal(false)}>
      <Formik
        initialValues={{
          id: clientConfig ? clientConfig.id : '',
          redirect: clientConfig ? clientConfig.redirect : '',
        }}
        validationSchema={yup.object({
          id: yup.string().required('Required'),
          redirect: yup.string().required('Required'),
        })}
        onSubmit={(values) => {
          onSubmit && onSubmit(values);
        }}
      >
        {(formik) => (
          <form onSubmit={formik.handleSubmit}>
            <DialogTitle>{clientId ? clientId : "New client"}</DialogTitle>
            {openNewModal && <>
              <DialogContent>
                <DialogContentText>
                  <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                    <TextField
                      fullWidth
                      id="id"
                      name="id"
                      label="ID"
                      value={formik.values.id}
                      onChange={formik.handleChange}
                      error={formik.touched.id && Boolean(formik.errors.id)}
                      helperText={formik.touched.id && formik.errors.id}
                    />
                    <TextField
                      fullWidth
                      id="redirect"
                      name="redirect"
                      label="Redirect"
                      value={formik.values.redirect}
                      onChange={formik.handleChange}
                      error={formik.touched.redirect && Boolean(formik.errors.redirect)}
                      helperText={formik.touched.redirect && formik.errors.redirect}
                    />
                  </Stack>
                </DialogContentText>
              </DialogContent>
              <DialogActions>
                {submitErrors && submitErrors.length > 0 && <Stack spacing={2} direction={"column"}>
                  <Alert severity="error">{submitErrors.map((err) => {
                    return <div>{err}</div>
                  })}</Alert>
                </Stack>}
                <Button onClick={() => setOpenNewModal(false)}>Cancel</Button>
                <Button color="primary" variant="contained" type="submit" onClick={() => {
                  formik.handleSubmit();
                }}>{clientId ? "Edit" : "Create"}</Button>
              </DialogActions>
            </>}
          </form>

        )}
      </Formik>
    </Dialog>
  </>;
}

export default OpenIdEditModal;