import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Checkbox, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField, FormLabel } from "@mui/material";
import React, {useState} from "react";
import { CosmosCheckbox } from "../config/users/formShortcuts";
import { Formik, FormikProvider, useFormik } from "formik";
import * as yup from "yup";

import * as API from '../../api';
import { MountPicker } from "./mountPicker";
import ResponsiveButton from "../../components/responseiveButton";
import { PlusCircleOutlined } from "@ant-design/icons";
import { crontabToText } from "../../utils/indexs";
import { MountPickerEntry } from "./mountPickerEntry";
import { json } from "react-router";
import { Trans, useTranslation } from 'react-i18next';

const SnapRAIDDialogInternal = ({ refresh, open, setOpen, data = {}}) => {
  const { t } = useTranslation();
  const formik = useFormik({
    initialValues: {
      name: data.Name || t('mgmt.storage.snapraid.storageParity'),
      parity: data.Parity || [],
      data: data.Data || {},
      syncCronTab: data.SyncCrontab || '0 0 2 * * *',
      scrubCronTab: data.ScrubCrontab || '0 0 4 */2 * *',
    },
    validateOnChange: false,
    validationSchema: yup.object({
      name: yup.string().required(t('global.required')).min(3, t('mgmt.storage.snapraid.min3chars')).matches(/^[a-zA-Z0-9_ ]+$/, t('mgmt.storage.snapraid.notAlphanumeric')),
      parity: yup.array().min(1, t('mgmt.storage.snapraid.min1parity')),
      data: yup.object().test('data', t('mgmt.storage.snapraid.min2datadisks'), (value) => {
        return Object.keys(value).length >= 2;
      })
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      (data.Name ? API.storage.snapRAID.update(data.Name, {
        enabled: true,
        name: values.name,
        parity: values.parity,
        data: values.data,
        syncCronTab: values.syncCronTab,
        scrubCronTab: values.scrubCronTab,
      }) : API.storage.snapRAID.create({
        enabled: true,
        name: values.name,
        parity: values.parity,
        data: values.data,
        syncCronTab: values.syncCronTab,
        scrubCronTab: values.scrubCronTab,
      })).then((res) => {
        setStatus({ success: true });
        setSubmitting(false);
        setOpen(false);
        refresh && refresh();
      }).catch((err) => {
        setStatus({ success: false });
        setErrors({ submit: err.message });
        setSubmitting(false);
      });
    },
  });

  return <>
    <Dialog open={open} onClose={() => setOpen(false)}>
          <FormikProvider value={formik}>
            <form onSubmit={formik.handleSubmit}>
            <DialogTitle>
              {data.Name ? (t('global.edit') + " "+ data.Name) : t('mgmt.storage.snapraid.createParityDisksButton')}
            </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info">
                          <Trans i18nKey="mgmt.storage.snapraid.createParityInfo" />
                        </Alert>
                      </div>
                      <TextField
                        fullWidth
                        id="name"
                        name="name"
                        label={t('mgmt.storage.configName.configNameLabel')}
                        value={formik.values.name}
                        onChange={formik.handleChange}
                        error={formik.touched.name && Boolean(formik.errors.name)}
                        helperText={formik.touched.name && formik.errors.name}
                      />
                      <FormLabel>
                        <strong>{t('mgmt.storage.snapraid.createParity.step')} 1:</strong> {t('mgmt.storage.snapraid.createParity.Step1Text')}
                      </FormLabel>
                      <MountPicker onChange={(value) => formik.setFieldValue('parity', value)} value={formik.values.parity} />
                      {formik.errors.parity && (
                          <Grid item xs={12}>
                              <FormHelperText error>{formik.errors.parity}</FormHelperText>
                          </Grid>
                      )}
                      <FormLabel>
                        <strong>{t('mgmt.storage.snapraid.createParity.step')} 2:</strong> {t('mgmt.storage.snapraid.createParity.Step2Text')}
                      </FormLabel>
                      

                      {Object.keys(formik.values.data).map((dataName) => {
                        const dataDisk = formik.values.data[dataName];
                        
                        return <MountPickerEntry
                            id={`data.${dataName}`}
                            name={`data.${dataName}`}
                            label={`${dataName}`}
                            value={dataDisk}
                            onChange={formik.handleChange}
                            formik={formik}
                            touched={formik.touched.data && formik.touched.data[dataName]}
                            error={formik.touched.data && formik.touched.data[dataName] && Boolean(formik.errors.data)}
                            helperText={formik.touched.data && formik.touched.data[dataName] && formik.errors.data}
                          />
                      })}
                      
                      <Stack direction="row" spacing={2}>
                        <Button onClick={() => {
                          const data = formik.values.data;
                          data[`disk${Object.keys(data).length}`] = '';
                          formik.setFieldValue('data', data);
                        }}>{t('mgmt.storage.snapraid.addDatadisk')}</Button>
                        <Button onClick={() => {
                          const data = formik.values.data;
                          delete data[Object.keys(data).pop()];
                          formik.setFieldValue('data', data);
                        }}>{t('mgmt.storage.snapraid.removeDatadisk')}</Button>
                      </Stack>

                      {formik.errors.data && (
                          <Grid item xs={12}>
                              <FormHelperText error>{formik.errors.data}</FormHelperText>
                          </Grid>
                      )}
                      {formik.errors.submit && (
                        <Grid item xs={12}>
                          <FormHelperText error>{formik.errors.submit}</FormHelperText>
                        </Grid>
                      )}
                      <Stack spacing={2}>
                        <FormLabel>
                          <strong>{t('mgmt.storage.snapraid.createParity.step')} 3:</strong> {t('mgmt.storage.snapraid.createParity.Step3Text')}
                        </FormLabel>
                        <TextField
                          fullWidth
                          id="syncCronTab"
                          name="syncCronTab"
                          label={t('mgmt.storage.snapraid.syncInterval.syncIntervalLabel')}
                          value={formik.values.syncCronTab}
                          onChange={formik.handleChange}
                          error={formik.touched.syncCronTab && Boolean(formik.errors.syncCronTab)}
                          helperText={formik.touched.syncCronTab && formik.errors.syncCronTab}
                        />
                        <FormLabel>
                          {crontabToText(formik.values.syncCronTab, t)}
                        </FormLabel>
                        <TextField
                          fullWidth
                          id="scrubCronTab"
                          name="scrubCronTab"
                          label={t('mgmt.storage.snapraid.scrubInterval.scrubIntervalLabel')}
                          value={formik.values.scrubCronTab}
                          onChange={formik.handleChange}
                          error={formik.touched.scrubCronTab && Boolean(formik.errors.scrubCronTab)}
                          helperText={formik.touched.scrubCronTab && formik.errors.scrubCronTab}
                        />
                        <FormLabel>
                          {crontabToText(formik.values.scrubCronTab, t)}
                        </FormLabel>
                      </Stack>
                    </Stack>
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setOpen(false)}>{t('global.cancelAction')}</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>
                    {data.Name ? t('global.update') : t('global.createAction')}
                  </LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const SnapRAIDDialog = ({ refresh, data, disabled }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return <>
    {open && <SnapRAIDDialogInternal refresh={refresh} open={open} setOpen={setOpen} data={data} />}
    
    <div>
      {!data ? <ResponsiveButton
        onClick={() => {setOpen(true);}}
        variant="contained"
        size="small"
        disabled={disabled}
        startIcon={<PlusCircleOutlined />}
      >{t('mgmt.storage.snapraid.createParity.newDisks')}</ResponsiveButton> :
      <div onClick={() => setOpen(true)}>{t('global.edit')}</div>}
    </div>
  </>
}

export default SnapRAIDDialog;
export { SnapRAIDDialog, SnapRAIDDialogInternal };