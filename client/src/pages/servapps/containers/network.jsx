import React from 'react';
import { Formik } from 'formik';
import { Button, Stack, Grid, MenuItem, TextField, IconButton, FormHelperText, CircularProgress, useTheme } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect }
  from '../../config/users/formShortcuts';
import { ApiOutlined, CheckCircleOutlined, CloseCircleOutlined, DeleteOutlined, PlusCircleOutlined } from '@ant-design/icons';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { NetworksColumns } from '../networks';
import NewNetworkButton from '../createNetwork';

const NetworkContainerSetup = ({ config, containerInfo, refresh }) => {
  const restartPolicies = [
    ['no', 'No Restart'],
    ['always', 'Always Restart'],
    ['on-failure', 'Restart On Failure'],
    ['unless-stopped', 'Restart Unless Stopped'],
  ];

  const [networks, setNetworks] = React.useState([]);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';

  React.useEffect(() => {
    API.docker.networkList().then((res) => {
      setNetworks(res.data);
    });
  }, []);

  const refreshAll = () => {
    setNetworks(null);
    refresh().then(() => {
      API.docker.networkList().then((res) => {
        setNetworks(res.data);
      });
    });
  };

  const connect = (network) => {
    setNetworks(null);
    return API.docker.attachNetwork(containerInfo.Id, network).then(() => {
      refreshAll();
    });
  }

  const disconnect = (network) => {
    return API.docker.detachNetwork(containerInfo.Id, network).then(() => {
      refreshAll();
    });
  }

  return (<Stack spacing={2}>
    <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
      <Formik
        initialValues={{
          ports: Object.keys(containerInfo.NetworkSettings.Ports).map((port) => {
            return {
              port: port.split('/')[0],
              protocol: port.split('/')[1],
              hostPort: containerInfo.NetworkSettings.Ports[port] ?
                containerInfo.NetworkSettings.Ports[port][0].HostPort : '',
            };
          })
        }}
        validate={(values) => {
          const errors = {};
          // check unique
          const ports = values.ports.map((port) => {
            return `${port.port}/${port.protocol}`;
          });
          const unique = [...new Set(ports)];
          if (unique.length !== ports.length) {
            errors.submit = 'Ports must be unique';
          }
          return errors;
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          setSubmitting(true);
          const realvalues = {
            portBindings: {},
          };
          values.ports.forEach((port) => {
            if (port.hostPort) {
              realvalues.portBindings[`${port.port}/${port.protocol}`] = [
                {
                  HostPort: port.hostPort,
                }
              ];
            }
          });
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
            <Stack spacing={2}>
              <MainCard title={'Exposed Ports'}>
                <Grid container spacing={4}>
                  <Grid item xs={12}>
                    {formik.values.ports.map((port, idx) => (
                      <Grid container spacing={2} key={idx}>
                        <Grid item xs={4} style={{ padding: '20px 10px' }}>
                          <TextField
                            label="Container Port"
                            fullWidth
                            value={port.port}
                            onChange={(e) => {
                              const newports = [...formik.values.ports];
                              newports[idx].port = e.target.value;
                              formik.setFieldValue('ports', newports);
                            }}
                          />
                        </Grid>
                        <Grid item xs={4} style={{ padding: '20px 10px' }}>
                          <TextField
                            fullWidth
                            label="Host port"
                            value={'' + port.hostPort}
                            onChange={(e) => {
                              const newports = [...formik.values.ports];
                              newports[idx].hostPort = e.target.value;
                              formik.setFieldValue('ports', newports);
                            }}
                          />
                        </Grid>
                        <Grid item xs={3} style={{ padding: '20px 10px' }}>
                          <TextField
                            className="px-2 my-2"
                            variant="outlined"
                            name='protocol'
                            id='protocol'
                            select
                            value={port.protocol}
                            onChange={(e) => {
                              const newports = [...formik.values.ports];
                              newports[idx].protocol = e.target.value;
                              formik.setFieldValue('ports', newports);
                            }}
                          >
                            <MenuItem value="tcp">TCP</MenuItem>
                            <MenuItem value="udp">UDP</MenuItem>
                          </TextField>
                        </Grid>
                        <Grid item xs={1} style={{ padding: '20px 10px' }}>
                          <IconButton
                            fullWidth
                            variant="outlined"
                            color="error"
                            onClick={() => {
                              const newports = [...formik.values.ports];
                              newports.splice(idx, 1);
                              formik.setFieldValue('ports', newports);
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
                        const newports = [...formik.values.ports];
                        newports.push({
                          port: '',
                          protocol: 'tcp',
                          hostPort: '',
                        });
                        formik.setFieldValue('ports', newports);
                      }}
                    >
                      <PlusCircleOutlined />
                    </IconButton>
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
                        Update Ports
                      </LoadingButton>
                    </Stack>
                  </Grid>
                </Grid>
              </MainCard>
              <MainCard title={'Connected Networks'}>
                {networks && <PrettyTableView
                  data={networks}
                  buttons={[
                    <NewNetworkButton refresh={refreshAll} />,
                  ]}
                  onRowClick={() => { }}
                  getKey={(r) => r.Id}
                  columns={[
                    {
                      title: '',
                      field: (r) => {
                        const isConnected = containerInfo.NetworkSettings.Networks[r.Name];
                        // icon 
                        return isConnected ?
                          <ApiOutlined style={{ color: 'green' }} /> :
                          <ApiOutlined style={{ color: 'red' }} />
                      }
                    },
                    ...NetworksColumns(theme, isDark),
                    {
                      title: '',
                      field: (r) => {
                        const isConnected = containerInfo.NetworkSettings.Networks[r.Name];
                        return (<Button
                          variant="outlined"
                          color="primary"
                          onClick={() => {
                            isConnected ? disconnect(r.Name) : connect(r.Name);
                          }}>
                          {isConnected ? 'Disconnect' : 'Connect'}
                        </Button>)
                      }
                    }
                  ]}
                />}
                {!networks && (<div style={{ height: '550px' }}>
                  <center>
                    <br />
                    <CircularProgress />
                  </center>
                </div>
                )}
              </MainCard>
            </Stack>
          </form>
        )}
      </Formik>
    </div>
  </Stack>);
};

export default NetworkContainerSetup;