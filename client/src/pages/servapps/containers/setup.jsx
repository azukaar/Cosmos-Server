import React from 'react';
import { Formik } from 'formik';
import { Button, Stack, Grid, MenuItem, TextField, IconButton, FormHelperText } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect }
   from '../../config/users/formShortcuts';
import { DeleteOutlined, PlusCircleOutlined } from '@ant-design/icons';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';

const DockerContainerSetup = ({config, containerInfo, refresh}) => {
  const restartPolicies = [
    ['no', 'No Restart'],
    ['always', 'Always Restart'],
    ['on-failure', 'Restart On Failure'],
    ['unless-stopped', 'Restart Unless Stopped'],
  ];

  return (
    <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
      <Formik
        initialValues={{
          image: containerInfo.Config.Image,
          restartPolicy: containerInfo.HostConfig.RestartPolicy.Name,
          envVars: containerInfo.Config.Env.map((envVar) => {
            const [key, value] = envVar.split('=');
            return { key, value }; 
          }),
          labels: Object.keys(containerInfo.Config.Labels).map((key) => {
            return { key, value: containerInfo.Config.Labels[key] };
          }),
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
          return errors;
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          setSubmitting(true);
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
            <Stack spacing={2}>
              <MainCard title={'Docker Container Setup'}>
                <Grid container spacing={4}>
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
                  <CosmosFormDivider title={'Environment Variables'} />
                  <Grid item xs={12}>
                    {formik.values.envVars.map((envVar, idx) => (
                      <Grid container spacing={2} key={idx}>
                        <Grid item xs={5} style={{padding: '20px 10px'}}>
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
                        <Grid item xs={6} style={{padding: '20px 10px'}}>
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
                        <Grid item xs={1} style={{padding: '20px 10px'}}>
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
                      <Grid container spacing={2} key={idx}>
                        <Grid item xs={5} style={{padding: '20px 10px'}}>
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
                        <Grid item xs={6} style={{padding: '20px 10px'}}>
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
                        <Grid item xs={1} style={{padding: '20px 10px'}}>
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
              <MainCard>
                <Stack direction="column" spacing={2}>
                  {formik.errors.submit && (
                    <Grid item xs={12}>
                      <FormHelperText error>{formik.errors.submit}</FormHelperText>
                    </Grid>
                  )}
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
                    Update
                  </LoadingButton>
                </Stack>
              </MainCard>
            </Stack>
          </form>
        )}
      </Formik>
    </div>);
};

export default DockerContainerSetup;