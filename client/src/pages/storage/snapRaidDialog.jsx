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
      name: data.Name || t('StorageParity'),
      parity: data.Parity || [],
      data: data.Data || {},
      syncCronTab: data.SyncCrontab || '0 0 2 * * *',
      scrubCronTab: data.ScrubCrontab || '0 0 4 */2 * *',
    },
    validateOnChange: false,
    validationSchema: yup.object({
      name: yup.string().required(t('Required')).min(3, t('NameAtLeast3Chars')).matches(/^[a-zA-Z0-9_ ]+$/, t('NameNotAlphanumeric')),
      parity: yup.array().min(1, t('Atleast1Parity')),
      data: yup.object().test('data', t('AtLeast2DataDisks'), (value) => {
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
              {data.Name ? (t('Edit') + " "+ data.Name) : t('CreateParityDisks')}
            </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info"><Trans i18nKey="CreateParityText">
                          You are about to create parity disks. <strong>This operation is safe and reversible</strong>.
                          Parity disks are used to protect your data from disk failure.
                          When creating a parity disk, the data disks you want to protect. Do not add a disk containing the system or another parity disk.
                        </Trans></Alert>
                      </div>
                      <TextField
                        fullWidth
                        id="name"
                        name="name"
                        label={t('ConfigName')}
                        value={formik.values.name}
                        onChange={formik.handleChange}
                        error={formik.touched.name && Boolean(formik.errors.name)}
                        helperText={formik.touched.name && formik.errors.name}
                      />
                      <FormLabel>
                        <strong>{t('Step1')}</strong> {t('ParityStep1')}
                      </FormLabel>
                      <MountPicker onChange={(value) => formik.setFieldValue('parity', value)} value={formik.values.parity} />
                      {formik.errors.parity && (
                          <Grid item xs={12}>
                              <FormHelperText error>{formik.errors.parity}</FormHelperText>
                          </Grid>
                      )}
                      <FormLabel>
                        <strong>{t('Step2')}</strong> {t('ParityStep2')}
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
                        }}>{t('AddDataDisk')}</Button>
                        <Button onClick={() => {
                          const data = formik.values.data;
                          delete data[Object.keys(data).pop()];
                          formik.setFieldValue('data', data);
                        }}>{t('RemoveDataDisk')}</Button>
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
                          <strong>{t('Step3')}</strong> {t('ParityStep3')}
                        </FormLabel>
                        <TextField
                          fullWidth
                          id="syncCronTab"
                          name="syncCronTab"
                          label={t('SyncInterval')}
                          value={formik.values.syncCronTab}
                          onChange={formik.handleChange}
                          error={formik.touched.syncCronTab && Boolean(formik.errors.syncCronTab)}
                          helperText={formik.touched.syncCronTab && formik.errors.syncCronTab}
                        />
                        <FormLabel>
                          {crontabToText(formik.values.syncCronTab)}
                        </FormLabel>
                        <TextField
                          fullWidth
                          id="scrubCronTab"
                          name="scrubCronTab"
                          label={t('ScrubInterval')}
                          value={formik.values.scrubCronTab}
                          onChange={formik.handleChange}
                          error={formik.touched.scrubCronTab && Boolean(formik.errors.scrubCronTab)}
                          helperText={formik.touched.scrubCronTab && formik.errors.scrubCronTab}
                        />
                        <FormLabel>
                          {crontabToText(formik.values.scrubCronTab)}
                        </FormLabel>
                      </Stack>
                    </Stack>
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setOpen(false)}>{t('Cancel')}</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>
                    {data.Name ? t('Update') : t('Create')}
                  </LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const SnapRAIDDialog = ({ refresh, data }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return <>
    {open && <SnapRAIDDialogInternal refresh={refresh} open={open} setOpen={setOpen} data={data} />}
    
    <div>
      {!data ? <ResponsiveButton
        onClick={() => {setOpen(true);}}
        variant="contained"
        size="small"
        startIcon={<PlusCircleOutlined />}
      >{t('NewParityDisks')}</ResponsiveButton> :
      <div onClick={() => setOpen(true)}>{t('Edit')}</div>}
    </div>
  </>
}

export default SnapRAIDDialog;
export { SnapRAIDDialog, SnapRAIDDialogInternal };