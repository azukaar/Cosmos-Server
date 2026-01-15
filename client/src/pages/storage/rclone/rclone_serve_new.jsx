import React, { useState, useEffect } from 'react';
import { Formik, Form, Field } from 'formik';
import { TextField, Select, MenuItem, FormControl, InputLabel, Button, Typography, Stack, Alert, Dialog, DialogTitle, DialogContentText, DialogContent, DialogActions, Checkbox } from '@mui/material';
import { ServeConfig } from './rclone-serve';
import { CosmosCollapse, CosmosFormDivider } from '../../config/users/formShortcuts';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import { useTranslation } from 'react-i18next';
import { FilePickerButton } from '../../../components/filePicker';

const RCloneNewServeConfig = ({ onClose, initialValues }) => {
  const isEdit = typeof initialValues === 'object';
  const [selectedProvider, setSelectedProvider] = useState(isEdit ? initialValues.Protocol : '');
  const [providerOptions, setProviderOptions] = useState([]);
  const { t } = useTranslation();

  if (isEdit) {
    initialValues.named = initialValues.Name;
    initialValues.SmartShield = initialValues.Route.SmartShield.Enabled;
    let settings = initialValues.Settings;
    for (let key in settings) {
      initialValues[key] = settings[key];
    }
  }

  useEffect(() => {
    setProviderOptions(ServeConfig.map(config => ({
      value: config.Name,
      label: config.Name + ' - ' + config.Description
    })));
  }, []);

  const handleProviderChange = (event) => {
    setSelectedProvider(event.target.value);
  };

  const getSelectedProviderConfig = () => {
    return ServeConfig.find(config => config.Name === selectedProvider);
  };

  const renderFormFields = (config, isSubmitting, setFieldValue) => {
    const standardFields = [];
    const advancedFields = [
      <Field
        key="InternalTarget"
        as={TextField}
        name="InternalTarget"
        label="Internal Server Hostname"
        helperText="What port should the internal server listen on, do not change unless you know what you are doing"
        fullWidth
        margin="normal"
        disabled={isSubmitting}
      />
    ];

    config.Options.forEach(option => {
      const field = (
        <Field
          key={option.Name}
          as={TextField}
          name={option.Name}
          label={option.Name}
          helperText={option.Help}
          fullWidth
          margin="normal"
          required={option.Required}
          type={option.IsPassword ? 'password' : 'text'}
          disabled={isSubmitting}
        />
      );

      if (option.Advanced) {
        advancedFields.push(field);
      } else {
        standardFields.push(field);
      }
    });

    return (
      <>
        <Alert severity="info">
          {t('mgmt.storage.rclone.doclink')}&nbsp;<a href={`https://rclone.org/commands/rclone_serve_${config.Name}/`} target="_blank" rel="noreferrer">link</a>
        </Alert>
        <Field
          as={TextField}
          name="named"
          label="Name"
          helperText="Name of the remote for easy reference"
          fullWidth
          margin="normal"
          required
          disabled={isEdit}
        />
        <Stack direction="row" spacing={2}>
          <FilePickerButton raw onPick={(path) => setFieldValue('Target', path)} size="150%" select="folder" />
          <Field
            as={TextField}
            name="Target"
            label="Target Folder"
            helperText="What do you want to share"
            fullWidth
            margin="normal"
            required
          />
        </Stack>
          {selectedProvider === "samba" ? (<Alert severity="info">
            {t('mgmt.storage.rclone.samba_notice')}
          </Alert>) : <>
            <Field
              as={TextField}
              name="Source"
              label="Source Hostname"
              helperText="Hostname or IP address of the source"
              fullWidth
              margin="normal"
              required
            />
            <Stack direction="row" spacing={2} alignItems="center">
              <Field
                type="checkbox"
                name="SmartShield"
                as={Checkbox}
              />
              <div>{t('Enable Smart Shield')}</div>
            </Stack>
          </>}
          <CosmosFormDivider />
        {standardFields}
        {advancedFields.length > 0 && (
          <CosmosCollapse title="Advanced Options">
            {advancedFields}
          </CosmosCollapse>
        )}
      </>
    );
  };

  let provConf = ServeConfig.find(config => config.Name === selectedProvider) || {};

  return (
    <Dialog open={true} onClose={() => onClose()} fullWidth maxWidth="sm">
      <DialogTitle>{isEdit ? t('global.edit') : t('global.createAction')}</DialogTitle>
      <Formik
        initialValues={isEdit ? initialValues : {
          Source: provConf.DefaultSource,
          SmartShield: true,
        }}
        enableReinitialize={true}
        onSubmit={(values, { setSubmitting }) => {
          setSubmitting(true);

          let originalName;
          if (isEdit) {
            originalName = initialValues.Name;
          }

          console.log(values)

          let smartShieldEnabled = values.SmartShield;

          let fullValues = {};
          fullValues.Protocol = selectedProvider;
          fullValues.Name = values.named + "";
          fullValues.Target = values.Target;
          fullValues.Source = values.Source;
          fullValues.Settings = values;
          delete fullValues.Settings.named;
          delete fullValues.Settings.Settings;
          delete fullValues.Settings.Protocol;
          delete fullValues.Settings.Name;
          delete fullValues.Settings.Source;
          delete fullValues.Settings.Route;
          delete fullValues.Settings.Target;
          delete fullValues.Settings.SmartShield;

          return API.config.get().then(({data}) => {
            if(!data.RemoteStorage) data.RemoteStorage = {};

            let shares = data.RemoteStorage.Shares || [];
            if (isEdit) {
              shares = shares.filter((s) => s.Name !== originalName);
            }
            
            // if samba we do not add a route
            const isSamba = selectedProvider === "samba";

            if (!isSamba) {
              let nextFreePort = 12000;
              let busyPorts = [];
              shares.forEach((s) => {
                if (s.Route && s.Route.Target) {
                  let port = parseInt(s.Route.Target.split(":")[s.Route.Target.split(":").length - 1]);
                  if (!port.isNaN) {
                    busyPorts = [...busyPorts, port];
                  }
                }
              });
              
              while (busyPorts.includes(nextFreePort)) {
                nextFreePort++;
              }

              let scheme = ServeConfig.find(config => config.Name === selectedProvider).Proxy;

              let calculatedInternalTarget = scheme + "://127.0.0.1:" + nextFreePort;

              fullValues.Route = {
                Name: "netshare_" + fullValues.Name,
                Target: calculatedInternalTarget,
                Mode: "PROXY",
                UseHost: true,
                Host: fullValues.Source,
                SmartShield: {
                  Enabled: smartShieldEnabled
                }
              }
            } else {
              fullValues.Route = {Disabled: true};
            }

            shares.push(fullValues);

            return API.config.set({
              ...data,
              RemoteStorage: {
                ...data.RemoteStorage,
                Shares: shares
              }
            }).then(() => {
              return onClose();
            })
          });
        }}
      >
        {({ isSubmitting, setFieldValue }) => (
          <Form>
            <DialogContent>
              <DialogContentText>
                <FormControl fullWidth margin="normal">
                  <InputLabel>{t('mgmt.storage.rclone.select')}</InputLabel>
                  <Select
                    value={selectedProvider}
                    onChange={handleProviderChange}
                  >
                    {providerOptions.map(option => (
                      <MenuItem key={option.value} value={option.value}>
                        {option.label}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>

                {selectedProvider && (
                  <Stack spacing={2}>
                    <Typography variant="h6" gutterBottom>
                      {getSelectedProviderConfig().Description}
                    </Typography>
                    {renderFormFields(getSelectedProviderConfig(), isSubmitting, setFieldValue)}
                  </Stack>
                )}
              </DialogContentText>
            </DialogContent>
            <DialogActions>
              <LoadingButton type="submit" variant="contained" color="primary" disabled={isSubmitting} loading={isSubmitting}>
                {isEdit ? t('global.edit') : t('global.createAction')}
              </LoadingButton>
              <Button onClick={() => onClose()}>{t('global.cancelAction')}</Button>
            </DialogActions>
          </Form>
        )}
      </Formik>
    </Dialog>
  );
};

export default RCloneNewServeConfig;