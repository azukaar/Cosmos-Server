import React, { useState, useEffect } from 'react';
import { Formik, Form, Field } from 'formik';
import { TextField, Select, MenuItem, FormControl, InputLabel, Button, Typography, Stack, Alert, Dialog, DialogTitle, DialogContentText, DialogContent, DialogActions } from '@mui/material';
import { ProvConfig, ProvAuth } from './rclone-providers';
import { CosmosCollapse } from '../../config/users/formShortcuts';
import * as API from '../../../api';
import { LoadingButton } from '@mui/lab';
import { useTranslation } from 'react-i18next';

const RCloneNewConfig = ({ onClose, initialValues }) => {
  const isEdit = typeof initialValues === 'object';
  const [selectedProvider, setSelectedProvider] = useState(isEdit ? initialValues.type : '');
  const [providerOptions, setProviderOptions] = useState([]);
  const { t } = useTranslation();

  if (isEdit)
    initialValues.named = initialValues.name;

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
    const advancedFields = [
      <Field
        key="cosmos-chown"
        as={TextField}
        name="cosmos-chown"
        label="Chown (permission will be 700)"
        fullWidth
        margin="normal"
        required
      />,
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
          {t('mgmt.storage.rclone.doclink')}&nbsp;<a href={`https://rclone.org/${config.Name}/`} target="_blank" rel="noreferrer">link</a>
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
        {standardFields}
        {ProvAuth.includes(config.Name) && (<>
          <Alert severity="info">
            {t('mgmt.storage.rclone.tokenInfo')}&nbsp;<a href={`https://cosmos-cloud.io/doc/81%20Storage/#remote-storage`} target="_blank" rel="noreferrer">link</a> 
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
      <DialogTitle>{isEdit ? t('global.edit') : t('global.createAction')}</DialogTitle>
      <Formik
        initialValues={isEdit ? initialValues : {
          "cosmos-chown": "1000:1000",
        }}
        onSubmit={(values, { setSubmitting }) => {
          let originalName;
          if (isEdit) {
            originalName = initialValues.name;
          }

          let fullValues = {};
          fullValues.type = selectedProvider;
          fullValues.name = values.named + "";
          fullValues.parameters = values;
          delete fullValues.parameters.named;
          delete fullValues.parameters.parameters;
          delete fullValues.parameters.type;

          setSubmitting(false);

          (isEdit ? 
            API.rclone.update(fullValues)
            : API.rclone.create(fullValues)).then(() => {
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

export default RCloneNewConfig;