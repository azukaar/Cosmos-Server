import { LoadingButton } from "@mui/lab";
import { Alert, Grid, Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, FormHelperText, Stack, TextField } from "@mui/material";
import React, { useState } from "react";
import { FormikProvider, useFormik } from "formik";
import * as yup from "yup";
import * as API from '../../api';
import ResponsiveButton from "../../components/responseiveButton";
import { PlusCircleOutlined, UpOutlined } from "@ant-design/icons";
import { crontabToText } from "../../utils/indexs";
import { Trans, useTranslation } from 'react-i18next';
import { FilePickerButton } from '../../components/filePicker';
import { CosmosInputText } from "../config/users/formShortcuts";

const isAbsolutePath = (path) => path.startsWith('/') || /^[a-zA-Z]:\\/.test(path); // Unix & Windows support

const RestoreDialogInternal = ({ refresh, open, setOpen, candidatePaths, originalSource, selectedSnapshot, backupName }) => {
  const { t } = useTranslation();
  const [done, setDone] = useState(false);

  const formik = useFormik({
    initialValues: {
      target: originalSource || '',
    },
    validationSchema: yup.object({
          target: yup
            .string()
            .required(t('global.required'))
            .test('is-absolute', t('global.absolutePath'), isAbsolutePath),
        }),
    onSubmit: async (values, { setErrors, setStatus, setSubmitting }) => {
      console.log(values);
      setSubmitting(true);

      await API.backups.restoreBackup(backupName, {
        snapshotID: selectedSnapshot,
        target: values.target,
        include: candidatePaths,
      });

      setDone(true);
    },
  });

  return (
    <Dialog open={open} onClose={() => setOpen(false)}>
      <FormikProvider value={formik}>
        <form onSubmit={formik.handleSubmit}>
          <DialogTitle>
            Restore Snapshot
          </DialogTitle>
          <DialogContent>
            <DialogContentText>
              {!done ? <Stack spacing={3} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                <Alert severity="info">
                  You are about to restore {candidatePaths.length ? candidatePaths.length : "every"} file(s)/folder(s). This will overwrite existing files.
                </Alert>
                
                <Stack direction="row" spacing={2} alignItems="flex-end">
                  <FilePickerButton
                    canCreate
                    onPick={(path) => {
                      if(path)
                        formik.setFieldValue('target', path);
                    }} 
                    size="150%" 
                    select="folder" 
                  />
                  <CosmosInputText
                    name="target"
                    label={"Where do you want to restore to?"}
                    placeholder="/path/to/restore"
                    formik={formik}
                  />
                </Stack>

                {formik.errors.submit && (
                  <Grid item xs={12}>
                    <FormHelperText error>{formik.errors.submit}</FormHelperText>
                  </Grid>
                )}
              </Stack> : <Stack>
                <Alert severity="success">
                  Restore operation has been queued. You can monitor the progress in the <strong>Tasks</strong> log on the top right.
                </Alert>
              </Stack>}
            </DialogContentText>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setOpen(false)}>{t('global.cancelAction')}</Button>
            <LoadingButton 
              color="primary" 
              variant="contained" 
              type="submit" 
              loading={formik.isSubmitting}
            >
              Restore
            </LoadingButton>
          </DialogActions>
        </form>
      </FormikProvider>
    </Dialog>
  );
};

const RestoreDialog = ({ refresh, candidatePaths, originalSource, selectedSnapshot, backupName }) => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);

  return (
    <>
      {open && <RestoreDialogInternal refresh={refresh} open={open} setOpen={setOpen} originalSource={originalSource} candidatePaths={candidatePaths} selectedSnapshot={selectedSnapshot} backupName={backupName} />}
      <div>
          <ResponsiveButton
            onClick={() => setOpen(true)}
            variant="contained"
            size="small"
            startIcon={<UpOutlined />}
          >
            {candidatePaths.length === 0 ? 'Restore All' : 'Restore Selected'}
          </ResponsiveButton>
      </div>
    </>
  );
};

export default RestoreDialog;
export { RestoreDialog, RestoreDialogInternal };