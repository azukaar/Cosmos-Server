// material-ui
import * as React from 'react';
import { Alert, Button, FormLabel, Stack, Typography } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined, SyncOutlined, UserOutlined, KeyOutlined, ArrowUpOutlined, FileZipOutlined, ArrowDownOutlined } from '@ant-design/icons';
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

const getHostnameFromName = (name) => {
  return name.replace('/', '').replace(/_/g, '-').replace(/[^a-zA-Z0-9-]/g, '').toLowerCase().replace(/\s/g, '-') + '.' + window.location.origin.split('://')[1]
}

const DockerComposeImport = ({ refresh, dockerComposeInit, installer, defaultName }) => {
  const [step, setStep] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [openModal, setOpenModal] = useState(false);
  const [dockerCompose, setDockerCompose] = useState('');
  const [service, setService] = useState({});
  const [ymlError, setYmlError] = useState('');
  const [serviceName, setServiceName] = useState(defaultName || 'my-service');
  const [hostnames, setHostnames] = useState([]);

  useEffect(() => {
    const text = fetch(dockerComposeInit)
      .then((res) => res.text())
      .then((text) => {
        setDockerCompose(text);
    });
  }, [dockerComposeInit]);

  useEffect(() => {
    if(!openModal) {
      return;
    }
    setYmlError('');
    if (dockerCompose === '') {
      return;
    }

    let isJson = dockerCompose.trim().startsWith('{') && dockerCompose.trim().endsWith('}');

    // if Json 
    if (isJson) {
      try {
        let doc = JSON.parse(dockerCompose);

        if(installer) {
          doc = JSON.parse(dockerCompose.replace(/\{\{self\}\}/gi, serviceName));

          // check hostnames for each service
          let hostnames = [];

          if(doc.services)
            Object.keys(doc.services).forEach((key) => {
              if (doc.services[key].routes) {
                let routeId = 0;
                doc.services[key].routes.forEach((route) => {
                  if (route.useHost) {
                    let newRoute = Object.assign({}, route);
                    if(route.useHost === true) {
                      newRoute.host = getHostnameFromName(key + (routeId > 0 ? '-' + routeId : '')) 
                    }
                    hostnames.push(newRoute);
                  }
                });
              }
            });

          setHostnames(hostnames);
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
            if(Array.isArray(doc.services[key].volumes)) {
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
            if(!volume) {
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
    if(!openModal) {
      return;
    }

    try {
      if (installer) {
        let doc = JSON.parse(dockerCompose.replace(/\{\{self\}\}/gi, serviceName));

        if(doc.services)
          Object.keys(doc.services).forEach((key) => {
            if (doc.services[key].routes) {
              doc.services[key].routes.forEach((route, index) => {
                doc.services[key].routes[index] = {
                  ...hostnames[index],
                  name: route.name,
                  description: route.description,
                };
              });
            }
          });

        setService(doc);
      }
    } catch (e) {
      setYmlError(e.message);
      return;
    }
  }, [openModal, installer, serviceName, hostnames]);

  return <>
    <Dialog open={openModal} onClose={() => setOpenModal(false)} >
      <DialogTitle>{installer ? "Installation" : "Import Compose File"}</DialogTitle>
      <DialogContent style={{width: '800px', maxWidth: '100%'}}>
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
              {hostnames.map((hostname, index) => {
                return <>
                  <FormLabel>Choose URL for {hostname.name}</FormLabel>
                  <div style={{ opacity: 0.9, fontSize: '0.8em', textDecoration: 'italic'}}
                  >{hostname.description}</div>
                  <TextField key={index} label="Hostname" value={hostname.host} onChange={(e) => {
                    const newHostnames = [...hostnames];
                    newHostnames[index].host = e.target.value;
                    setHostnames(newHostnames);
                  }} />
                </>
              })}
              </>)}
          </Stack></>}

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

    <ResponsiveButton
      color="primary"
      onClick={() => setOpenModal(true)}
      variant={(installer ? "contained" : "outlined")} 
      startIcon={(installer ? <ArrowDownOutlined /> : <ArrowUpOutlined />)}
    >
      {installer ? 'Install' : 'Import Compose File'}
    </ResponsiveButton>
  </>;
};

export default DockerComposeImport;
