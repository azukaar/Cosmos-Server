import dnsList from './dns-list.json';
import dnsConfig from './dns-config.json';

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
    Alert,
  
  } from '@mui/material';
  import { Field } from 'formik';
  import { DownOutlined, UpOutlined } from '@ant-design/icons';

  import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import { useState } from 'react';
import { CosmosCollapse, CosmosSelect } from '../pages/config/users/formShortcuts';
  

export const DnsChallengeComp = ({ name, configName, style, multiline, type, placeholder, onChange, label, formik }) => {
    const filterVars = (obj) => {
      const newObj = {};
      Object.keys(obj).forEach((key) => {
        if (obj[key] !== '' && dnsConfig[formik.values[name]].vars.includes(key)) {
          newObj[key] = obj[key];
        }
      });
      return newObj;
    };

    return <><CosmosSelect
      name={name}
      label={label}
      formik={formik}
      onChange={(e) => {
        onChange && onChange(e);
      }}
      options={[["", "DISABLE"], ...(dnsList).map((dns) => ([dns,dns]))]}
    />

      <Grid item xs={12}>
        <Stack spacing={2}>
        {formik.values[name] && dnsConfig[formik.values[name]] &&<>
          {dnsConfig[formik.values[name]].vars.length > 0 && <CosmosCollapse title="DNS Challenge setup">
            
          <Stack spacing={2}>
          <Alert severity="info">
            Please be careful you are filling the correct values. Check the doc if unsure. Leave blank unused variables. <br />
            Doc link: <a href={dnsConfig[formik.values[name]].url} rel="noopener noreferrer" target="_blank">{dnsConfig[formik.values[name]].url}</a>
          </Alert>
          <div className="raw-table">
            <div dangerouslySetInnerHTML={{__html: dnsConfig[formik.values[name]].docs}}></div>
          </div>
          {dnsConfig[formik.values[name]].vars.map((dnsVar) => <div>
            {dnsVar}:
              <OutlinedInput
                type={type ? type : 'text'}
                value={formik.values[configName] ? (formik.values[configName][dnsVar] || '') : ''}
                onChange={(...ar) => {
                  const newConfig = {
                    ...formik.values[configName],
                    [dnsVar]: ar[0].target.value
                  };
                  formik.setFieldValue(configName, filterVars(newConfig));
                }}
                placeholder={"leave blank if unused"}
                fullWidth
              />
          </div>)}
          </Stack>
          </CosmosCollapse>}
        </>}
        </Stack>
      </Grid>
    </>;
  }