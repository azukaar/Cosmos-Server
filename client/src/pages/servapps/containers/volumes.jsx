import React from 'react';
import { Formik } from 'formik';
import { Button, Stack, Grid, MenuItem, TextField, IconButton, FormHelperText, CircularProgress, useTheme, Checkbox } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect }
  from '../../config/users/formShortcuts';
import { ApiOutlined, CheckCircleOutlined, CloseCircleOutlined, DeleteOutlined, PlusCircleOutlined } from '@ant-design/icons';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { NetworksColumns } from '../networks';
import NewNetworkButton from '../createNetwork';

const VolumeContainerSetup = ({ config, containerInfo, refresh }) => {
  const restartPolicies = [
    ['no', 'No Restart'],
    ['always', 'Always Restart'],
    ['on-failure', 'Restart On Failure'],
    ['unless-stopped', 'Restart Unless Stopped'],
  ];

  const [volumes, setVolumes] = React.useState([]);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  React.useEffect(() => {
    API.docker.networkList().then((res) => {
      setVolumes(res.data);
    });
  }, []);

  const refreshAll = () => {
    setVolumes(null);
    refresh().then(() => {
      API.docker.networkList().then((res) => {
        setVolumes(res.data);
      });
    });
  };

  return (<Stack spacing={2}>
    <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
      <Formik
        initialValues={{
          volumes: ([
            ...(containerInfo.HostConfig.Mounts || []),
            ...(containerInfo.HostConfig.Binds || []).map((bind) => {
              const [source, destination, mode] = bind.split(':');
              return {
                Type: 'bind',
                Source: source,
                Target: destination,
              }
            })
          ])
        }}
        validate={(values) => {
          const errors = {};
          // check unique
          const volumes = values.volumes.map((volume) => {
            return `${volume.Destination}`;
          });
          const unique = [...new Set(volumes)];
          if (unique.length !== volumes.length) {
            errors.submit = 'Mounts must have unique destinations';
          }
          return errors;
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          setSubmitting(true);
          const realvalues = {
            Volumes: values.volumes
          };
          return API.docker.updateContainer(containerInfo.Name.replace('/', ''), realvalues)
            .then((res) => {
              setStatus({ success: true });
              setSubmitting(false);
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
            <MainCard title={'Volume Mounts'}>
              <Grid container spacing={4}>
                <Grid item xs={12}>
                      {volumes && <PrettyTableView
                        data={formik.values.volumes}
                        onRowClick={() => { }}
                        getKey={(r) => r.Id}
                        fullWidth
                        buttons={[
                          <Button variant="outlined" color="primary" onClick={() => {
                            formik.setFieldValue('volumes', [...formik.values.volumes, {
                              Type: 'volume',
                              Name: '',
                              Driver: 'local',
                              Source: '',
                              Destination: '',
                              RW: true
                            }]);
                          }}>
                            New Mount Point
                          </Button>
                        ]}
                        columns={[
                          {
                            title: 'Type',
                            field: (r, k) => (
                              <div style={{fontWeight: 'bold', wordSpace: 'nowrap', overflow:'hidden', textOverflow:'ellipsis', maxWidth: '100px'}}>
                                <TextField
                                  className="px-2 my-2"
                                  variant="outlined"
                                  name='Type'
                                  id='Type'
                                  select
                                  value={r.Type}
                                  onChange={(e) => {
                                    const newVolumes = [...formik.values.volumes];
                                    newVolumes[k].Type = e.target.value;
                                    formik.setFieldValue('volumes', newVolumes);
                                  }}
                                >
                                  <MenuItem value="bind">Bind</MenuItem>
                                  <MenuItem value="volume">Volume</MenuItem>
                                </TextField>
                              </div>),
                          },
                          {
                            title: 'Source',
                            field: (r, k) => (
                              <div style={{fontWeight: 'bold', wordSpace: 'nowrap', overflow:'hidden', textOverflow:'ellipsis', maxWidth: '300px'}}>
                                {(r.Type == "bind") ? 
                                  <TextField
                                    className="px-2 my-2"
                                    variant="outlined"
                                    name='Source'
                                    id='Source'
                                    style={{minWidth: '200px'}}
                                    value={r.Source}
                                    onChange={(e) => {
                                      const newVolumes = [...formik.values.volumes];
                                      newVolumes[k].Source = e.target.value;
                                      formik.setFieldValue('volumes', newVolumes);
                                    }}
                                    /> : 
                                  <TextField
                                    className="px-2 my-2"
                                    variant="outlined"
                                    name='Source'
                                    id='Source'
                                    select
                                    style={{minWidth: '200px'}}
                                    value={r.Source}
                                    onChange={(e) => {
                                      const newVolumes = [...formik.values.volumes];
                                      newVolumes[k].Source = e.target.value;
                                      formik.setFieldValue('volumes', newVolumes);
                                    }}
                                  >
                                    {volumes.map((volume) => (
                                      <MenuItem key={volume.Id} value={volume.Name}>
                                        {volume.Name}
                                      </MenuItem>
                                    ))}
                                  </TextField>
                                }
                              </div>),
                          },
                          {
                            title: 'Target',
                            field: (r, k) => (
                              <div style={{fontWeight: 'bold', wordSpace: 'nowrap', overflow:'hidden', textOverflow:'ellipsis', maxWidth: '300px'}}>
                                <TextField
                                    className="px-2 my-2"
                                    variant="outlined"
                                    name='Target'
                                    id='Target'
                                    style={{minWidth: '200px'}}
                                    value={r.Target}
                                    onChange={(e) => {
                                      const newVolumes = [...formik.values.volumes];
                                      newVolumes[k].Target = e.target.value;
                                      formik.setFieldValue('volumes', newVolumes);
                                    }}
                                    />
                              </div>),
                          },
                          {
                            title: '',
                            field: (r) => {
                              console.log(r);
                              return (<Button
                                variant="outlined"
                                color="primary"
                                onClick={() => {
                                  const newVolumes = [...formik.values.volumes];
                                  newVolumes.splice(newVolumes.indexOf(r), 1);
                                  formik.setFieldValue('volumes', newVolumes);
                                }}>
                                Unmount
                              </Button>)
                            }
                          }
                        ]}
                      />}
                      {!volumes && (<div style={{ height: '550px' }}>
                        <center>
                          <br />
                          <CircularProgress />
                        </center>
                      </div>
                      )}
                </Grid>
                <Grid item xs={12}>
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
                      Update Volumes
                    </LoadingButton>
                  </Stack>
                </Grid>
              </Grid>
            </MainCard>
          </form>
        )}
      </Formik>
    </div>
  </Stack>);
};

export default VolumeContainerSetup;