import * as React from 'react';
import isLoggedIn from '../../../isLoggedIn';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik, Field } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Checkbox,
  Divider,
  FormControlLabel,
  Grid,
  IconButton,
  InputAdornment,
  InputLabel,
  Link,
  OutlinedInput,
  Stack,
  Typography,
  FormHelperText,
  Collapse,
  TextField,
  MenuItem,
  
} from '@mui/material';
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import AnimateButton from '../../../components/@extended/AnimateButton';
import RestartModal from './restart';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined } from '@ant-design/icons';


const ConfigManagement = () => {
  isLoggedIn();
  const [config, setConfig] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  React.useEffect(() => {
    refresh();
  }, []);

  return <div style={{maxWidth: '1000px', margin: ''}}>
    <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
        refresh();
    }}>Refresh</Button><br /><br />
    {config && <>
      <RestartModal openModal={openModal} setOpenModal={setOpenModal} />
      <Formik
        initialValues={{
          MongoDB: config.MongoDB,
          LoggingLevel: config.LoggingLevel,

          Hostname: config.HTTPConfig.Hostname,
          GenerateMissingTLSCert: config.HTTPConfig.GenerateMissingTLSCert,
          GenerateMissingAuthCert: config.HTTPConfig.GenerateMissingAuthCert,
          HTTPPort: config.HTTPConfig.HTTPPort,
          HTTPSPort: config.HTTPConfig.HTTPSPort,
        }}
        validationSchema={Yup.object().shape({
          Hostname: Yup.string().max(255).required('Hostname is required'),
          MongoDB: Yup.string().max(512),
          LoggingLevel: Yup.string().max(255).required('Logging Level is required'),
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            let toSave = {
              ...config,
              MongoDB: values.MongoDB,
              LoggingLevel: values.LoggingLevel,
              HTTPConfig: {
                ...config.HTTPConfig,
                Hostname: values.Hostname,
                GenerateMissingTLSCert: values.GenerateMissingTLSCert,
                GenerateMissingAuthCert: values.GenerateMissingAuthCert,
                HTTPPort: values.HTTPPort,
                HTTPSPort: values.HTTPSPort,
              }
            }
            
            API.config.set(toSave).then((data) => {
              if (data.status == 'error') {
                setStatus({ success: false });
                if (data.code == 'UL001') {
                  setErrors({ submit: 'Wrong nickname or password. Try again or try resetting your password' });
                } else if (data.status == 'error') {
                  setErrors({ submit: 'Unexpected error. Try again later.' });
                }
                setSubmitting(false);
                return;
              } else {
                setStatus({ success: true });
                setSubmitting(false);
                setOpenModal(true);
              }
            })
          } catch (err) {
            setStatus({ success: false });
            setErrors({ submit: err.message });
            setSubmitting(false);
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
          <form noValidate onSubmit={handleSubmit}>
            <MainCard title="General">
              <Grid container spacing={3}>
                <Grid item xs={12}>
                  <Alert severity="info">This page allow you to edit the configuration file. Any Environment Variable overwritting configuration won't appear here.</Alert>
                </Grid>
                <Grid item xs={12}>
                  <Stack spacing={1}>
                    <InputLabel htmlFor="MongoDB-login">MongoDB connection string. It is advised to use Environment variable to store this securely instead. (Optional)</InputLabel>
                    <OutlinedInput
                      id="MongoDB-login"
                      type="password"
                      value={values.MongoDB}
                      name="MongoDB"
                      onBlur={handleBlur}
                      onChange={handleChange}
                      placeholder="MongoDB"
                      fullWidth
                      error={Boolean(touched.MongoDB && errors.MongoDB)}
                    />
                    {touched.MongoDB && errors.MongoDB && (
                      <FormHelperText error id="standard-weight-helper-text-MongoDB-login">
                        {errors.MongoDB}
                      </FormHelperText>
                    )}
                  </Stack>
                </Grid>

                <Grid item xs={12}>
                  <Stack spacing={1}>
                    <InputLabel htmlFor="LoggingLevel-login">Level of logging (Default: INFO)</InputLabel>
                    <TextField
                      className="px-2 my-2"
                      variant="outlined"
                      name="LoggingLevel"
                      id="LoggingLevel"
                      select
                      value={values.LoggingLevel}
                      onChange={handleChange}
                      error={
                        touched.LoggingLevel &&
                        Boolean(errors.LoggingLevel)
                      }
                      helperText={
                        touched.LoggingLevel && errors.LoggingLevel
                      }
                    >
                      <MenuItem key={"DEBUG"} value={"DEBUG"}>
                        DEBUG
                      </MenuItem>
                      <MenuItem key={"INFO"} value={"INFO"}>
                        INFO
                      </MenuItem>
                      <MenuItem key={"WARNING"} value={"WARNING"}>
                        WARNING
                      </MenuItem>
                      <MenuItem key={"ERROR"} value={"ERROR"}>
                        ERROR
                      </MenuItem>
                    </TextField>
                  </Stack>
                </Grid>
              </Grid>
            </MainCard>

            <br /><br />

            <MainCard title="HTTP">
              <Grid container spacing={3}>
                <Grid item xs={12}>
                  <Stack spacing={1}>
                    <InputLabel htmlFor="Hostname-login">Hostname: This will be used to restrict access to your Cosmos Server (Default: 0.0.0.0)</InputLabel>
                    <OutlinedInput
                      id="Hostname-login"
                      type="text"
                      value={values.Hostname}
                      name="Hostname"
                      onBlur={handleBlur}
                      onChange={handleChange}
                      placeholder="Hostname"
                      fullWidth
                      error={Boolean(touched.Hostname && errors.Hostname)}
                    />
                    {touched.Hostname && errors.Hostname && (
                      <FormHelperText error id="standard-weight-helper-text-Hostname-login">
                        {errors.Hostname}
                      </FormHelperText>
                    )}
                  </Stack>
                </Grid>

                <Grid item xs={12}>
                  <Stack spacing={1}>
                    <InputLabel htmlFor="HTTPPort-login">HTTP Port (Default: 80)</InputLabel>
                    <OutlinedInput
                      id="HTTPPort-login"
                      type="text"
                      value={values.HTTPPort}
                      name="HTTPPort"
                      onBlur={handleBlur}
                      onChange={handleChange}
                      placeholder="HTTPPort"
                      fullWidth
                      error={Boolean(touched.HTTPPort && errors.HTTPPort)}
                    />
                    {touched.HTTPPort && errors.HTTPPort && (
                      <FormHelperText error id="standard-weight-helper-text-HTTPPort-login">
                        {errors.HTTPPort}
                      </FormHelperText>
                    )}
                  </Stack>
                </Grid>

                <Grid item xs={12}>
                  <Stack spacing={1}>
                    <InputLabel htmlFor="HTTPSPort-login">HTTPS Port (Default: 443)</InputLabel>
                    <OutlinedInput
                      id="HTTPSPort-login"
                      type="text"
                      value={values.HTTPSPort}
                      name="HTTPSPort"
                      onBlur={handleBlur}
                      onChange={handleChange}
                      placeholder="HTTPSPort"
                      fullWidth
                      error={Boolean(touched.HTTPSPort && errors.HTTPSPort)}
                    />
                    {touched.HTTPSPort && errors.HTTPSPort && (
                      <FormHelperText error id="standard-weight-helper-text-HTTPSPort-login">
                        {errors.HTTPSPort}
                      </FormHelperText>
                    )}
                  </Stack>
                </Grid>
                </Grid>
                </MainCard>
                <br /><br />
                <MainCard title="Security Certificates">
                <Grid container spacing={3}>
                <Grid item xs={12}>
                  <Alert severity="info">For security reasons, It is not possible to remotely change the Private keys of any certificates on your instance. It is advised to manually edit the config file, or better, use Environment Variables to store them.</Alert>
                </Grid>
                <Grid item xs={12}>
                  <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                    <Field
                      type="checkbox"
                      name="GenerateMissingTLSCert"
                      as={FormControlLabel}
                      control={<Checkbox size="large" />}
                      label="Generate missing HTTPS Certificates automatically (Default: true)"
                    />
                  </Stack>
                </Grid>

                <Grid item xs={12}>
                  <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                    <Field
                      type="checkbox"
                      name="GenerateMissingAuthCert"
                      as={FormControlLabel}
                      control={<Checkbox size="large" />}
                      label="Generate missing Authentication Certificates automatically (Default: true)"
                    />
                  </Stack>
                </Grid>

                <Grid item xs={12}>
                  <h4>Authentication Public Key</h4>
                  <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                    <pre className='code'>
                      {config.HTTPConfig.AuthPublicKey}
                    </pre>
                  </Stack>
                </Grid>

                <Grid item xs={12}>
                  <h4>Root HTTPS Public Key</h4>
                  <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                    <pre className='code'>
                      {config.HTTPConfig.TLSCert}
                    </pre>
                  </Stack>
                </Grid>

              </Grid>
            </MainCard>

            <br /><br />

            <MainCard>
              {errors.submit && (
                <Grid item xs={12}>
                  <FormHelperText error>{errors.submit}</FormHelperText>
                </Grid>
              )}
              <Grid item xs={12}>
                <AnimateButton>
                  <Button
                    disableElevation
                    disabled={isSubmitting}
                    fullWidth
                    size="large"
                    type="submit"
                    variant="contained"
                    color="primary"
                  >
                    Save
                  </Button>
                </AnimateButton>
              </Grid>
            </MainCard>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default ConfigManagement;