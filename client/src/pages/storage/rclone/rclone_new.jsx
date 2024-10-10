import React, { useState, useEffect } from 'react';
import { Formik, Form, Field } from 'formik';
import { TextField, Select, MenuItem, FormControl, InputLabel, Button, Typography, Stack, Alert, Dialog, DialogTitle, DialogContentText, DialogContent, DialogActions } from '@mui/material';
import { ProvConfig, ProvAuth } from './rclone-providers';
import { CosmosCollapse } from '../../config/users/formShortcuts';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import { useTranslation } from 'react-i18next';

const RCloneNewConfig = ({ onClose }) => {
  const [selectedProvider, setSelectedProvider] = useState('');
  const [providerOptions, setProviderOptions] = useState([]);
  const { t } = useTranslation();

  useEffect(() => {
    setProviderOptions(ProvConfig.map(config => ({
      value: config.Name,
      label: config.Name + ' - ' + config.Description
    })));
  }, []);

  const handleProviderChange = (event) => {
    setSelectedProvider(event.target.value);
  };

  const getSelectedProviderConfig = () => {
    return ProvConfig.find(config => config.Name === selectedProvider);
  };

  const renderFormFields = (config, isSubmitting, setFieldValue) => {
    const standardFields = [];
    const advancedFields = [];

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
          Documentation for this provider can be found <a href={`https://rclone.org/${config.Name}/`} target="_blank" rel="noreferrer">here</a>
        </Alert>
        <Field
          as={TextField}
          name="named"
          label="Name"
          helperText="Name of the remote for easy reference"
          fullWidth
          margin="normal"
          required
        />
        {standardFields}
        {ProvAuth.includes(config.Name) && (<>
          <Alert severity="info">
            This provider requires interactive authentication. In order to proceed do BLABLABLABLABLA
          </Alert>
          <Field
            as={TextField}
            name="token"
            label="Token"
            helperText="Token"
            fullWidth
            margin="normal"
            required
          /></>)}
        {advancedFields.length > 0 && (
          <CosmosCollapse title="Advanced Options">
            {advancedFields}
          </CosmosCollapse>
        )}
      </>
    );
  };

  return (
    <Dialog open={true} onClose={() => onClose()} fullWidth maxWidth="sm">
      <DialogTitle>{t('mgmt.storage.rclone.create')}</DialogTitle>
      <Formik
        initialValues={{}}
        onSubmit={(values, { setSubmitting }) => {
          let fullValues = {};
          fullValues.type = selectedProvider;
          fullValues.name = values.named + "";
          delete values.named;
          fullValues.parameters = values;
          setSubmitting(false);

          API.rclone.create(fullValues).then(() => {
            onClose();
          }).catch((err) => {
            onClose();
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
                {t('global.createAction')}
              </LoadingButton>
              <Button onClick={() => onClose()}>{t('global.cancelAction')}</Button>
            </DialogActions>
          </Form>
        )}
      </Formik>
    </Dialog>
  );
};

export default RCloneNewConfig;