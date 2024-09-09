import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CheckCircleOutlined, CloudOutlined, CompassOutlined, DesktopOutlined, ExpandOutlined, InfoCircleOutlined, LaptopOutlined, MenuFoldOutlined, MenuOutlined, MinusCircleFilled, MobileOutlined, NodeCollapseOutlined, PlusCircleFilled, PlusCircleOutlined, ReloadOutlined, SettingFilled, TabletOutlined, WarningFilled, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, InputLabel, LinearProgress, ListItemIcon, ListItemText, MenuItem, Stack, TextField, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik, FormikProvider, useFormik } from "formik";
import { LoadingButton, TreeItem, TreeView } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { PascalToSnake, crontabToText, isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import { useTheme } from '@mui/material/styles';
import MiniPlotComponent from '../dashboard/components/mini-plot';
import LogLine from "../../components/logLine";
import { CosmosContainerPicker } from '../config/users/containerPicker';
import * as yup from "yup";
import { useTranslation } from 'react-i18next';

const NewJobDialog = ({job, OnClose, refresh}) => {
  const { t } = useTranslation();
  const isEdit = job && typeof job === 'object';
  const [config, setConfig] = useState(null);

  const formik = useFormik({
    initialValues: {
      Enabled: isEdit ? job.Enabled : true,
      Name: isEdit ? job.Name : '',
      Crontab: isEdit ? job.Crontab : '',
      Command: isEdit ? job.Command : '',
      Container: isEdit ? job.Container : '',
    },
    validateOnChange: false,
    validationSchema: yup.object({
    }),
    onSubmit: async (values, { setErrors, setStatus, setSubmitting }) => {
      setSubmitting(true);
      let configFile = (await API.config.get()).data;
      if(!configFile.CRON) {
        configFile.CRON = {};
      }

      if(isEdit) {
        delete configFile.CRON[job.Name];
      }

      configFile.CRON[values.Name] = values;

      return API.config.set(configFile).then((res) => {
        setStatus({ success: true });
        setSubmitting(false);
        refresh && refresh();
        OnClose && OnClose();
      }).catch((err) => {
        setStatus({ success: false });
        setErrors({ submit: err.message });
        setSubmitting(false);
        refresh && refresh();
      });
    },
  });
  
  useEffect(() => {
    if (isEdit && job) {
      API.config.get().then((res) => {
        setConfig(res.data);
        let configJob = res.data.CRON[job.Name];
        if (!configJob) {
          return;
        }
        formik.setValues(configJob);
      });
    }
  }, [job]);
  

  return <Dialog open={job} onClose={() => {
      OnClose && OnClose();
    }}>
      <FormikProvider value={formik}>
        <form onSubmit={formik.handleSubmit}>
          <DialogTitle>{t('mgmt.cron.editCronTitle')}</DialogTitle>
          <DialogContent>
              <DialogContentText>
                <Stack spacing={2} style={{ marginTop: '10px', width: '500px', maxWidth: '100%' }}>
                  <div>
                    {t('mgmt.cron.editCron.customText')} (<strong>{t('mgmt.cron.editCron.customText.onHostOnly')}</strong>).
                  </div>
                  <TextField
                    fullWidth
                    id="Name"
                    name="Name"
                    label={t('mgmt.cron.newCron.cronNameInput.cronNameLabel')}
                    value={formik.values.Name}
                    onChange={formik.handleChange}
                    error={formik.touched.Name && Boolean(formik.errors.Name)}
                    helperText={formik.touched.Name && formik.errors.Name}
                  />
                  
                  <TextField
                    fullWidth
                    id="Crontab"
                    name="Crontab"
                    label={t('mgmt.cron.newCron.crontabInput.crontabLabel')}
                    value={formik.values.Crontab}
                    onChange={formik.handleChange}
                    error={formik.touched.Crontab && Boolean(formik.errors.Crontab)}
                    helperText={formik.touched.Crontab && formik.errors.Crontab}
                  />
                  <InputLabel>{crontabToText(formik.values.Crontab, t)}</InputLabel>

                  <TextField
                    fullWidth
                    id="Command"
                    name="Command"
                    label={t('mgmt.cron.newCron.commandInput.commandLabel')}
                    value={formik.values.Command}
                    onChange={formik.handleChange}
                    error={formik.touched.Command && Boolean(formik.errors.Command)}
                    helperText={formik.touched.Command && formik.errors.Command}
                  />

                  <CosmosContainerPicker
                    formik={formik}
                    onTargetChange={(_, name) => {
                      formik.setFieldValue('Container', name);
                    }}
                    name='Container'
                    nameOnly
                  />
                   
                  {formik.errors.submit && <Alert severity="error">{formik.errors.submit}</Alert>}
                </Stack>
              </DialogContentText>
          </DialogContent>
          <DialogActions>
              <Button onClick={() => {
                  OnClose && OnClose();
              }}>{t('global.close')}</Button>
              <LoadingButton color="primary" variant="contained" type="submit" onClick={() => {
                formik.handleSubmit();
              }}>{t('mgmt.cron.newCron.submitButton')}</LoadingButton>
          </DialogActions>
        </form>
      </FormikProvider>
  </Dialog>
};

export default NewJobDialog;