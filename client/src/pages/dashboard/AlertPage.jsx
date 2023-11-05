import * as React from 'react';
import IsLoggedIn from '../../isLoggedIn';
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
  const formik = useFormik({
    initialValues: {
      name: open.Name || 'New Alert',
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
      <DialogTitle>Edit Alert</DialogTitle>
      <FormikProvider value={formik}>
      <form onSubmit={formik.handleSubmit}>
        <DialogContent>
          <Stack spacing={2}>
            <CosmosInputText
              name="name"
              label="Name of the alert"
              formik={formik}
              required
            />
            <MetricPicker
              name="trackingMetric"
              label="Metric to track"
              formik={formik}
              required
            />
            <Stack direction="row" spacing={2} alignItems="center">
              <CosmosSelect
                name="conditionOperator"
                label="Trigger Condition Operator"
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
                label="Trigger Condition Value"
                formik={formik}
                required
              />
              <CosmosCheckbox
                style={{paddingTop: '20px'}}
                name="conditionPercent"
                label="Condition is a percent of max value"
                formik={formik}
              />
            </Stack>

            <CosmosSelect
              name="period"
              label="Period (how often to check the metric)"
              formik={formik}
              options={[
                ['latest', 'Latest'],
                ['hourly', 'Hourly'],
                ['daily', 'Daily'],
            ]}></CosmosSelect>

            <CosmosSelect
              name="severity"
              label="Severity"
              formik={formik}
              options={[
                ['info', 'Info'],
                ['warn', 'Warning'],
                ['error', 'Error'],
            ]}></CosmosSelect>

            <CosmosCheckbox
              name="throttled"
              label="Throttle (only triggers a maximum of once a day)"
              formik={formik}
            />

            <CosmosFormDivider title={'Action Triggers'} />
            
            <Stack direction="column" spacing={2}>
              {formik.values.actions
              .map((action, index) => {
                return !action.removed && <>
                  {action.Type === 'stop' && 
                    <Alert severity="info">Stop action will attempt to stop/disable any resources (ex. Containers, routes, etc... ) attachted to the metric.
                    This will only have an effect on metrics specific to a resources (ex. CPU of a specific container). It will not do anything on global metric such as global used CPU</Alert>
                  }
                  <Stack direction="row" spacing={2} key={index}>
                    <Box style={{
                      width: '100%',
                    }}>
                      <CosmosSelect
                        name={`actions.${index}.Type`}
                        label="Action Type"
                        formik={formik}
                        options={[
                          ['notification', 'Send a notification'],
                          ['email', 'Send an email'],
                          ['stop', 'Stop resources causing the alert'],
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
                Add Action
              </Button>
            </Stack>

          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose}>Cancel</Button>
          <Button variant='contained' type="submit">Save</Button>
        </DialogActions>
      </form>
      </FormikProvider>
    </Dialog>
  );
};

const AlertPage = () => {
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
              "Type": "notification",
              "Target": ""
            },
            {
              "Type": "email",
              "Target": ""
            },
            {
              "Type": "stop",
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
              "Type": "notification",
              "Target": ""
            },
            {
              "Type": "email",
              "Target": ""
            },
            {
              "Type": "stop",
              "Target": ""
            }
          ],
          "LastTriggered": "0001-01-01T00:00:00Z",
          "Throttled": false,
          "Severity": "warn"
        },
        "Disk Full Notification": {
          "Name": "Disk Full Notification",
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
              "Type": "notification",
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
    <IsLoggedIn />

    {openModal && <EditAlertModal open={openModal} onClose={() => setOpenModal(false)} onSave={saveAlert} />}

    <Stack direction="row" spacing={2} style={{marginBottom: '15px'}}>
      <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          refresh();
      }}>Refresh</Button>
      <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
          setOpenModal(true);
      }}>Create</Button>
      <Button variant="outlined" color="warning" startIcon={<WarningOutlined />} onClick={() => {
        resetTodefault();
      }}>Reset to default</Button>
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
                  title: 'Enabled', 
                  clickable:true, 
                  field: (r, k) => <Checkbox disabled={isLoading} size='large' color={r.Enabled ? 'success' : 'default'}
                    onChange={setEnabled(Object.keys(config.MonitoringAlerts)[k])}
                    checked={r.Enabled}
                  />,
                  style: {
                  },
                },
                { 
                  title: 'Name', 
                  field: (r) => <><GetSevIcon level={r.Severity} /> {r.Name}</>,
                },
                { 
                  title: 'Tracking Metric', 
                  field: (r) => metrics[r.TrackingMetric] ? metrics[r.TrackingMetric] : r.TrackingMetric,
                },
                { 
                  title: 'Condition', 
                  screenMin: 'md',
                  field: (r) => DisplayOperator(r.Condition.Operator) + ' ' + r.Condition.Value + (r.Condition.Percent ? '%' : ''),
                },
                { 
                  title: 'Period',
                  field: (r) => r.Period,
                },
                { 
                  title: 'Last Triggered',
                  screenMin: 'md',
                  field: (r) => (r.LastTriggered != "0001-01-01T00:00:00Z") ? new Date(r.LastTriggered).toLocaleString() : 'Never',
                },
                { 
                  title: 'Actions', 
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