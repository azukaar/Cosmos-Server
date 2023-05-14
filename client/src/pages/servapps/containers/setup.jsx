import React from 'react';
import { Field, Formik } from 'formik';
import { Button, Stack, Grid, MenuItem, TextField, IconButton, FormHelperText, useMediaQuery, useTheme, Alert, FormControlLabel, Checkbox } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect }
   from '../../config/users/formShortcuts';
import { DeleteOutlined, PlusCircleOutlined } from '@ant-design/icons';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import LogsInModal from '../../../components/logsInModal';

const containerInfoFrom = (values) => {
  const labels = {};
  values.labels.forEach((label) => {
    labels[label.key] = label.value;
  });
  const envVars = values.envVars.map((envVar) => {
    return `${envVar.key}=${envVar.value}`;
  });
  const realvalues = {
    ...values,
    envVars: envVars,
    labels: labels,
  };
  realvalues.interactive = realvalues.interactive ? 2 : 1;
  
  return realvalues;
}

const DockerContainerSetup = ({config, containerInfo, OnChange, refresh, newContainer, OnForceSecure}) => {
  const restartPolicies = [
    ['no', 'No Restart'],
    ['always', 'Always Restart'],
    ['on-failure', 'Restart On Failure'],
    ['unless-stopped', 'Restart Unless Stopped'],
  ];
  const [pullRequest, setPullRequest] = React.useState(null);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const padding = isMobile ? '6px 4px' : '12px 10px';
  const [latestImage, setLatestImage] = React.useState(containerInfo.Config.Image);

  return (
    <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
      <Formik
        initialValues={{
          name: containerInfo.Name.replace('/', ''),
          image: containerInfo.Config.Image,
          restartPolicy: containerInfo.HostConfig.RestartPolicy.Name,
          envVars: containerInfo.Config.Env.map((envVar) => {
            const [key, value] = envVar.split('=');
            return { key, value }; 
          }),
          labels: Object.keys(containerInfo.Config.Labels).map((key) => {
            return { key, value: containerInfo.Config.Labels[key] };
          }),
          interactive: containerInfo.Config.Tty && containerInfo.Config.OpenStdin,
        }}
        validate={(values) => {
          const errors = {};
          if (!values.image) {
            errors.image = 'Required';
          }
          // env keys and labels key mustbe unique
          const envKeys = values.envVars.map((envVar) => envVar.key);
          const labelKeys = values.labels.map((label) => label.key);
          const uniqueEnvKeysKeys = [...new Set(envKeys)];
          const uniqueLabelKeys = [...new Set(labelKeys)];
          if (uniqueEnvKeysKeys.length !== envKeys.length) {
            errors.submit = 'Environment Variables must be unique';
          }
          if (uniqueLabelKeys.length !== labelKeys.length) {
            errors.submit = 'Labels must be unique';
          }
          OnChange && OnChange(containerInfoFrom(values));
          return errors;
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          if(values.image !== latestImage) {
            setPullRequest(() => ((cb) => API.docker.pullImage(values.image,cb, true)));
            return;
          }

          if(newContainer) return false;
          delete values.name;

          setSubmitting(true);

          let realvalues = containerInfoFrom(values);

          return API.docker.updateContainer(containerInfo.Name.replace('/', ''), realvalues)
            .then((res) => {
              setStatus({ success: true });
              setSubmitting(false);
              refresh && refresh();
            }
          ).catch((err) => {
            setStatus({ success: false });
            setErrors({ submit: err.message });
            setSubmitting(false);
          });
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <LogsInModal
              request={pullRequest}
              title="Pulling New Image..."
              OnSuccess={() => {
                setPullRequest(null);
                setLatestImage(formik.values.image);
              }}
            />
            <Stack spacing={2}>
              <MainCard title={'Docker Container Setup'}>
              {containerInfo.State && containerInfo.State.Status !== 'running' && (
              <Alert severity="warning" style={{ marginBottom: '15px' }}>
                  This container is not running. Editing any settings will cause the container to start again.
                </Alert>
              )}
                <Grid container spacing={4}>
                    {newContainer && <CosmosInputText
                      name="name"
                      label="Name"
                      placeholder="Name"
                      formik={formik}
                      />}
                    <CosmosInputText
                      name="image"
                      label="Image"
                      placeholder="Image"
                      formik={formik}
                      />
                    <CosmosSelect
                      name="restartPolicy"
                      label="Restart Policy"
                      placeholder="Restart Policy"
                      options={restartPolicies}
                      formik={formik}
                      />
                    <CosmosCheckbox
                      name="interactive"
                      label="Interactive Mode"
                      formik={formik}
                    />
                    {OnForceSecure && <Grid item xs={12}>
                    <Checkbox
                        type="checkbox"
                        as={FormControlLabel}
                        control={<Checkbox size="large" />}
                        label={'Force secure container'}
                        checked={
                          containerInfo.Config.Labels.hasOwnProperty('cosmos-force-network-secured') &&
                          containerInfo.Config.Labels['cosmos-force-network-secured'] === 'true'
                        }
                        onChange={(e) => {
                          OnForceSecure(e.target.checked);
                        }}
                      />
                  </Grid>}

                  <CosmosFormDivider title={'Environment Variables'} />
                  <Grid item xs={12}>
                    {formik.values.envVars.map((envVar, idx) => (
                      <Grid container key={idx}>
                        <Grid item xs={5} style={{padding}}>
                          <TextField
                            label="Key"
                            fullWidth
                            value={envVar.key}
                            onChange={(e) => {
                              const newEnvVars = [...formik.values.envVars];
                              newEnvVars[idx].key = e.target.value;
                              formik.setFieldValue('envVars', newEnvVars);
                            }}
                          />
                        </Grid>
                        <Grid item xs={6} style={{padding}}>
                          <TextField
                            fullWidth
                            label="Value"
                            value={envVar.value}
                            onChange={(e) => {
                              const newEnvVars = [...formik.values.envVars];
                              newEnvVars[idx].value = e.target.value;
                              formik.setFieldValue('envVars', newEnvVars);
                            }}
                          />
                        </Grid>
                        <Grid item xs={1} style={{padding}}>
                          <IconButton
                            fullWidth
                            variant="outlined"
                            color="error"
                            onClick={() => {
                              const newEnvVars = [...formik.values.envVars];
                              newEnvVars.splice(idx, 1);
                              formik.setFieldValue('envVars', newEnvVars);
                            }}
                          >
                            <DeleteOutlined />
                          </IconButton>
                        </Grid>
                      </Grid>
                    ))}
                    <IconButton
                      fullWidth
                      variant="outlined"
                      color="primary"
                      size='large'
                      onClick={() => {
                        const newEnvVars = [...formik.values.envVars];
                        newEnvVars.push({ key: '', value: '' });
                        formik.setFieldValue('envVars', newEnvVars);
                      }}
                    >
                      <PlusCircleOutlined />
                    </IconButton>
                  </Grid>
                  <CosmosFormDivider title={'Labels'} />
                  <Grid item xs={12}>
                    {formik.values.labels.map((label, idx) => (
                      <Grid container key={idx}>
                        <Grid item xs={5} style={{padding}}>
                          <TextField
                            fullWidth
                            label="Key"
                            value={label.key}
                            onChange={(e) => {
                              const newLabels = [...formik.values.labels];
                              newLabels[idx].key = e.target.value;
                              formik.setFieldValue('labels', newLabels);
                            }}
                          />
                        </Grid>
                        <Grid item xs={6} style={{padding}}>
                          <TextField
                            label="Value"
                            fullWidth
                            value={label.value}
                            onChange={(e) => {
                              const newLabels = [...formik.values.labels];
                              newLabels[idx].value = e.target.value;
                              formik.setFieldValue('labels', newLabels);
                            }}
                          />
                        </Grid>
                        <Grid item xs={1} style={{padding}}>
                          <IconButton
                            fullWidth
                            variant="outlined"
                            color="error"
                            onClick={() => {
                              const newLabels = [...formik.values.labels];
                              newLabels.splice(idx, 1);
                              formik.setFieldValue('labels', newLabels);
                            }}
                          >
                            <DeleteOutlined />
                          </IconButton>
                        </Grid>
                      </Grid>
                    ))}
                    <IconButton
                      fullWidth
                      variant="outlined"
                      color="primary"
                      size='large'
                      onClick={() => {
                        const newLabels = [...formik.values.labels];
                        newLabels.push({ key: '', value: '' });
                        formik.setFieldValue('labels', newLabels);
                      }}
                    >
                      <PlusCircleOutlined />
                    </IconButton>
                  </Grid>
                </Grid>
              </MainCard>
              {!newContainer && <MainCard>
                <Stack direction="column" spacing={2}>
                  {formik.errors.submit && (
                    <Grid item xs={12}>
                      <FormHelperText error>{formik.errors.submit}</FormHelperText>
                    </Grid>
                  )}
                  {formik.values.image !== latestImage && <Alert severity="warning" style={{ marginBottom: '15px' }}>
                    You have updated the image. Clicking the button below will pull the new image, and then only can you update the container.
                  </Alert>}
                  <LoadingButton
                    fullWidth
                    disableElevation
                    disabled={formik.errors.submit}
                    loading={formik.isSubmitting}
                    size="large"
                    type="submit"
                    variant="contained"
                    color="primary"
                  >
                    {formik.values.image !== latestImage ? 'Pull New Image' : 'Update Container'}
                  </LoadingButton>
                </Stack>
              </MainCard>}
            </Stack>
          </form>
        )}
      </Formik>
    </div>);
};

export default DockerContainerSetup;