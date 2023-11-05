import * as React from 'react';
import {
  Checkbox,
  Divider,
  FormControlLabel,
  Grid,
  InputLabel,
  OutlinedInput,
  Stack,
  Typography,
  FormHelperText,
  TextField,
  MenuItem,
  AccordionSummary,
  AccordionDetails,
  Accordion,
  Chip,
  Box,
  FormControl,
  IconButton,
  InputAdornment,
  Autocomplete,

} from '@mui/material';
import { Field } from 'formik';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import * as API from '../../api';

import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';

export const MetricPicker = ({ metricsInit, name, style, value, errors, placeholder, onChange, label, formik }) => {
  const [metrics, setMetrics] = React.useState(metricsInit || {});

  function refresh() {
    API.metrics.list().then((res) => {
      let m = [];
      let wildcards = {};
      
      Object.keys(res.data).forEach((key) => {
        m.push({
          label: res.data[key],
          value: key,
        });

        let keysplit = key.split('.');
        if (keysplit.length > 1) {
          for (let i = 0; i < keysplit.length - 1; i++) {
            let wildcard = keysplit.slice(0, i + 1).join('.') + '.*';
            wildcards[wildcard] = true;
          }
        }
      });

      Object.keys(wildcards).forEach((key) => {
        m.push({
          label: "Wildcard for " + key.split('.*')[0],
          value: key,
        });
      });

      setMetrics(m);
    });
  }

  React.useEffect(() => {
    if (!metricsInit)
      refresh();
  }, []);
  
  return <Grid item xs={12}>
    <Stack spacing={1} style={style}>
      {label && <InputLabel htmlFor={name}>{label}</InputLabel>}
      {/* <OutlinedInput
        id={name}
        type={'text'}
        value={value || (formik && formik.values[name])}
        name={name}
        onBlur={(...ar) => {
          return formik && formik.handleBlur(...ar);
        }}
        onChange={(...ar) => {
          onChange && onChange(...ar);
          return formik && formik.handleChange(...ar);
        }}
        placeholder={placeholder}
        fullWidth
        error={Boolean(formik && formik.touched[name] && formik.errors[name])}
      /> */}

      <Autocomplete
        disablePortal
        name={name}
        value={value || (formik && formik.values[name])}
        id="combo-box-demo"
        isOptionEqualToValue={(option, value) => option.value === value}
        options={metrics}
        freeSolo
        getOptionLabel={(option) => {
          return option.label ?
            `${option.value} - ${option.label}` : (formik && formik.values[name]);
        }}
        onChange={(event, newValue) => {
          onChange && onChange(newValue.value);
          return formik && formik.setFieldValue(name, newValue.value);
        }}
        renderInput={(params) => <TextField {...params} />}
      />

      {formik && formik.touched[name] && formik.errors[name] && (
        <FormHelperText error id="standard-weight-helper-text-name-login">
          {formik.errors[name]}
        </FormHelperText>
      )}
      {errors && (
        <FormHelperText error id="standard-weight-helper-text-name-login">
          {formik.errors[name]}
        </FormHelperText>
      )}
    </Stack>
  </Grid>
}