// material-ui
import * as React from 'react';
import { Alert, Button, Checkbox, FormControlLabel, FormLabel, Stack, Typography } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined, SyncOutlined, UserOutlined, KeyOutlined, ArrowUpOutlined, FileZipOutlined, ArrowDownOutlined, ConsoleSqlOutlined } from '@ant-design/icons';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import CircularProgress from '@mui/material/CircularProgress';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import IsLoggedIn from '../../../isLoggedIn';
import { useEffect, useState } from 'react';
import ResponsiveButton from '../../../components/responseiveButton';
import UploadButtons from '../../../components/fileUpload';
import NewDockerService from './newService';
import yaml from 'js-yaml';
import { CosmosCollapse, CosmosFormDivider, CosmosInputText } from '../../config/users/formShortcuts';
import VolumeContainerSetup from './volumes';
import DockerContainerSetup from './setup';
import whiskers from 'whiskers';

function checkIsOnline() {
  API.isOnline().then((res) => {
    window.location.reload();
  }).catch((err) => {
    setTimeout(() => {
      checkIsOnline();
    }, 1000);
  });
}

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '500px',
  maxWidth: '100%',
  width: '100%',
  margin: '0',
  position: 'relative',
  fontSize: '12px',
  fontFamily: 'monospace',
  whiteSpace: 'pre-wrap',
  wordWrap: 'break-word',
  wordBreak: 'break-all',
  lineHeight: '1.5',
  boxShadow: '0 0 10px rgba(0,0,0,0.5)',
  border: '1px solid rgba(255,255,255,0.1)',
  boxSizing: 'border-box',
  marginBottom: '10px',
  marginTop: '10px',
  marginLeft: '0',
  marginRight: '0',
}

const isNewerVersion = (v1, v2) => {
  return false;
}

const getHostnameFromName = (name) => {
  return name.replace('/', '').replace(/_/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase().replace(/\s/g, '-') + '.' + window.location.origin.split('://')[1]
}

const DockerComposeImport = ({ refresh, dockerComposeInit, installerInit, defaultName }) => {
  const [step, setStep] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [openModal, setOpenModal] = useState(false);
  const [dockerCompose, setDockerCompose] = useState('');
  const [service, setService] = useState({});
  const [ymlError, setYmlError] = useState('');
  const [serviceName, setServiceName] = useState(defaultName || 'my-service');
  const [hostnames, setHostnames] = useState({});
  const [overrides, setOverrides] = useState({});
  const [context, setContext] = useState({});
  const [installer, setInstaller] = useState(installerInit);

  useEffect(() => {
    if (!openModal) {
      return;
    }
    // fetch(dockerComposeInit)
    //   .then((res) => res.text())
    //   .then((text) => {
    //     setDockerCompose(text);
    // });
    if(dockerComposeInit)
    setDockerCompose(`
      {
        "cosmos-installer": {
          "form": [
            {
              "name": "caca",
              "label": "Caca?",
              "type": "text"
            }
          ]
        },
        "services": {
          "{ServiceName}": {
            "image": "lscr.io/linuxserver/jellyfin:latest",
            "container_name": "{ServiceName}",
            "restart": "unless-stopped",
            "environment": [
              "PUID=1000",
              "PGID=1000",
              "TZ=auto"
            ],
            "labels": {
              "cosmos-force-network-secured": "true",
              "caca": "{Context.caca}"
            },
            "volumes": [{
              "source": "{ServiceName}-config",
              "target": "/config",
              "type": "volume"
            }],
            "routes": [
              {
                "name": "{ServiceName}",
                "description": "Expose {ServiceName} to the web",
                "useHost": true,
                "target": "http://{ServiceName}:8096",
                "mode": "SERVAPP"
              }
            ]
          }
        }
      }
    `);
  }, [openModal, dockerComposeInit]);

  useEffect(() => {
    if (!openModal || installer) {
      return;
    }

    setYmlError('');
    if (dockerCompose === '') {
      return;
    }

    let isJson = dockerCompose && dockerCompose.trim().startsWith('{') && dockerCompose.trim().endsWith('}');

    // if Json 
    if (isJson) {
      try {
        let doc = JSON.parse(dockerCompose);
        if(typeof doc['cosmos-installer'] === 'object') {
          setInstaller(true);
        }
        setService(doc);
      } catch (e) {
        setYmlError(e.message);
      }
    }
    else {
      // if Yml
      let doc;
      let newService = {};
      try {
        doc = yaml.load(dockerCompose);

        if (typeof doc === 'object' && doc !== null && Object.keys(doc).length > 0 &&
          !doc.services && !doc.networks && !doc.volumes) {
          doc = {
            services: Object.assign({}, doc)
          }
        }

        // convert to the proper format
        if (doc.services) {
          Object.keys(doc.services).forEach((key) => {

            // convert volumes
            if (doc.services[key].volumes) {
              if (Array.isArray(doc.services[key].volumes)) {
                let volumes = [];
                doc.services[key].volumes.forEach((volume) => {
                  if (typeof volume === 'object') {
                    volumes.push(volume);
                  } else {
                    let volumeSplit = volume.split(':');
                    let volumeObj = {
                      Source: volumeSplit[0],
                      Target: volumeSplit[1],
                      Type: volume[0] === '/' ? 'bind' : 'volume',
                    };
                    volumes.push(volumeObj);
                  }
                });
                doc.services[key].volumes = volumes;
              }
            }

            // convert expose
            if (doc.services[key].expose) {
              doc.services[key].expose = doc.services[key].expose.map((port) => {
                return '' + port;
              })
            }

            //convert user
            if (doc.services[key].user) {
              doc.services[key].user = '' + doc.services[key].user;
            }

            // convert labels: 
            if (doc.services[key].labels) {
              if (Array.isArray(doc.services[key].labels)) {
                let labels = {};
                doc.services[key].labels.forEach((label) => {
                  const [key, value] = label.split('=');
                  labels['' + key] = '' + value;
                });
                doc.services[key].labels = labels;
              }
              if (typeof doc.services[key].labels == 'object') {
                let labels = {};
                Object.keys(doc.services[key].labels).forEach((keylabel) => {
                  labels['' + keylabel] = '' + doc.services[key].labels[keylabel];
                });
                doc.services[key].labels = labels;
              }
            }

            // convert environment
            if (doc.services[key].environment) {
              if (!Array.isArray(doc.services[key].environment)) {
                let environment = [];
                Object.keys(doc.services[key].environment).forEach((keyenv) => {
                  environment.push(keyenv + '=' + doc.services[key].environment[keyenv]);
                });
                doc.services[key].environment = environment;
              }
            }

            // convert network
            if (doc.services[key].networks) {
              if (Array.isArray(doc.services[key].networks)) {
                let networks = {};
                doc.services[key].networks.forEach((network) => {
                  if (typeof network === 'object') {
                    networks['' + network.name] = network;
                  }
                  else
                    networks['' + network] = {};
                });
                doc.services[key].networks = networks;
              }
            }

            // ensure container_name
            if (!doc.services[key].container_name) {
              doc.services[key].container_name = key;
            }
          });
        }

        // convert networks
        if (doc.networks) {
          if (Array.isArray(doc.networks)) {
            let networks = {};
            doc.networks.forEach((network) => {
              if (typeof network === 'object') {
                networks['' + network.name] = network;
              }
              else
                networks['' + network] = {};
            });
            doc.networks = networks;
          } else {
            let networks = {};
            Object.keys(doc.networks).forEach((key) => {
              networks['' + key] = doc.networks[key] || {};
            });
            doc.networks = networks;
          }
        }

        // convert volumes
        if (doc.volumes) {
          if (Array.isArray(doc.volumes)) {
            let volumes = {};
            doc.volumes.forEach((volume) => {
              if (!volume) {
                volume = {};
              }
              if (typeof volume === 'object') {
                volumes['' + volume.name] = volume;
              }
              else
                volumes['' + volume] = {};
            });
            doc.volumes = volumes;
          } else {
            let volumes = {};
            Object.keys(doc.volumes).forEach((key) => {
              volumes['' + key] = doc.volumes[key] || {};
            });
            doc.volumes = volumes;
          }
        }

      } catch (e) {
        console.log(e);
        setYmlError(e.message);
        return;
      }

      setService(doc);
    }
  }, [openModal, dockerCompose]);

  useEffect(() => {
    setOverrides({});
  }, [serviceName, hostnames, context]);

  useEffect(() => {
    if (!openModal || dockerCompose === '') {
      return;
    }

    try {
      if (installer) {
        const rendered = whiskers.render(dockerCompose, {
          ServiceName: serviceName,
          Hostnames: hostnames,
          Context: context,
        });

        const jsoned = JSON.parse(rendered);

        if (jsoned.services) {
          // GENERATE HOSTNAMES FORM
          let newHostnames = {};
          Object.keys(jsoned.services).forEach((key) => {
            if (jsoned.services[key].routes) {
              let routeId = 0;
              jsoned.services[key].routes.forEach((route) => {
                if (route.useHost) {
                  let newRoute = Object.assign({}, route);
                  if (route.useHost === true) {
                    newRoute.host = getHostnameFromName(key + (routeId > 0 ? '-' + routeId : ''))
                  }
                  
                  if(!newHostnames[key]) newHostnames[key] = {};


                  if(!newHostnames[key][route.name]) {
                    newHostnames[key][route.name] = newRoute;
                    if(hostnames[key] && hostnames[key][route.name]) {
                      newHostnames[key][route.name].host = hostnames[key][route.name].host;
                    }
                  }
                }
              });
            }
          });

          if(JSON.stringify(newHostnames) !== JSON.stringify(hostnames))
            setHostnames({...newHostnames});

          let newVolumes = [];

          Object.keys(jsoned.services).forEach((key) => {
            // APPLY OVERRIDE
            if (overrides[key]) {
              // prevent customizing static volumes
              if (jsoned.services[key].volumes && jsoned['cosmos-installer'] && jsoned['cosmos-installer']['frozen-volumes']) {
                jsoned['cosmos-installer']['frozen-volumes'].forEach((volume) => {
                  delete overrides[key].volumes;
                });
              }

              jsoned.services[key] = {
                ...jsoned.services[key],
                ...overrides[key],
              };
            }

            // APPLY HOSTNAMES
            if (hostnames[key]) {
              if (jsoned.services[key].routes) {
                jsoned.services[key].routes.forEach((route) => {
                  if (hostnames[key][route.name]) {
                    route.host = hostnames[key][route.name].host;
                  }
                });
              }
            }

            // CREATE NEW VOLUMES
            if (jsoned.services[key].volumes) {
              jsoned.services[key].volumes.forEach((volume) => {
                if (typeof volume === 'object' && !volume.existing) {
                  newVolumes.push(volume);
                } else if (typeof volume === 'object' && volume.existing) {
                  delete volume.existing;
                }
              });
            }
          });

          if (newVolumes.length > 0) {
            jsoned.volumes = jsoned.volumes || {};
            newVolumes.forEach((volume) => {
              jsoned.volumes[volume.source] = {
              };
            });
          }
        }

        setService(jsoned);
      }
    } catch (e) {
      setYmlError(e.message);
      return;
    }
  }, [openModal, dockerCompose, serviceName, hostnames, overrides]);

  const openModalFunc = () => {
    setOpenModal(true);
    setStep(0);
    setService({});
    setYmlError(null);
    setOverrides({});
    setHostnames({});
    setDockerCompose('');
    setInstaller(installerInit);
    setServiceName(defaultName || 'default-name');
  }

  return <>
    <Dialog open={openModal} onClose={() => setOpenModal(false)}>
      <DialogTitle>{installer ? "Installation" : "Import Compose File"}</DialogTitle>
      <DialogContent style={{ width: '800px', maxWidth: '100%' }}>
        <DialogContentText>
          {step === 0 && !installer && <><Stack spacing={2}>
            <Alert severity="warning" icon={<WarningOutlined />}>
              This is an experimental feature. It is recommended to use with caution. Please report any issue you find!
            </Alert>

            <UploadButtons
              accept='.yml,.yaml,.json'
              OnChange={(e) => {
                const file = e.target.files[0];
                const reader = new FileReader();
                reader.onload = (e) => {
                  setDockerCompose(e.target.result);
                };
                reader.readAsText(file);
              }}
            />

            <div style={{ color: 'red' }}>
              {ymlError}
            </div>

            <TextField
              multiline
              placeholder='Paste your docker-compose.yml / cosmos-compose.json here or use the file upload button.'
              fullWidth
              value={dockerCompose}
              onChange={(e) => setDockerCompose(e.target.value)}
              sx={preStyle}
              InputProps={{
                sx: {
                  color: '#EEE',
                }
              }}
              rows={20}></TextField>
          </Stack></>}

          {step === 0 && installer && <><Stack spacing={2}>
            <div style={{ color: 'red' }}>
              {ymlError}
            </div>

            {!ymlError && (<><FormLabel>Choose your service name</FormLabel>

              <TextField label="Service Name" value={serviceName} onChange={(e) => setServiceName(e.target.value)} />

              {service['cosmos-installer'] && service['cosmos-installer'].form.map((formElement) => {
                return formElement.type === 'checkbox' ? <FormControlLabel
                  control={<Checkbox checked={context[formElement.name] || formElement.initialValue} onChange={(e) => {
                    setContext({ ...context, [formElement.name]: e.target.checked });
                  }
                  } />}
                  label={formElement.label}
                /> : <TextField
                  label={formElement.label}
                  value={context[formElement.name] || formElement.initialValue}
                  onChange={(e) => {
                    setContext({ ...context, [formElement.name]: e.target.value });
                  }
                  } />
              })}

              {Object.keys(hostnames).map((serviceIndex) => {
                const service = hostnames[serviceIndex];
                return Object.keys(service).map((hostIndex) => {
                  const hostname = service[hostIndex];
                  return <>
                    <FormLabel>Choose URL for {hostname.name}</FormLabel>
                    <div style={{ opacity: 0.9, fontSize: '0.8em', textDecoration: 'italic' }}
                    >{hostname.description}</div>
                    <TextField key={serviceIndex + hostIndex} label="Hostname" value={hostname.host} onChange={(e) => {
                      hostnames[serviceIndex][hostname.name].host = e.target.value;
                      setHostnames({...hostnames});
                    }} />
                  </>
                })
              })}

              {service && service.services && Object.values(service.services).map((value) => {
                return <CosmosCollapse title={`Customize ${value.container_name}`}>
                  <Stack spacing={2}>
                    <DockerContainerSetup
                      newContainer
                      containerInfo={{
                        Name: '',
                        Image: '',
                        Config: {
                          Env: value.environment || [],
                          Labels: value.labels || {},
                        },
                        HostConfig: {
                          RestartPolicy: {},
                        }
                      }}
                      OnChange={(containerInfo) => {
                        setOverrides({
                          ...overrides,
                          [value.container_name]: {
                            environment: containerInfo.envVars,
                            labels: containerInfo.labels,
                          }
                        })
                      }}
                      noCard
                      installer
                    />
                    <CosmosFormDivider title="Volumes" />
                    <VolumeContainerSetup
                      newContainer
                      frozenVolumes={service['cosmos-installer'] && service['cosmos-installer']['frozen-volumes'] || []}
                      containerInfo={{
                        HostConfig: {
                          Binds: [],
                          Mounts: Object.keys(value.volumes).map(k => {
                            return {
                              Type: value.volumes[k].type || (k.startsWith('/') ? 'bind' : 'volume'),
                              Source: value.volumes[k].source || "",
                              Target: value.volumes[k].target || "",
                            }
                          }) || [],
                        }
                      }}
                      OnChange={(containerInfo, volumes) => {
                        setOverrides({
                          ...overrides,
                          [value.container_name]: {
                            volumes: containerInfo.volumes.map((v, k) => {
                              return {
                                type: v.Type,
                                source: v.Source,
                                target: v.Target,
                                existing: v.Type == 'volume' && volumes.find(v2 => v2.Source === v.Name),
                              }
                            })
                          }
                        })
                      }}
                      noCard
                    />
                  </Stack>
                </CosmosCollapse>
              })}

            </>)}
          </Stack></>}
          
          {step === 0 && dockerComposeInit && dockerCompose == '' && <Stack spacing={2} alignItems={'center'} style={{paddingTop: '20px'}}>
            <CircularProgress />
          </Stack>}

          {step === 1 && <Stack spacing={2}>
            <NewDockerService service={service} refresh={refresh} />
          </Stack>}
        </DialogContentText>
      </DialogContent>
      {!isLoading && <DialogActions>
        <Button onClick={() => {
          setOpenModal(false);
          setStep(0);
          setDockerCompose('');
          setYmlError('');
          setInstaller(false);
        }}>Cancel</Button>
        <Button disabled={!dockerCompose || ymlError} onClick={() => {
          if (step === 0) {
            setStep(1);
          } else {
            setStep(0);
          }
        }}>
          {step === 0 && 'Next'}
          {step === 1 && 'Back'}
        </Button>
      </DialogActions>}
    </Dialog>

    {(installerInit && service.minVersion && isNewerVersion(service.minVersion)) ?
      <Alert severity="error" icon={<WarningOutlined />}>
        This service requires a newer version of Cosmos. Please update Cosmos to install this service.
      </Alert>
      : <ResponsiveButton
        color="primary"
        onClick={() => {
          openModalFunc();
        }}
        variant={(installerInit ? "contained" : "outlined")}
        startIcon={(installerInit ? <ArrowDownOutlined /> : <ArrowUpOutlined />)}
      >
        {installerInit ? 'Install' : 'Import Compose File'}
      </ResponsiveButton>
    }
  </>;
};

export default DockerComposeImport;
