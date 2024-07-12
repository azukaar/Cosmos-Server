import React from 'react';
import { Formik } from 'formik';
import { Button, Stack, Grid, MenuItem, TextField, IconButton, FormHelperText, CircularProgress, useTheme, Alert, useMediaQuery } from '@mui/material';
import MainCard from '../../../components/MainCard';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect }
  from '../../config/users/formShortcuts';
import { ApiOutlined, CheckCircleOutlined, CloseCircleOutlined, DeleteOutlined, PlusCircleOutlined } from '@ant-design/icons';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { NetworksColumns } from '../networks';
import NewNetworkButton from '../createNetwork';
import LinkContainersButton from '../linkContainersButton';
import { useTranslation } from 'react-i18next';

const NetworkContainerSetup = ({ config, containerInfo, refresh, newContainer, OnChange, OnConnect, OnDisconnect }) => {
  const { t } = useTranslation();
  const [networks, setNetworks] = React.useState([]);
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const isMobile = useMediaQuery(theme.breakpoints.down('sm'));
  const padding = isMobile ? '6px 4px' : '12px 10px';

  const isForceSecure = containerInfo.Config.Labels.hasOwnProperty('cosmos-force-network-secured') &&
  containerInfo.Config.Labels['cosmos-force-network-secured'] === 'true';

  React.useEffect(() => {
    API.docker.networkList().then((res) => {
      setNetworks(res.data);
    });
  }, []);

  const refreshAll = () => {
    if(refresh)
      refresh().then(() => {
        API.docker.networkList().then((res) => {
          setNetworks(res.data);
        });
      });
    else 
      API.docker.networkList().then((res) => {
        setNetworks(res.data);
      });
  };

  const connect = (network) => {
    if(!OnConnect) {
      return API.docker.attachNetwork(containerInfo.Name.replace('/', ''), network).then(() => {
        refreshAll();
      });
    } else {
      OnConnect(network);
      refreshAll();
    }
  }

  const disconnect = (network) => {
    if(!OnDisconnect) {
      return API.docker.detachNetwork(containerInfo.Name.replace('/', ''), network).then(() => {
        refreshAll();
      });
    } else {
      OnDisconnect(network);
      refreshAll();
    }
  }

  const getPortBindings = () => {
    let portsBound = [];
    let portsConfig = containerInfo.NetworkSettings.Ports;

    return Object.keys(portsConfig).map((contPort) => {
      return portsConfig[contPort] ? portsConfig[contPort].map((hostPort) => {
        let key = contPort + ":" + hostPort.HostPort + "/" + hostPort.protocol;
        if (portsBound.includes(key)) return null;
        portsBound.push(key);
        return {
          port: contPort.split('/')[0],
          protocol: contPort.split('/')[1],
          hostPort: hostPort.HostPort,
        };
      }) : {
        port: contPort.split('/')[0],
        protocol: contPort.split('/')[1],
        hostPort: '',
      }
    }).flat().filter((p) => p);
  }

  return (<Stack spacing={2}>
    <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
      <Formik
        initialValues={{
          networkMode: containerInfo.HostConfig.NetworkMode,
          ports: getPortBindings(),
        }}
        validate={(values) => {
          const errors = {};
          OnChange && OnChange(values);
          return errors;
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          if(newContainer) return false;
          setSubmitting(true);
          const realvalues = {
            portBindings: {},
            networkMode: values.networkMode,
          };
          values.ports.forEach((port) => {
            let key = `${port.port}/${port.protocol}`;
            
            if (port.hostPort) {
              if(!realvalues.portBindings[key]) realvalues.portBindings[key] = [];
              
              realvalues.portBindings[`${port.port}/${port.protocol}`].push({
                HostPort: port.hostPort,
              })
            }
          });
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
              refresh && refresh();
            });
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={2}>

              <MainCard title={t('mgmt.servApps.newContainer.networkSettingsTitle')}>
                <Stack spacing={4}>
                  {containerInfo.State && containerInfo.State.Status !== 'running' && (
                  <Alert severity="warning" style={{ marginBottom: '0px' }}>
                      {t('mgmt.servApps.networks.containerotRunningWarning')}
                    </Alert>
                  )}
                  {isForceSecure && (
                    <Alert severity="warning" style={{ marginBottom: '0px' }}>
                      {t('mgmt.servApps.networks.forcedSecurityWarning')}          
                    </Alert>
                  )}
                  <CosmosInputText
                    label={t('mgmt.servApps.networks.modeInput.modeLabel')}
                    name="networkMode"
                    placeholder={'default'}
                    formik={formik}
                  />
                  <CosmosFormDivider title={t('mgmt.servApps.networks.exposePortsTitle')} />
                  <div>
                    {formik.values.ports.map((port, idx) => (
                      <Grid container key={idx}>
                        <Grid item xs={4} style={{ padding }}>
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
                        <Grid item xs={4} style={{ padding }}>
                          <TextField
                            label={t('mgmt.servApps.networks.containerPortInput.containerPortLabel')}
                            fullWidth
                            value={port.port}
                            onChange={(e) => {
                              const newports = [...formik.values.ports];
                              newports[idx].port = e.target.value;
                              formik.setFieldValue('ports', newports);
                            }}
                          />
                        </Grid>
                        <Grid item xs={3} style={{ padding }}>
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
                        <Grid item xs={1} style={{ padding }}>
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
                  </div>
                  <div>
                    <Stack direction="column" spacing={2}>
                      {formik.errors.submit && (
                        <Grid item xs={12}>
                          <FormHelperText error>{formik.errors.submit}</FormHelperText>
                        </Grid>
                      )}
                      {!newContainer && <LoadingButton
                        fullWidth
                        disableElevation
                        disabled={formik.errors.submit}
                        loading={formik.isSubmitting}
                        size="large"
                        type="submit"
                        variant="contained"
                        color="primary"
                      >
                        {t('mgmt.servApps.networks.updatePortsButton')}
                      </LoadingButton>}
                    </Stack>
                  </div>
                </Stack>
              </MainCard>
              <MainCard title={t('global.networks')}>
                <Stack spacing={2}>

                {networks && <Stack spacing={2}>
                  {Object.keys(containerInfo.NetworkSettings.Networks).map((networkName) => {
                    const network = networks.find((n) => n.Name === networkName);
                    if (!network) {
                      return <Alert severity="error">
                        {t('mgmt.servApps.networks.removedNetConnectedError')} <strong>{networkName}</strong>. 
                        {t('mgmt.servApps.networks.removedNetConnectedEitherRecreate')}
                        <Button
                          style={{ marginLeft: '10px' }}
                          variant="outlined"
                          color="primary"
                          onClick={() => {
                            disconnect(networkName);
                          }}
                        >
                          {t('mgmt.servApps.networks.removedNetConnectedDisconnect')}
                        </Button>
                      </Alert>
                    }
                  })}
                </Stack>}

                {networks && <PrettyTableView
                  data={networks}
                  sort={(a, b) => a.Name > b.Name}
                  buttons={[
                    <NewNetworkButton refresh={refreshAll} />,
                    <LinkContainersButton 
                      refresh={refreshAll} 
                      originContainer={containerInfo.Name.replace('/', '')}
                      newContainer={newContainer}
                      OnConnect={OnConnect}
                    />,
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
                    ...NetworksColumns(theme, isDark, t),
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
                          {isConnected ? t('mgmt.servapps.containers.terminal.disconnectButton') : t('mgmt.servapps.containers.terminal.connectButton')}
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
                </Stack>
              </MainCard>
            </Stack>
          </form>
        )}
      </Formik>
    </div>
  </Stack>);
};

export default NetworkContainerSetup;