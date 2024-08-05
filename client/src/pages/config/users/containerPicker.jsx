import * as React from 'react';
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
  AccordionSummary,
  AccordionDetails,
  Accordion,
  Chip,
} from '@mui/material';
import { Formik, Field } from 'formik';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import AnimateButton from '../../../components/@extended/AnimateButton';
import RestartModal from './restart';
import Autocomplete from '@mui/material/Autocomplete';
import CircularProgress from '@mui/material/CircularProgress';
import { useTranslation } from 'react-i18next';

import defaultport from '../../servapps/defaultport.json';

import * as API  from '../../../api';

export function CosmosContainerPicker({formik, nameOnly, lockTarget, TargetContainer, onTargetChange, label = "Container Name", name = "Target"}) {
  const { t } = useTranslation();
  const [open, setOpen] = React.useState(false);
  const [containers, setContainers] = React.useState([]);
  const [hasPublicPorts, setHasPublicPorts] = React.useState(false);
  const [isOnBridge, setIsOnBridge] = React.useState(false);
  const [options, setOptions] = React.useState(null);
  const [portsOptions, setPortsOptions] = React.useState(null);
  const loading = options === null;

  let targetResult = {
    container: 'null',
    port: "",
    protocol: "http",
  }

  let preview = formik.values[name];

  if(preview) {
    let protocols = preview.split("://")
    let p1_ = protocols.length > 1 ? protocols[1] : protocols[0]

    targetResult = {
      container: '/' + p1_.split(":")[0],
      port: p1_.split(":")[1],
      protocol: preview.split("://")[0],
      containerObject: !loading && containers.find((container) => container.Names[0] === '/' + p1_.split(":")[0]),
    }
  }

  function getTarget() {
    return targetResult.protocol + (targetResult.protocol != '' ? "://" : '') + targetResult.container.replace("/", "") + ":" + targetResult.port
  }

  const postContainerChange = (newContainer) => {
    let portsTemp = []
    newContainer.Ports.forEach((port) => {
      portsTemp.push(port.PrivatePort)
      if (port.PublicPort) {
        setHasPublicPorts(true);
      }
    })
    setPortsOptions(portsTemp)

    if(targetResult.port == '') {
        targetResult.port = '80'
        
        if(portsTemp.length > 0) {
          // pick best default port
          // set default to first port
        targetResult.port = portsTemp[0]
        // first, check if a manual override exists
        let override = Object.keys(defaultport.overrides).find((key) => {
          let keyMatch = new RegExp(key, "i");
          return newContainer.Image.match(keyMatch) && portsTemp.includes(defaultport.overrides[key])
        });
        
        if(override) {
          targetResult.port = defaultport.overrides[override]
        } else {
          // if not, check the default list of common ports
          let priorityList = defaultport.priority;
          priorityList.find((_portReg) => {
            return portsTemp.find((portb) => {
              let portReg = new RegExp(_portReg, "i");
              if(portb.toString().match(portReg)) {
                targetResult.port = portb
                return true;
              }
            })
          })
        }
      }
    }
    
    formik.setFieldValue(name, getTarget());

    if(newContainer.NetworkSettings.Networks["bridge"]) {
      setIsOnBridge(true);
    }
  }

  const onContainerChange = (newContainer) => {
    if(loading) return;
    targetResult.container = newContainer.Names[0]
    targetResult.containerObject = newContainer
    targetResult.port = ''
    targetResult.protocol = 'http'
    formik.setFieldValue(name, getTarget())

    postContainerChange(newContainer)
  }

  React.useEffect(() => {
    if(lockTarget) {
      targetResult.container = TargetContainer.Names[0]
      targetResult.containerObject = TargetContainer
      targetResult.port = ''
      targetResult.protocol = 'http'
      onContainerChange(TargetContainer)
    }
  }, [])

  React.useEffect(() => {
    let active = true;
    
    if (!loading) {
      return undefined;
    }

    (async () => {
      const res = await API.docker.list()
      setContainers(res.data);

      let names = res.data.map((container) => container.Names[0])

      if (active) {
        setOptions([...names]);
      }
      
      if (targetResult.container !== 'null') {
        postContainerChange(res.data.find((container) => container.Names[0] === targetResult.container) || targetResult.containerObject)
      }
    })();

    return () => {
      active = false;
    };
  }, [loading]);

  React.useEffect(() => {
    if (!open) {
      setOptions([]);
    }
  }, [open]);

  const newTarget = formik.values[name];
  React.useEffect(() => {
    if(onTargetChange) {
      onTargetChange(newTarget, targetResult.container !== 'null' ? targetResult.container.replace("/", "") : "", targetResult)
    }
  }, [newTarget])

  return ( <Grid item xs={12}>
    <Stack spacing={1}>
    <InputLabel htmlFor={name + "-autocomplete"}>{t('mgmt.config.containerPicker.containerPortInput')}</InputLabel>
    {!loading && <Autocomplete
      id={name + "-autocomplete"}
      open={open}
      disabled={lockTarget}
      onOpen={() => {
        !lockTarget && setOpen(true);
      }}
      onClose={() => {
        !lockTarget && setOpen(false);
      }}
      onChange={(event, newValue) => {
        !lockTarget && onContainerChange(newValue)
      }}
      isOptionEqualToValue={(option, value) => {
        return !lockTarget && (option.Names[0] === value.Names[0])
      }}
      getOptionLabel={(option) => {
        return !lockTarget ? option.Names[0] : TargetContainer.Names[0]
      }}
      options={containers}
      loading={loading}
      freeSolo={true}
      placeholder={t('mgmt.config.containerPicker.containerNameSelection.containerNameValidation')}
      value={lockTarget ? TargetContainer : (targetResult.containerObject || {Names: ['...']})}
      renderInput={(params) => (
        <TextField
          {...params}
          InputProps={{
            ...params.InputProps,
            endAdornment: (
              <React.Fragment>
                {loading ? <CircularProgress color="inherit" size={20} /> : null}
                {params.InputProps.endAdornment}
              </React.Fragment>
            ),
          }}
        />
      )}
    />}
    {!nameOnly && <>
      <InputLabel htmlFor={name + "-port"}>{t('mgmt.config.containerPicker.containerPortInput')}</InputLabel>
      <Autocomplete
        className="px-2 my-2"
        variant="outlined"
        name={name + "-port"}
        id={name + "-port"}
        value={targetResult.port}
        options={((portsOptions && portsOptions.length) ? portsOptions : [])}
        placeholder='Select a port'
        onBlur={(event) => {
          targetResult.port = event.target.value || '';
          formik.setFieldValue(name, getTarget())
        }}
        freeSolo
        filterOptions={(x) => x} // disable filtering
        getOptionLabel={(option) => '' + option}
        isOptionEqualToValue={(option, value) => {
          return ('' + option) === value
        }}
        onChange={(event, newValue) => {
          targetResult.port = newValue || '';
          formik.setFieldValue(name, getTarget())
        }}
        renderInput={(params) => <TextField {...params} />}
      />
        {targetResult.port == '' && targetResult.port == 0 && <FormHelperText error id="standard-weight-helper-text-name-login">
          {t('mgmt.config.containerPicker.containerPortSelection.containerPortValidation')}
        </FormHelperText>}
      

      <InputLabel htmlFor={name + "-protocol"}>{t('mgmt.config.containerPicker.containerProtocolInput')}</InputLabel>
      <TextField
        type="text"
        name={name + "-protocol"}
        defaultValue={targetResult.protocol}
        onChange={(event) => {
          targetResult.protocol = event.target.value && event.target.value.toLowerCase()
          formik.setFieldValue(name, getTarget())
        }}
      />

      <InputLabel htmlFor={name}>{t('mgmt.config.containerPicker.targetTypePreview')}</InputLabel>
      <TextField
        name={name}
        placeholder={t('mgmt.config.containerPicker.targetTypePreview.targetTypePreviewLabel')}
        id={name}
        value={formik.values[name]}
        disabled={true}
      />
      </>} 
      
      {formik.errors[name] && (
        <FormHelperText error id="standard-weight-helper-text-name-login">
          {formik.errors[name]}
        </FormHelperText>
      )}
  </Stack>
  </Grid>
  );
}