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

import * as API  from '../../../api';

export function CosmosContainerPicker({formik, lockTarget, TargetContainer}) {
  const [open, setOpen] = React.useState(false);
  const [containers, setContainers] = React.useState([]);
  const [hasPublicPorts, setHasPublicPorts] = React.useState(false);
  const [isOnBridge, setIsOnBridge] = React.useState(false);
  const [options, setOptions] = React.useState([]);
  const [portsOptions, setPortsOptions] = React.useState([]);
  const loading = open && options.length === 0;

  const name = "Target"
  const label = "Container Name"
  let targetResult = {
    container: 'null',
    port: "",
    protocol: "http",
  }

  let preview = formik.values[name];

  if(preview && preview.includes("://") && preview.includes(":")) {
    let p1_ = preview.split("://")[1]
    targetResult = {
      container: '/' + p1_.split(":")[0],
      port: p1_.split(":")[1],
      protocol: preview.split("://")[0],
      containerObject: containers.find((container) => container.Names[0] === '/' + p1_.split(":")[0]),
    }
  }

  function getTarget() {
    return targetResult.protocol + "://" + targetResult.container.replace("/", "") + ":" + targetResult.port
  }

  const onContainerChange = (newContainer) => {
    targetResult.container = newContainer.Names[0]
    targetResult.containerObject = newContainer
    targetResult.port = ''
    targetResult.protocol = 'http'
    formik.setFieldValue(name, getTarget())

    let portsTemp = []
    newContainer.Ports.forEach((port) => {
      portsTemp.push(port.PrivatePort)
      if (port.PublicPort) {
        setHasPublicPorts(true);
      }
    })
    setPortsOptions(portsTemp)

    if(newContainer.NetworkSettings.Networks["bridge"]) {
      setIsOnBridge(true);
    }
  }

  React.useEffect(() => {
    if(lockTarget) {
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

  return ( <Grid item xs={12}>
    <Stack spacing={1}>
    <InputLabel htmlFor={name + "-autocomplete"}>{label}</InputLabel>
    <Autocomplete
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
      placeholder={"Please select a container"}
      defaultValue={lockTarget ? TargetContainer : targetResult.containerObject}
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
    />
    
    {(portsOptions.length > 0) ? (<>
    <InputLabel htmlFor={name + "-port"}>Container Port</InputLabel>
    <TextField
      className="px-2 my-2"
      variant="outlined"
      name={name + "-port"}
      id={name + "-port"}
      defaultValue={targetResult.port}
      select
      placeholder='Select a port'
      onChange={(event) => {
        targetResult.port = event.target.value
        formik.setFieldValue(name, getTarget())
      }}
    >
      {portsOptions.map((option) => (
        <MenuItem key={option} value={option}>
          {option}
        </MenuItem>
      ))}
    </TextField></>) : ''}
    
    
    {(portsOptions.length > 0) ? (<>
      <Grid item xs={12}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
          <FormControlLabel
            type="checkbox"
            name={name + "-protocol"}
            control={<Checkbox size="large" />}
            onChange={(event) => {
              targetResult.protocol = event.target.checked ? "https" : "http"
              formik.setFieldValue(name, getTarget())
            }}
            label={"Container uses HTTPS internally (leave unchecked if not sure, they usually don't)"}
          />
        </Stack>
      </Grid>
    </>) : ''}

    <InputLabel htmlFor={name}>Result Target Preview</InputLabel>
    <TextField
      name={name}
      placeholder={"This will be generated automatically"}
      id={name}
      value={formik.values[name]}
      disabled={true}
    />
  </Stack>
  </Grid>
  );
}