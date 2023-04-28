import * as React from 'react';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Grid,
  Stack,
  FormHelperText,
} from '@mui/material';
import RestartModal from '../users/restart';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../users/formShortcuts';
import { CosmosContainerPicker } from '../users/containerPicker';
import { snackit } from '../../../api/wrap';

export const ValidateRoute = Yup.object().shape({
  Name: Yup.string().required('Name is required'),
  Mode: Yup.string().required('Mode is required'),
  Target: Yup.string().required('Target is required').when('Mode', {
    is: 'SERVAPP',
    then: Yup.string().matches(/:[0-9]+$/, 'Invalid Target, must have a port'),
  }),

  Host: Yup.string().when('UseHost', {
    is: true,
    then: Yup.string().required('Host is required')
      .matches(/[\.|\:]/, 'Host must be full domain ([sub.]domain.com) or an IP')
  }),

  PathPrefix: Yup.string().when('UsePathPrefix', {
    is: true,
    then: Yup.string().required('Path Prefix is required').matches(/^\//, 'Path Prefix must start with / (e.g. /api). Do not include a domain/subdomain in it, use the Host for this.')
  }),

  UseHost: Yup.boolean().when('UsePathPrefix',
    {
      is: false,
      then: Yup.boolean().oneOf([true], 'Source must at least be either Host or Path Prefix')
    }),
})

const Hide = ({ children, h }) => {
  return h ? <div style={{ display: 'none' }}>
    {children}
  </div> : <>{children}</>
}

const RouteManagement = ({ routeConfig, TargetContainer, noControls = false, lockTarget = false, title, setRouteConfig, submitButton = false }) => {
  const [openModal, setOpenModal] = React.useState(false);

  return <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} />

    {routeConfig && <>
      <Formik
        initialValues={{
          Name: routeConfig.Name,
          Description: routeConfig.Description,
          Mode: routeConfig.Mode || "SERVAPP",
          Target: routeConfig.Target,
          UseHost: routeConfig.UseHost,
          Host: routeConfig.Host,
        }}
        validationSchema={ValidateRoute}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          if(!submitButton) {
            return false;
          } else {
            const fullValues = {
              ...routeConfig,
              ...values,
            }
            API.config.replaceRoute(routeConfig.Name, fullValues).then((res) => {
              if (res.status == "OK") {
                setStatus({ success: true });
                snackit('Route updated successfully', 'success')
                setSubmitting(false);
                setOpenModal(true);
              } else {
                setStatus({ success: false });
                setErrors({ submit: res.status });
                setSubmitting(false);
              }
            });
          }
        }}
        validate={(values) => {
          setRouteConfig && setRouteConfig(values);
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={2}>
              <MainCard name={routeConfig.Name} title={
                noControls ? 'New URL' :
                  <div>{title || routeConfig.Name}</div>
              }>
                <Grid container spacing={2}>
                  {formik.errors.submit && (
                    <Grid item xs={12}>
                      <FormHelperText error>{formik.errors.submit}</FormHelperText>
                    </Grid>
                  )}

                  <CosmosInputText
                    name="Name"
                    label="Name"
                    placeholder="Name"
                    formik={formik}
                  />

                  <CosmosInputText
                    name="Description"
                    label="Description"
                    placeholder="Description"
                    formik={formik}
                  />

                  <Hide h={lockTarget}>
                    <CosmosFormDivider title={'Target Type'} />
                    <Grid item xs={12}>
                      <Alert color='info'>What are you trying to access with this route?</Alert>
                    </Grid>

                    <CosmosSelect
                      name="Mode"
                      label="Mode"
                      formik={formik}
                      disabled={lockTarget}
                      options={[
                        ["SERVAPP", "ServApp - Docker Container"],
                        ["PROXY", "Proxy"],
                        ["STATIC", "Static Folder"],
                        ["SPA", "Single Page Application"],
                        ["REDIRECT", "Redirection"]
                      ]}
                    />
                  </Hide>
                  <CosmosFormDivider title={'Target Settings'} />

                  {
                    (formik.values.Mode === "SERVAPP") ?
                      <CosmosContainerPicker
                        formik={formik}
                        lockTarget={lockTarget}
                        TargetContainer={TargetContainer}
                        onTargetChange={() => {
                          setRouteConfig && setRouteConfig(formik.values);
                        }}
                      />
                      : <CosmosInputText
                        name="Target"
                        label={formik.values.Mode == "PROXY" ? "Target URL" : "Target Folder Path"}
                        placeholder={formik.values.Mode == "PROXY" ? "localhost:8080" : "/path/to/my/app"}
                        formik={formik}
                      />
                  }

                  <CosmosFormDivider title={'Source'} />

                  <Grid item xs={12}>
                    <Alert color='info'>What URL do you want to access your target from?</Alert>
                  </Grid>

                  <CosmosCheckbox
                    name="UseHost"
                    label="Use Host"
                    formik={formik}
                  />

                  {formik.values.UseHost && <CosmosInputText
                    name="Host"
                    label="Host"
                    placeholder="Host"
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}

                  <CosmosCheckbox
                    name="UsePathPrefix"
                    label="Use Path Prefix"
                    formik={formik}
                  />

                  {formik.values.UsePathPrefix && <CosmosInputText
                    name="PathPrefix"
                    label="Path Prefix"
                    placeholder="Path Prefix"
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}

                  {formik.values.UsePathPrefix && <CosmosCheckbox
                    name="StripPathPrefix"
                    label="Strip Path Prefix"
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}
                </Grid>
              </MainCard>
              {submitButton && <MainCard ><Button
                fullWidth
                disableElevation
                size="large"
                type="submit"
                variant="contained"
                color="primary"
              >
                Save
              </Button></MainCard>}
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteManagement;