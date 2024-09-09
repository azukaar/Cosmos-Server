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
  Select,

} from '@mui/material';
import { Field } from 'formik';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import { strengthColor, strengthIndicator } from '../../../utils/password-strength';
import { useTranslation } from 'react-i18next';

import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';

export const getNestedValue = (values, path) => {
  return path.split('.').reduce((current, key) => {
    if (current && current[key] !== undefined) {
      return current[key];
    }
    if (Array.isArray(current)) {
      const index = parseInt(key, 10);
      return current[index];
    }
    return undefined;
  }, values);
};

export const CosmosInputText = ({ name, style, value, errors, multiline, type, placeholder, onChange, label, formik }) => {
  return <Grid item xs={12}>
    <Stack spacing={1} style={style}>
      {label && <InputLabel htmlFor={name}>{label}</InputLabel>}
      <OutlinedInput
        id={name}
        type={type ? type : 'text'}
        value={value || (formik && getNestedValue(formik.values, name))}
        name={name}
        multiline={multiline}
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

export const CosmosInputPassword = ({ name, noStrength, type, placeholder, autoComplete, onChange, label, formik }) => {
  const { t } = useTranslation();
  const [level, setLevel] = React.useState();
  const [showPassword, setShowPassword] = React.useState(false);
  const handleClickShowPassword = () => {
      setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event) => {
      event.preventDefault();
  };

  const changePassword = (value) => {
      const temp = strengthIndicator(value);
      setLevel(strengthColor(temp, t));
  }; 
  
  React.useEffect(() => {
      changePassword('');
  }, []);

  return <Grid item xs={12}>
    <Stack spacing={1}>
      <InputLabel htmlFor={name}>{label}</InputLabel>
      <OutlinedInput
        id={name}
        type={showPassword ? 'text' : 'password'}
        value={getNestedValue(formik.values, name)}
        name={name}
        autoComplete={autoComplete}
        onBlur={formik.handleBlur}
        onChange={(e) => {
          changePassword(e.target.value);
          formik.handleChange(e);
          onChange && onChange(e);
        }}
        placeholder="********"
        endAdornment={
            <InputAdornment position="end">
                <IconButton
                    aria-label="toggle password visibility"
                    onClick={handleClickShowPassword}
                    onMouseDown={handleMouseDownPassword}
                    edge="end"
                    size="large"
                >
                    {showPassword ? <EyeOutlined /> : <EyeInvisibleOutlined />}
                </IconButton>
            </InputAdornment>
        }
        fullWidth
        error={Boolean(formik.touched[name] && formik.errors[name])}
      />
      {formik.touched[name] && formik.errors[name] && (
        <FormHelperText error id="standard-weight-helper-text-name-login">
          {formik.errors[name]}
        </FormHelperText>
      )}

      {!noStrength && <FormControl fullWidth sx={{ mt: 2 }}>
          <Grid container spacing={2} alignItems="center">
              <Grid item>
                  <Box sx={{ bgcolor: level?.color, width: 85, height: 8, borderRadius: '7px' }} />
              </Grid>
              <Grid item>
                  <Typography variant="subtitle1" fontSize="0.75rem">
                      {level?.label}
                  </Typography>
              </Grid>
          </Grid>
      </FormControl>}
    </Stack>
  </Grid>
}


export const CosmosSelect = ({ name, onChange, label, formik, disabled, options, style }) => {
  return (
    <Grid item xs={12}>
      <Stack spacing={1}>
        <InputLabel htmlFor={name}>{label}</InputLabel>
        <FormControl variant="outlined" fullWidth error={getNestedValue(formik.touched, name) && Boolean(getNestedValue(formik.errors, name))} disabled={disabled} style={style}>
          <Select
            className="px-2 my-2"
            variant="outlined"
            name={name}
            id={name}
            disabled={disabled}
            value={getNestedValue(formik.values, name) || ''}
            displayEmpty
            onChange={(...ar) => {
              onChange && onChange(...ar);
              formik.handleChange(...ar);
            }}
          >
            {options.map((option) => (
              <MenuItem key={option[0]} value={option[0]}>
                {option[1]}
              </MenuItem>
            ))}
          </Select>
          {getNestedValue(formik.touched, name) && getNestedValue(formik.errors, name) && (
            <FormHelperText>{getNestedValue(formik.errors, name)}</FormHelperText>
          )}
        </FormControl>
      </Stack>
    </Grid>
  );
};

export const CosmosCheckbox = ({ name, label, formik, style }) => {
  return <Grid item xs={12}>
    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
      <Field
        type="checkbox"
        name={name}
        as={FormControlLabel}
        control={<Checkbox size="large" />}
        label={label}
        style={style}
      />
    </Stack>
    {formik.touched[name] && formik.errors[name] && (
      <FormHelperText error id="standard-weight-helper-text-name-login">
        {formik.errors[name]}
      </FormHelperText>
    )}
  </Grid>
}

export const CosmosCollapse = ({ children, title }) => {
  const [open, setOpen] = React.useState(false);
  return <Grid item xs={12}>
    <Stack spacing={1}>
      <Accordion>
          <AccordionSummary
            expandIcon={<DownOutlined />}
            aria-controls="panel1a-content"
            id="panel1a-header"
          >
            <Typography variant="h6" style={{width: '100%', marginRight: '20px'}}>
              {title}
            </Typography>
          </AccordionSummary>
          <AccordionDetails>
            {children}
          </AccordionDetails>
        </Accordion>
    </Stack>
  </Grid>
}

export function CosmosFormDivider({title}) {
  return <Grid item xs={12}>
    <Divider>
    {title && <Chip label={title} />}
    </Divider>
  </Grid>
}
