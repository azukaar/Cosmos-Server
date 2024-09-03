import * as React from 'react';
import * as API from '../../api';
import MainCard from '../../components/MainCard';
import { Formik, Field, useFormik, FormikProvider } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Checkbox,
  FormControlLabel,
  Grid,
  InputLabel,
  OutlinedInput,
  Stack,
  FormHelperText,
  TextField,
  MenuItem,
  Skeleton,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  Box,
} from '@mui/material';
import { ExclamationCircleOutlined, InfoCircleOutlined, PlusCircleOutlined, SyncOutlined, WarningOutlined } from '@ant-design/icons';
import PrettyTableView from '../../components/tableView/prettyTableView';
import { DeleteButton } from '../../components/delete';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../config/users/formShortcuts';
import { MetricPicker } from './MetricsPicker';
import { useTranslation } from 'react-i18next';

const DisplayOperator = (operator) => {
  switch (operator) {
    case 'gt':
      return '>';
    case 'lt':
      return '<';
    case 'eq':
      return '=';
    default:
      return '?';
  }
}
const AlertValidationSchema = Yup.object().shape({
  name: Yup.string().required('Name is required'),
  trackingMetric: Yup.string().required('Tracking metric is required'),
  conditionOperator: Yup.string().required('Condition operator is required'),
  conditionValue: Yup.number().required('Condition value is required'),
  period: Yup.string().required('Period is required'),
});

const EditAlertModal = ({ open, onClose, onSave }) => {
  const { t } = useTranslation();
  const formik = useFormik({
    initialValues: {
      name: open.Name || t('navigation.monitoring.alerts.newAlertButton'),
      trackingMetric: open.TrackingMetric || '',
      conditionOperator: (open.Condition && open.Condition.Operator) || 'gt',
      conditionValue: (open.Condition && open.Condition.Value) || 0,
      conditionPercent: (open.Condition && open.Condition.Percent) || false,
      period: open.Period || 'latest',
      actions: open.Actions || [],
      throttled: typeof open.Throttled === 'boolean' ? open.Throttled : true,
      severity: open.Severity || 'error',
    },
    validationSchema: AlertValidationSchema,
    onSubmit: (values) => {
      values.actions = values.actions.filter((a) => !a.removed);
      onSave(values);
      onClose();
    },
  });

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('navigation.monitoring.alerts.action.edit')}</DialogTitle>
      <FormikProvider value={formik}>
      <form onSubmit={formik.handleSubmit}>
        <DialogContent>
          <Stack spacing={2}>
            <CosmosInputText
              name="name"
              label={t('navigation.monitoring.alerts.alertNameLabel')}
              formik={formik}
              required
            />
            <MetricPicker
              name="trackingMetric"
              label={t('navigation.monitoring.alerts.trackingMetricLabel')}
              formik={formik}
              required
            />
            <Stack direction="row" spacing={2} alignItems="center">
              <CosmosSelect
                name="conditionOperator"
                label={t('navigation.monitoring.alerts.conditionOperatorLabel')}
                formik={formik}
                options={[
                  ['gt', '>'],
                  ['lt', '<'],
                  ['eq', '='],
                ]}
              >
              </CosmosSelect>
              <CosmosInputText
                name="conditionValue"
                label={t('navigation.monitoring.alerts.conditionValueLabel')}
                formik={formik}
                required
              />
              <CosmosCheckbox
                style={{paddingTop: '20px'}}
                name="conditionPercent"
                label={t('navigation.monitoring.alerts.conditionLabel')}
                formik={formik}
              />
            </Stack>

            <CosmosSelect
              name="period"
              label={t('navigation.monitoring.alerts.periodLabel')}
              formik={formik}
              options={[
                ['latest', t('navigation.monitoring.latest')],
                ['hourly', t('navigation.monitoring.hourly')],
                ['daily', t('navigation.monitoring.daily')],
            ]}></CosmosSelect>

            <CosmosSelect
              name="severity"
              label={t('navigation.monitoring.alerts.action.edit.severitySelection.severityLabel')}
              formik={formik}
              options={[
                ['info', 'Info'],
                ['warn', 'Warning'],
                ['error', 'Error'],
            ]}></CosmosSelect>

            <CosmosCheckbox
              name="throttled"
              label={t('navigation.monitoring.alerts.throttleCheckbox.throttleLabel')}
              formik={formik}
            />

            <CosmosFormDivider title={t('mgmt.monitoring.alerts.actionTriggersTitle')} />
            
            <Stack direction="column" spacing={2}>
              {formik.values.actions
              .map((action, index) => {
                return !action.removed && <>
                  {action.Type === 'stop' && 
                    <Alert severity="info">{t('navigation.monitoring.alerts.actions.stopActionInfo')}</Alert>
                  }
                  {action.Type === 'restart' &&
                    <Alert severity="info">{t('navigation.monitoring.alerts.actions.restartActionInfo')}</Alert>
                  }
                  <Stack direction="row" spacing={2} key={index}>
                    <Box style={{
                      width: '100%',
                    }}>
                      <CosmosSelect
                        name={`actions.${index}.Type`}
                        label={t('navigation.monitoring.alerts.action.edit.actionTypeInput.actionTypeLabel')}
                        formik={formik}
                        options={[
                          ['notification', t('navigation.monitoring.alerts.actions.sendNotification')],
                          ['email', t('navigation.monitoring.alerts.actions.sendEmail')],
                          ['stop', t('navigation.monitoring.alerts.actions.stop')],
                          ['restart', t('navigation.monitoring.alerts.actions.restart')],
                        ]}
                      />
                    </Box>
                    
                    <Box  style={{
                      height: '95px',
                      display: 'flex',
                      alignItems: 'center',
                    }}>
                      <DeleteButton
                        onDelete={() => {
                          formik.setFieldValue(`actions.${index}.removed`, true);
                        }}
                      />
                    </Box>
                  </Stack>
                </>
              })}

              <Button
                variant="outlined"
                color="primary"
                startIcon={<PlusCircleOutlined />}
                onClick={() => {
                  formik.setFieldValue('actions', [
                    ...formik.values.actions,
                    {
                      Type: 'notification',
                    },
                  ]);
                }}>
                {t('mgmt.monitoring.alerts.addActionButton')}
              </Button>
            </Stack>

          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose}>{t('global.cancelAction')}</Button>
          <Button variant='contained' type="submit">{t('global.saveAction')}</Button>
        </DialogActions>
      </form>
      </FormikProvider>
    </Dialog>
  );
};

const AlertPage = () => {
  const { t } = useTranslation();
  const [config, setConfig] = React.useState(null);
  const [isLoading, setIsLoading] = React.useState(false);
  const [openModal, setOpenModal] = React.useState(false);
  const [metrics, setMetrics] = React.useState({});

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
      setIsLoading(false);
    });
    API.metrics.list().then((res) => {
      setMetrics(res.data);
    });
  }

  React.useEffect(() => {
    refresh();
  }, []);

  const setEnabled = (name) => (event) => {
    setIsLoading(true);
    let toSave = {
      ...config,
      MonitoringAlerts: {
        ...config.MonitoringAlerts,
        [name]: {
          ...config.MonitoringAlerts[name],
          Enabled: event.target.checked,
        }
      }
    };

    API.config.set(toSave).then(() => {
      refresh();
    });
  }

  const deleteAlert = (name) => {
    setIsLoading(true);
    let toSave = {
      ...config,
      MonitoringAlerts: {
        ...config.MonitoringAlerts,
      }
    };
    delete toSave.MonitoringAlerts[name];

    API.config.set(toSave).then(() => {
      refresh();
    });
  }
  
  const saveAlert = (data) => {
    setIsLoading(true);
    
    data.conditionValue = parseInt(data.conditionValue);

    let toSave = {
      ...config,
      MonitoringAlerts: {
        ...config.MonitoringAlerts,
        [data.name]: {
          Name: data.name,
          Enabled: true,
          TrackingMetric: data.trackingMetric,
          Condition: {
            Operator: data.conditionOperator,
            Value: data.conditionValue,
            Percent: data.conditionPercent,
          },
          Period: data.period,
          Actions: data.actions,
          LastTriggered: null,
          Throttled: data.throttled,
          Severity: data.severity,
        }
      }
    };

    API.config.set(toSave).then(() => {
      refresh();
    });
  }
  
  const resetTodefault = () => {
    setIsLoading(true);

    let toSave = {
      ...config,
      MonitoringAlerts: {
        "Anti Crypto-Miner": {
          "Name": "Anti Crypto-Miner",
          "Enabled": false,
          "Period": "daily",
          "TrackingMetric": "cosmos.system.docker.cpu.*",
          "Condition": {
            "Operator": "gt",
            "Value": 80
          },
          "Actions": [
            {
              "mgmt.storage.typeTitle": "notification",
              "Target": ""
            },
            {
              "mgmt.storage.typeTitle": "email",
              "Target": ""
            },
            {
              "mgmt.storage.typeTitle": "stop",
              "Target": ""
            }
          ],
          "LastTriggered": "0001-01-01T00:00:00Z",
          "Throttled": false,
          "Severity": "warn"
        },
        "Anti Memory Leak": {
          "Name": "Anti Memory Leak",
          "Enabled": false,
          "Period": "daily",
          "TrackingMetric": "cosmos.system.docker.ram.*",
          "Condition": {
            "Percent": true,
            "Operator": "gt",
            "Value": 80
          },
          "Actions": [
            {
              "mgmt.storage.typeTitle": "notification",
              "Target": ""
            },
            {
              "mgmt.storage.typeTitle": "email",
              "Target": ""
            },
            {
              "mgmt.storage.typeTitle": "stop",
              "Target": ""
            }
          ],
          "LastTriggered": "0001-01-01T00:00:00Z",
          "Throttled": false,
          "Severity": "warn"
        },
        "Disk Full Notification": {
          "Name": 'Disk Full Notification',
          "Enabled": true,
          "Period": "latest",
          "TrackingMetric": "cosmos.system.disk./",
          "Condition": {
            "Percent": true,
            "Operator": "gt",
            "Value": 95
          },
          "Actions": [
            {
              "mgmt.storage.typeTitle": "notification",
              "Target": ""
            }
          ],
          "LastTriggered": "0001-01-01T00:00:00Z",
          "Throttled": true,
          "Severity": "warn"
        }
      }
    };

    API.config.set(toSave).then(() => {
      refresh();
    });
  }

  const GetSevIcon = ({level}) => {
    switch (level) {
      case 'info':
        return <span style={{color: '#2196f3'}}><InfoCircleOutlined /></span>;
      case 'warn':
        return <span style={{color: '#ff9800'}}><WarningOutlined /></span>;
      case 'error':
        return <span style={{color: '#f44336'}}><ExclamationCircleOutlined /></span>;
      default:
        return '';
    }
  }

  return <div style={{maxWidth: '1200px', margin: ''}}>
    {openModal && <EditAlertModal open={openModal} onClose={() => setOpenModal(false)} onSave={saveAlert} />}

    <Stack direction="row" spacing={2} style={{marginBottom: '15px'}}>
      <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          refresh();
      }}>{t('global.refresh')}</Button>
      <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
          setOpenModal(true);
      }}>{t('global.createAction')}</Button>
      <Button variant="outlined" color="warning" startIcon={<WarningOutlined />} onClick={() => {
        resetTodefault();
      }}>{t('navigation.monitoring.alerts.resetToDefaultButton')}</Button>
    </Stack>
    
    {config && <>
      <Formik
        initialValues={{
          Actions: config.MonitoringAlerts
        }}

        // validationSchema={Yup.object().shape({
        // })}

        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          setSubmitting(true);
        
          let toSave = {
            ...config,
            MonitoringAlerts: values.Actions
          };

          return API.config.set(toSave);
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={3}>
            {!config && <Skeleton variant="rectangular" height={300} />}
            {config && (!config.MonitoringAlerts || !Object.values(config.MonitoringAlerts).length) ? <Alert severity="info">No alerts configured.</Alert> : ''}
            {config && config.MonitoringAlerts && Object.values(config.MonitoringAlerts).length ? <PrettyTableView 
              data={Object.values(config.MonitoringAlerts)}
              getKey={(r) => r.Name + r.Target + r.Mode}
              onRowClick={(r, k) => {
                setOpenModal(r);
              }}

              columns={[
                { 
                  title: t('global.enabled'), 
                  clickable:true, 
                  field: (r, k) => <Checkbox disabled={isLoading} size='large' color={r.Enabled ? 'success' : 'default'}
                    onChange={setEnabled(Object.keys(config.MonitoringAlerts)[k])}
                    checked={r.Enabled}
                  />,
                  style: {
                  },
                },
                { 
                  title: t('global.nameTitle'), 
                  field: (r) => <><GetSevIcon level={r.Severity} /> {r.Name}</>,
                },
                { 
                  title: t('navigation.monitoring.alerts.trackingMetricTitle'), 
                  field: (r) => metrics[r.TrackingMetric] ? metrics[r.TrackingMetric] : r.TrackingMetric,
                },
                { 
                  title: t('navigation.monitoring.alerts.conditionTitle'), 
                  screenMin: 'md',
                  field: (r) => DisplayOperator(r.Condition.Operator) + ' ' + r.Condition.Value + (r.Condition.Percent ? '%' : ''),
                },
                { 
                  title: t('navigation.monitoring.alerts.periodTitle'),
                  field: (r) => t(r.Period),
                },
                { 
                  title: t('navigation.monitoring.alerts.astTriggeredTitle'),
                  screenMin: 'md',
                  field: (r) => (r.LastTriggered != "0001-01-01T00:00:00Z") ? new Date(r.LastTriggered).toLocaleString() : t('global.never'),
                },
                { 
                  title: t('navigation.monitoring.alerts.actionsTitle'), 
                  field: (r) => r.Actions.map((a) => a.Type).join(', '),
                  screenMin: 'md',
                },
                { title: '', clickable:true, field: (r, k) =>  <DeleteButton disabled={isLoading} onDelete={() => {
                  deleteAlert(Object.keys(config.MonitoringAlerts)[k])
                }}/>,
                  style: {
                    textAlign: 'right',
                  }
                },
              ]}
            /> : ''}
            </Stack>
          </form>
        )}
      </Formik>
    </>}

  </div>;
}

export default AlertPage;