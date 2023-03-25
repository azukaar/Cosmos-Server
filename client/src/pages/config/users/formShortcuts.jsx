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

} from '@mui/material';
import { Field } from 'formik';
import { DownOutlined, UpOutlined } from '@ant-design/icons';


export const CosmosInputText = ({ name, type, placeholder, label, formik }) => {
  return <Grid item xs={12}>
    <Stack spacing={1}>
      <InputLabel htmlFor={name}>{label}</InputLabel>
      <OutlinedInput
        id={name}
        type={type ? type : 'text'}
        value={formik.values[name]}
        name={name}
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        placeholder={placeholder}
        fullWidth
        error={Boolean(formik.touched[name] && formik.errors[name])}
      />
      {formik.touched[name] && formik.errors[name] && (
        <FormHelperText error id="standard-weight-helper-text-name-login">
          {formik.errors[name]}
        </FormHelperText>
      )}
    </Stack>
  </Grid>
}

export const CosmosSelect = ({ name, label, formik, options }) => {
  return <Grid item xs={12}>
    <Stack spacing={1}>
      <InputLabel htmlFor={name}>{label}</InputLabel>
      <TextField
        className="px-2 my-2"
        variant="outlined"
        name={name}
        id={name}
        select
        value={formik.values[name]}
        onChange={formik.handleChange}
        error={
          formik.touched[name] &&
          Boolean(formik.errors[name])
        }
        helperText={
          formik.touched[name] && formik.errors[name]
        }
      >
        {options.map((option) => (
          <MenuItem key={option[0]} value={option[0]}>
            {option[1]}
          </MenuItem>
        ))}
      </TextField>
    </Stack>
  </Grid>;
}

export const CosmosCheckbox = ({ name, label, formik }) => {
  return <Grid item xs={12}>
    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
      <Field
        type="checkbox"
        name={name}
        as={FormControlLabel}
        control={<Checkbox size="large" />}
        label={label}
      />
    </Stack>
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
            <Typography>{title}</Typography>
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
    <Chip label={title} />
    </Divider>
  </Grid>
}
