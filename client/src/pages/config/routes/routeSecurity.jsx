import * as React from 'react';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik } from 'formik';
import {
  Alert,
  Button,
  Grid,
  Stack,

} from '@mui/material';
import RestartModal from '../users/restart';
import { CosmosCheckbox, CosmosInputText } from '../users/formShortcuts';
import { snackit } from '../../../api/wrap';

const RouteSecurity = ({ routeConfig }) => {
  const [openModal, setOpenModal] = React.useState(false);

  return <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} />

    {routeConfig && <>
      <Formik
        initialValues={{
          AuthEnabled: routeConfig.AuthEnabled,
          Timeout: routeConfig.Timeout,
          ThrottlePerMinute: routeConfig.ThrottlePerMinute,
          CORSOrigin: routeConfig.CORSOrigin,
          MaxBandwith: routeConfig.MaxBandwith,
          _SmartShield_Enabled: (routeConfig.SmartShield ? routeConfig.SmartShield.Enabled : false),
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          const fullValues = {
            ...routeConfig,
            ...values,
          }

          if(!fullValues.SmartShield) {
            fullValues.SmartShield = {};
          }
          fullValues.SmartShield.Enabled = values._SmartShield_Enabled;
          delete fullValues._SmartShield_Enabled;

          API.config.replaceRoute(routeConfig.Name, fullValues).then((res) => {
            if (res.status == "OK") {
              setStatus({ success: true });
              snackit('Route updated successfully', 'success');
              setSubmitting(false);
              setOpenModal(true);
            } else {
              setStatus({ success: false });
              setErrors({ submit: res.status });
              setSubmitting(false);
            }
          });
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={2}>
              <MainCard name={routeConfig.Name} title={'Security'}>
                <Grid container spacing={2}>
                    <Grid item xs={12}>
                      <Alert color='info'>Additional security settings. MFA and Captcha are not yet implemented.</Alert>
                    </Grid>

                    <CosmosCheckbox
                      name="AuthEnabled"
                      label="Authentication Required"
                      formik={formik}
                    />

                    <CosmosCheckbox
                      name="_SmartShield_Enabled"
                      label="Smart Shield Protection"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="Timeout"
                      label="Timeout in milliseconds (0 for no timeout, at least 30000 or less recommended)"
                      placeholder="Timeout"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="MaxBandwith"
                      label="Maximum Bandwith limit per user in bytes per seconds (0 for no limit)"
                      placeholder="Maximum Bandwith"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="ThrottlePerMinute"
                      label="Maximum number of requests Per Minute (0 for no limit, at least 100 or less recommended)"
                      placeholder="Throttle Per Minute"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="CORSOrigin"
                      label="Custom CORS Origin (Recommended to leave blank)"
                      placeholder="CORS Origin"
                      formik={formik}
                    />
                </Grid>
              </MainCard>
              <MainCard ><Button
                fullWidth
                disableElevation
                size="large"
                type="submit"
                variant="contained"
                color="primary"
              >
                Save
              </Button></MainCard>
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteSecurity;