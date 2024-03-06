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

const SnapRAIDDialogInternal = ({ refresh, open, setOpen }) => {  
  const formik = useFormik({
    initialValues: {
      name: 'Storage Parity',
      parity: [],
      data: [],
      syncCronTab: '* 0 2 * * *',
      scrubCronTab: '* 0 4 * * */2',
    },
    validateOnChange: false,
    validationSchema: yup.object({
      name: yup.string().required('Required').min(3, 'Name should be at least 3 characters').matches(/^[a-zA-Z0-9_ ]+$/, 'Name should be alphanumeric'),
      parity: yup.array().min(1, 'Select at least 1 parity disk'),
      data: yup.array().min(2, 'Select at least 2 data disks'),
    }),
    onSubmit: (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      return API.storage.snapRAID.create({
        enabled: true,
        name: values.name,
        parity: values.parity,
        data: values.data,
        syncCronTab: values.syncCronTab,
        scrubCronTab: values.scrubCronTab,
      }).then((res) => {
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
            <DialogTitle>Create Parity Disks</DialogTitle>
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
                      <MountPicker onChange={(value) => formik.setFieldValue('parity', value)} />
                      {formik.errors.parity && (
                          <Grid item xs={12}>
                              <FormHelperText error>{formik.errors.parity}</FormHelperText>
                          </Grid>
                      )}
                      <FormLabel>
                        <strong>Step 2</strong>: Select the data disks you want to protect with the parity disk(s).
                      </FormLabel>
                      <MountPicker onChange={(value) => formik.setFieldValue('data', value)} />
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
                      <div>
                        <Checkbox /> Automatically map a mergerfs pool to the data disks (allows you to access the data disks as a single folder)
                      </div>
                    </Stack>
                  </DialogContentText>
                </DialogContent>
                <DialogActions>
                  <Button onClick={() => setOpen(false)}>Cancel</Button>
                  <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                    formik.handleSubmit();
                  }}>Create</LoadingButton>
                </DialogActions>
            </form>
        </FormikProvider>
    </Dialog>
  </>
}

const SnapRAIDDialog = ({ refresh }) => {
  const [open, setOpen] = useState(false);

  return <>
    {open && <SnapRAIDDialogInternal refresh={refresh} open={open} setOpen={setOpen}/>}
    
    <div>
      <ResponsiveButton
        onClick={() => {setOpen(true);}}
        variant="contained"
        size="small"
        startIcon={<PlusCircleOutlined />}
      >New Parity Disks</ResponsiveButton>
    </div>
  </>
}

export default SnapRAIDDialog;