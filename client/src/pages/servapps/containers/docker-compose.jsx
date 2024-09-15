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
import { useEffect, useState } from 'react';
import ResponsiveButton from '../../../components/responseiveButton';
import UploadButtons from '../../../components/fileUpload';
import NewDockerService from './newService';
import yaml from 'js-yaml';
import { CosmosCollapse, CosmosFormDivider, CosmosInputPassword, CosmosInputText, CosmosSelect } from '../../config/users/formShortcuts';
import VolumeContainerSetup from './volumes';
import DockerContainerSetup from './setup';
import whiskers from 'whiskers';
import {version} from '../../../../../package.json';
import cmp from 'semver-compare';
import { HostnameChecker, getHostnameFromName } from '../../../utils/routes';
import { CosmosContainerPicker } from '../../config/users/containerPicker';
import { randomString } from '../../../utils/indexs';
import { useTranslation } from 'react-i18next';

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
  maxHeight: '520px',
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

const isNewerVersion = (minver) => {
  return cmp(version, minver) === -1;
}

const cleanUpStore = (service) => {
  let newService = Object.assign({}, service);
  delete newService['cosmos-installer'];
  delete newService['x-cosmos-installer'];
  return newService;
}

const convertDockerCompose = (config, serviceName, dockerCompose, setYmlError) => {
      let doc;

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
                      source: volumeSplit[0],
                      target: volumeSplit[1],
                      type: (volume[0] === '/' || volume[0] === '.') ? 'bind' : 'volume',
                    };
                    volumes.push(volumeObj);
                  }
                });
                doc.services[key].volumes = volumes;
              }
            }

            if(doc.services[key].volumes)
              Object.values(doc.services[key].volumes).forEach((volume) => {
                if (volume.source && volume.source[0] === '.') {
                  let defaultPath = (config && config.DockerConfig && config.DockerConfig.DefaultDataPath) || "/usr"
                  volume.source = defaultPath + volume.source.replace('.', '');
                }
              });

            // convert expose
            if (doc.services[key].expose) {
              doc.services[key].expose = doc.services[key].expose.map((port) => {
                return '' + port;
              })
            }

            // convert ports
            if (doc.services[key].ports && Array.isArray(doc.services[key].ports)) {
              let ports = [];
              doc.services[key].ports.forEach((port) => {
                if (typeof port === 'string') {
                  ports.push(port);
                  return;
                }
                ports.push(`${port.published}:${port.target}/${port.protocol || 'tcp'}`);
              });
              doc.services[key].ports = ports;
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
                  const [key, value] = label.split(/=(.*)/s);
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

            // convert Sysctls array to map
            if (doc.services[key].sysctls) {
              if (Array.isArray(doc.services[key].sysctls)) {
                let sysctls = {};
                doc.services[key].sysctls.forEach((sysctl) => {
                  sysctls['' + sysctl] = '';
                });
                doc.services[key].sysctls = sysctls;
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

            // convert devices
            if (doc.services[key].devices) {
              if (Array.isArray(doc.services[key].devices)) {
                let devices = [];
                doc.services[key].devices.forEach((device) => {
                  if(device.indexOf(':') === -1) {
                    devices.push(device + ':' + device);
                  } else {
                    devices.push(device);
                  }
                });
                doc.services[key].devices = devices;
              }
            }

            // convert command 
            if (doc.services[key].command) {
              if (typeof doc.services[key].command !== 'string') {
                doc.services[key].command = doc.services[key].command.join(' ');
              }
            }

            // convert DependsOn
            if (doc.services[key].depends_on) {
              if (Array.isArray(doc.services[key].depends_on)) {
                let depends_on = {};
                doc.services[key].depends_on.forEach((depend, index) => {
                  if (typeof depend === 'object') {
                    depends_on[index] = depend;
                  } else {
                    depends_on['' + depend] = {}; 
                  }
                });
                doc.services[key].depends_on = depends_on;
              }
            }

            // convert healthcheck
            if (doc.services[key].healthcheck) {
              const toConvert = ["timeout", "interval", "start_period"];
              toConvert.forEach((valT) => {
                if(typeof doc.services[key].healthcheck[valT] === 'string') {
                  let original = doc.services[key].healthcheck[valT];
                  let value = parseInt(original);
                  if (original.endsWith('m')) {
                    value = value * 60;
                  } else if (original.endsWith('h')) {
                    value = value * 60 * 60;
                  } else if (original.endsWith('d')) {
                    value = value * 60 * 60 * 24;
                  }
                  doc.services[key].healthcheck[valT] = value;
                }
              });
            }

            // ensure hostname
            if (!doc.services[key].hostname) {
              doc.services[key].hostname = key;
            }
            
            // if service name is set, namespace the keys
            if (serviceName) {
              let newKey = serviceName + '-' + key;
              doc.services[newKey] = doc.services[key];
              doc.services[newKey].old_key = key;
              delete doc.services[key];
              key = newKey;
            }

            // ensure container_name
            if (!doc.services[key].container_name) {
              doc.services[key].container_name = key;
            }
          });

          // ensure depends on names
          Object.keys(doc.services).forEach((key) => {
            if (doc.services[key].depends_on) {
              Object.keys(doc.services[key].depends_on).forEach((depend) => {
                Object.keys(doc.services).forEach((potentialMatch) => {
                  if (doc.services[potentialMatch].old_key === depend) {
                    let name = doc.services[potentialMatch].container_name || potentialMatch;
                    doc.services[key].depends_on[name] = doc.services[key].depends_on[depend];
                    if (name !== depend) {
                      delete doc.services[key].depends_on[depend];
                    }
                  }
                });
              });
            }
          });

          // ensure network mode names
          Object.keys(doc.services).forEach((key) => {
            if (doc.services[key].network_mode) {
              if (doc.services[key].network_mode && (doc.services[key].network_mode.startsWith('service:') || doc.services[key].network_mode.startsWith('container:'))) {
                let service = doc.services[key].network_mode.split(':')[1];
                let found = false;
                Object.keys(doc.services).forEach((potentialMatch) => {
                  if (doc.services[potentialMatch].old_key === service) {
                    found = true;
                    doc.services[key].network_mode = 'container:' + (doc.services[potentialMatch].container_name || potentialMatch);
                  }
                });
                if (!found) {
                  doc.services[key].network_mode = 'container:NOT_FOUND';
                }
              }
            }
          });

          // for each network mode that are container, add a label and remove hostname
          Object.keys(doc.services).forEach((key) => {
            if (doc.services[key].network_mode && doc.services[key].network_mode.startsWith('container:')) {
              doc.services[key].labels = doc.services[key].labels || {};
              doc.services[key].labels['cosmos-force-network-mode'] = doc.services[key].network_mode;
              
              // remove hostname
              if (doc.services[key].hostname) {
                delete doc.services[key].hostname;
              }
            }
          });
          
          // clean up old-keys
          Object.keys(doc.services).forEach((key) => {
            delete doc.services[key].old_key;
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

        // create default network
        let hasDefaultNetwork = false;
        if (doc.services) {
          Object.keys(doc.services).forEach((key) => {
            if(!doc.services[key].network_mode) {
              doc.services[key].network_mode = 'cosmos-' + serviceName + '-default';
              hasDefaultNetwork = true;
            }
          });
        }

        if(hasDefaultNetwork) {
          if(!doc.networks) {
            doc.networks = {}
          }
          
          doc.networks['cosmos-' + serviceName + '-default'] = {}
        }

        // stack up
        if (doc.services && Object.keys(doc.services).length > 1) {
          let hasMain = false;
          Object.keys(doc.services).forEach((key) => {
            if(!doc.services[key].labels) {
              doc.services[key].labels = {};
            }
            if(!doc.services[key].labels['cosmos.stack'])
              doc.services[key].labels['cosmos.stack'] = serviceName;
            if(doc.services[key].labels['cosmos.stack.main'])
              hasMain = true;
          });

          if(!hasMain) {
            Object.keys(doc.services).forEach((key) => {
              if(!hasMain) {
                if(doc.services[key].labels['cosmos.stack'] == serviceName) {
                  doc.services[key].labels['cosmos.stack.main'] = "true";
                  hasMain = true;
                }
              }
            });
          }
        }

        // cosmos features
        if (doc.services) {
          Object.keys(doc.services).forEach((key) => {
            if(doc.services[key]['x-routes']) {
              doc.services[key].routes = doc.services[key]['x-routes'];
              delete doc.services[key]['x-routes'];
            }
          });
          
          if(doc.services['x-post_install']) {
            doc['cosmos-installer'] = {
              post_install: doc.services['x-post_install'],
            }
            delete doc.services['x-post_install'];
          }
        }
        
        // enable market features
        if (doc['x-cosmos-installer']) {
          doc['cosmos-installer'] = doc['x-cosmos-installer'];
          delete doc['x-cosmos-installer'];
        }

        return doc;
      } catch (e) {
        setYmlError(e.message);
        return;
      }
}

const DockerComposeImport = ({ refresh, dockerComposeInit, installerInit, defaultName }) => {
  const { t, i18n } = useTranslation();
  const cleanDefaultName = defaultName && defaultName.replace(/\s/g, '-').replace(/[^a-zA-Z0-9-]/g, '');
  const [step, setStep] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [openModal, setOpenModal] = useState(false);
  const [dockerCompose, setDockerCompose] = useState('');
  const [service, setService] = useState({});
  const [ymlError, setYmlError] = useState('');
  const [serviceName, setServiceName] = useState(null);
  const [hostnames, setHostnames] = useState({});
  const [overrides, setOverrides] = useState({});
  const [context, setContext] = useState({});
  const [installer, setInstaller] = useState(installerInit);
  const [config, setConfig] = useState({});

  let hostnameErrors = () => {
    let broken = false;
    Object.values(hostnames).forEach((service) => {
      Object.values(service).forEach((route) => {
        if(!route.host || route.host.match(/\s/g)) {
          broken = true;
        }
      });
    });
    return broken;
  }

  const [passwords, setPasswords] = useState([
    randomString(24),
    randomString(24),
    randomString(24),
    randomString(24)
  ]);

  const resetPassword = () => {
    setPasswords([
      randomString(24),
      randomString(24),
      randomString(24),
      randomString(24)
    ]);
  }


  function refreshConfig() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  React.useEffect(() => {
    refreshConfig();
  }, []);

  useEffect(() => {
    if (!openModal) {
      return;
    }
    if(dockerComposeInit)
      fetch(dockerComposeInit)
        .then((res) => res.text())
        .then((text) => {
          setDockerCompose(text);
      });
  }, [openModal, dockerComposeInit]);

  useEffect(() => {
    setOverrides({});
  }, [serviceName, hostnames, context]);

  useEffect(() => {
    if (!openModal || dockerCompose === '') {
      return;
    }

      setYmlError('');
      if (dockerCompose === '') {
        return;
      }
    
    let isJson = dockerCompose && dockerCompose.trim().startsWith('{') && dockerCompose.trim().endsWith('}');

    try {
      const rendered = whiskers.render(dockerCompose.replace(/{StaticServiceName}/ig, serviceName), {
        ServiceName: serviceName,
        Hostnames: hostnames,
        Context: context,
        Passwords: passwords,
        CPU_ARCH: API.CPU_ARCH,
        CPU_AVX: API.CPU_AVX,
        DefaultDataPath: (config && config.DockerConfig && config.DockerConfig.DefaultDataPath) || "/usr",
      });

      let jsoned;
      if(isJson) {
        jsoned = JSON.parse(rendered);
      } else {
        jsoned = convertDockerCompose(config, serviceName, rendered, setYmlError);
      }
      
      if(!serviceName && !Object.keys(service).length) {
        if(jsoned['name'] && jsoned['name'].trim() !== '') {
          setServiceName(jsoned['name']);
        } else if (jsoned['services'] && Object.keys(jsoned['services']).length > 0 && Object.keys(jsoned['services'])[0].trim() !== '') {
          setServiceName(Object.keys(jsoned['services'])[0]);
        } else {
          setServiceName(cleanDefaultName || 'default-service');
        }
      }

      if (typeof jsoned['cosmos-installer'] === 'object' || typeof jsoned['x-cosmos-installer'] === 'object') {
        setInstaller(true);

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
                    newRoute.host = getHostnameFromName(key + (routeId > 0 ? '-' + routeId : ''), newRoute, config);
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

          // SET DEFAULT CONTEXT
          let newContext = {};
          if (jsoned['cosmos-installer'] && jsoned['cosmos-installer'].form) {
            jsoned['cosmos-installer'].form.forEach((field) => {
              [field.name, field['name-container']].forEach((fieldName) => {
                if(typeof context[fieldName] === "undefined" && typeof field.initialValue !== "undefined") {
                  newContext[fieldName] = field.initialValue;
                } else if (typeof context[fieldName] !== "undefined") {
                  newContext[fieldName] = context[fieldName];
                }
              });
            });
          }
          
          if(JSON.stringify(Object.keys(newContext)) !== JSON.stringify(Object.keys(context)))
            setContext({...newContext});

          Object.keys(jsoned.services).forEach((key) => {
            // APPLY OVERRIDE
            if (overrides[key]) {
              // prevent customizing static volumes
              if (jsoned.services[key].volumes && jsoned['cosmos-installer'] && jsoned['cosmos-installer']['frozen-volumes']) {
                jsoned['cosmos-installer']['frozen-volumes'].forEach((volumeName) => {
                  const keyVolume = overrides[key].volumes.findIndex((v) => {
                    return v.source === volumeName;
                  });
                  delete overrides[key].volumes[keyVolume];
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
                if (typeof volume === 'object' && !volume.source.startsWith('/') && !volume.existing) {
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

          // CREATE DEFAULT NETWORK
          if(!jsoned['cosmos-installer'] || !jsoned['cosmos-installer']['skip-default-network']) {
            let hasDefaultNetwork = false;
            if (jsoned.services) {
              Object.keys(jsoned.services).forEach((key) => {
                if(!jsoned.services[key].network_mode) {
                  jsoned.services[key].network_mode = 'cosmos-' + serviceName + '-default';
                  hasDefaultNetwork = true;
                }
              });
            }

            if(hasDefaultNetwork) {
              if(!jsoned.networks) {
                jsoned.networks = {}
              }
              
              jsoned.networks['cosmos-' + serviceName + '-default'] = {
                Labels: {
                  'cosmos.stack': serviceName,
                }
              }
            }
          }
        }

        setService(jsoned);
      } else {
        if(!isJson) {
          setService(convertDockerCompose(config, serviceName, dockerCompose, setYmlError));
        } else {
          setService(JSON.parse(dockerCompose));
        }
      }
    } catch (e) {
      setYmlError(e.message);
      return;
    }
  }, [openModal, dockerCompose, serviceName, hostnames, overrides, installer, config]);

  const openModalFunc = () => {
    setOpenModal(true);
    setStep(0);
    setService({});
    setYmlError(null);
    setOverrides({});
    setHostnames({});
    setContext({});
    setDockerCompose('');
    setInstaller(installerInit);
    setServiceName(null);
    resetPassword();
  }

  return <>
    <Dialog open={openModal} onClose={() => setOpenModal(false)} fullWidth maxWidth={'sm'}>
      <DialogTitle>{installer ? t('mgmt.servapps.compose.installTitle') : t('mgmt.servapps.importComposeFileButton')}</DialogTitle>
      <DialogContent style={{ width: '100%' }}>
        <DialogContentText>
          {step === 0 && !installer && <><Stack spacing={2}>


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
              placeholder={t('mgmt.servapps.pasteComposeButton.pasteComposePlaceholder')}
              fullWidth
              value={dockerCompose}
              onChange={(e) => setDockerCompose(e.target.value)}
              sx={preStyle}
              InputProps={{
                sx: {
                  color: '#EEE',
                }
              }}
              minRows={20}></TextField>
          </Stack></>}

          {step === 0 && installer && <><Stack spacing={2}>
            <div style={{ color: 'red' }}>
              {ymlError}
            </div>

            {!ymlError && (<><FormLabel>{t('mgmt.servApps.newContainer.serviceNameInput')}</FormLabel>

              <TextField label="" value={serviceName} onChange={(e) => setServiceName(e.target.value)} />

              {service['cosmos-installer'] && service['cosmos-installer'].form && service['cosmos-installer'].form.map((formElement) => {
                return formElement.type === 'checkbox' ?
                <FormControlLabel
                  control={<Checkbox checked={context[formElement.name]} onChange={(e) => {
                    setContext({ ...context, [formElement.name]: e.target.checked });
                  }
                  } />}
                  label={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                /> : (formElement.type === 'password' || formElement.type === 'email') ?
                  <TextField
                  label={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                  value={context[formElement.name]}
                  type={formElement.type}
                  onChange={(e) => {
                    setContext({ ...context, [formElement.name]: e.target.value });
                  }
                  } />
                :  (formElement.type === 'select') ?
                    <CosmosSelect
                    name={formElement.name} 
                    label={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                    formik={{
                      values: {
                        [formElement.name] : context[formElement.name]
                      },
                      touched: {},
                      errors: {},
                      setFieldValue: () => {},
                      handleChange: () => {}
                    }}
                    onChange={(e) => {
                      setContext({ ...context, [formElement.name]: e.target.value });
                    }}
                    options={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.options'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.options'] || formElement.options }
                  />
                : formElement.type === 'hostname' ? 
                  <>
                    <TextField
                      label={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                      value={context[formElement.name]}
                      onChange={(e) => {
                        setContext({ ...context, [formElement.name]: e.target.value });
                      }
                      } />
                    <HostnameChecker hostname={context[formElement.name]} />
                  </>
                : formElement.type === 'container' || formElement.type === 'container-full' ? 
                  <CosmosContainerPicker
                      name={formElement.name} 
                      formik={{
                        values: {
                          [formElement.name] : context[formElement.name]
                        },
                        errors: {},
                        setFieldValue: (name, value) => {
                          setContext({ ...context, [formElement.name]: value });
                        },
                      }}
                      nameOnly={formElement.type === 'container'}
                      label={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                      onTargetChange={(_, name) => {
                        setContext({ ...context, [formElement['name-container']]: name });
                      }}
                  />
                : formElement.type === 'error' || formElement.type === 'info' || formElement.type === 'warning' ?
                  <Alert severity={formElement.type}>
                    { service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                  </Alert>
                
                : <TextField
                  label={ service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage]?.['form.'+formElement.name+'.label'] || service['cosmos-installer']?.translation?.[i18n?.resolvedLanguage.substr?.(0,2)]?.['form.'+formElement.name+'.label'] || formElement.label }
                  value={context[formElement.name]}
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
                    <FormLabel>{t('mgmt.servApps.newContainer.chooseUrl')} {hostname.name}</FormLabel>
                    <div style={{ opacity: 0.9, fontSize: '0.8em', textDecoration: 'italic' }}
                    >{hostname.description}</div>
                    <TextField key={serviceIndex + hostIndex} label="Hostname" value={hostname.host} onChange={(e) => {
                      hostnames[serviceIndex][hostname.name].host = e.target.value;
                      setHostnames({...hostnames});
                    }} />
                    <HostnameChecker hostname={hostname.host} />
                  </>
                })
              })}

              {service && service.services && Object.values(service.services).map((value) => {
                return <CosmosCollapse title={t('mgmt.servApps.newContainer.customize', {container_name: value.container_name})}>
                  <Stack spacing={2}>
                    <DockerContainerSetup
                      newContainer
                      containerInfo={{
                        Name: '',
                        Image: '',
                        Config: {
                          Env: value.environment || [],
                          Labels: value.labels || {},
                          User: value.user || '',
                        },
                        HostConfig: {
                          RestartPolicy: {},
                          Devices: value.devices || [],
                        }
                      }}
                      OnChange={(containerInfo) => {
                        setOverrides({
                          ...overrides,
                          [value.container_name]: {
                            ...overrides[value.container_name],
                            environment: containerInfo.envVars,
                            labels: containerInfo.labels,
                            devices: containerInfo.devices,
                          }
                        })
                      }}
                      noCard
                      installer
                    />
                    <CosmosFormDivider title={t('mgmt.servapps.networks.volumes')} />
                    <VolumeContainerSetup
                      newContainer
                      frozenVolumes={service['cosmos-installer'] && service['cosmos-installer']['frozen-volumes'] || []}
                      containerInfo={{
                        HostConfig: {
                          Binds: [],
                          Mounts: value.volumes && Object.keys(value.volumes).map(k => {
                            return {
                              Type: value.volumes[k].type || (k.startsWith('/') ? t('mgmt.servapps.newContainer.volumes.bindInput') : t('global.volume')),
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
                            ...overrides[value.container_name],
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
            <NewDockerService service={cleanUpStore(service)} refresh={refresh} />
          </Stack>}
        </DialogContentText>
      </DialogContent>
      {(installerInit && service.minVersion && isNewerVersion(service.minVersion)) ?
      <Alert severity="error" icon={<WarningOutlined />}>
        {t('mgmt.servApps.newContainer.cosmosOutdatedError')}
      </Alert>
      : 
      (!isLoading && <DialogActions>
        <Button onClick={() => {
          setOpenModal(false);
          setStep(0);
          setDockerCompose('');
          setYmlError('');
          setInstaller(false);
          setServiceName(null);
          setContext({});
          setHostnames({});
          setOverrides({});
        }}>{t('global.close')}</Button>
        <Button disabled={!dockerCompose || ymlError || hostnameErrors()} onClick={() => {
          if (step === 0) {
            setStep(1);
          } else {
            setStep(0);
          }
        }}>
          {step === 0 && t('global.next')}
          {step === 1 && t('global.backAction')}
        </Button>
      </DialogActions>)}
    </Dialog>

    <ResponsiveButton
      color="primary"
      onClick={() => {
        openModalFunc();
      }}
      variant={(installerInit ? "contained" : "outlined")}
      startIcon={(installerInit ? <ArrowDownOutlined /> : <ArrowUpOutlined />)}
    >
      {installerInit ? t('mgmt.servapps.compose.installButton') : t('mgmt.servapps.importComposeFileButton')}
    </ResponsiveButton>
    
  </>;
};

export default DockerComposeImport;
