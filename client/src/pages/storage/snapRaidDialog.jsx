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

const SnapRAIDDialogInternal = ({ refresh, open, setOpen, data = {}}) => {  
  const formik = useFormik({
    initialValues: {
      name: data.Name || 'Storage Parity',
      parity: data.Parity || [],
      data: data.Data || {},
      syncCronTab: data.SyncCrontab || '0 0 2 * * *',
      scrubCronTab: data.ScrubCrontab || '0 0 4 */2 * *',
    },
    validateOnChange: false,
    validationSchema: yup.object({
      name: yup.string().required('Required').min(3, 'Name should be at least 3 characters').matches(/^[a-zA-Z0-9_ ]+$/, 'Name should be alphanumeric'),
      parity: yup.array().min(1, 'Select at least 1 parity disk'),
      data: yup.object().test('data', 'Select at least 2 data disk', (value) => {
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
              {data.Name ? ('Edit ' + data.Name) : 'Create Parity Disks'}
            </DialogTitle>
                <DialogContent>
                  <DialogContentText>
                    <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                      <div>
                        <Alert severity="info">
                          You are about to create parity disks. <strong>This operation is safe and reversible</strong>.
                          Parity disks are used to protect your data from disk failure.
                          When creating a parity disk, the data disks you want to protect. Do not add a disk containing the system or another parity disk.
                        </Alert>
                      </div>
                      <TextField
                        fullWidth
                        id="name"
                        name="name"
                        label="Config Name"
                        value={formik.values.name}
                        onChange={formik.handleChange}
                        error={formik.touched.name && Boolean(formik.errors.name)}
                        helperText={formik.touched.name && formik.errors.name}
                      />
                      <FormLabel>
                        <strong>Step 1</strong>: First, select the parity disk(s). One parity disk will protect against one disk failure, two parity disks will protect against two disk failures, and so on.
                        Remember that those disks will be used only for parity, and will not be available for data storage.
                        Parity disks must be at least as large as the largest data disk, and should be empty.
                      </FormLabel>
                      <MountPicker onChange={(value) => formik.setFieldValue('parity', value)} value={formik.values.parity} />
                      {formik.errors.parity && (
                          <Grid item xs={12}>
                              <FormHelperText error>{formik.errors.parity}</FormHelperText>
                          </Grid>
                      )}
                      <FormLabel>
                        <strong>Step 2</strong>: Select the data disks you want to protect with the parity disk(s).
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
                        }}>Add Data Disk</Button>
                        <Button onClick={() => {
                          const data = formik.values.data;
                          delete data[Object.keys(data).pop()];
                          formik.setFieldValue('data', data);
                        }}>Remove Data Disk</Button>
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
                          <strong>Step 3</strong>: Set the sync and scrub intervals. The sync interval is the time at which parity is updated. The scrub interval is the time at which the parity is checked for errors. This is using the CRONTAB syntax with seconds.
                        </FormLabel>
                        <TextField
                          fullWidth
                          id="syncCronTab"
                          name="syncCronTab"
                          label="Sync Interval"
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
                          label="Scrub Interval"
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
                  <Button onClick={() => setOpen(false)}>Cancel</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>
                    {data.Name ? 'Update' : 'Create'}
                  </LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const SnapRAIDDialog = ({ refresh, data, disabled }) => {
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
      >New Parity Disks</ResponsiveButton> :
      <div onClick={() => setOpen(true)}>Edit</div>}
    </div>
  </>
}

export default SnapRAIDDialog;
export { SnapRAIDDialog, SnapRAIDDialogInternal };