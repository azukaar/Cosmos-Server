import * as React from 'react';
import isLoggedIn from '../../../isLoggedIn';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik, Field } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Checkbox,
  Divider,
  FormControlLabel,
  Grid,
  IconButton,
  InputAdornment,
  InputLabel,
  Link,
  OutlinedInput,
  Stack,
  Typography,
  FormHelperText,
  Collapse,
  TextField,
  MenuItem,
  Card,
  Chip,

} from '@mui/material';
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import AnimateButton from '../../../components/@extended/AnimateButton';
import RestartModal from './restart';
import { CosmosCheckbox, CosmosCollapse, CosmosFormDivider, CosmosInputText, CosmosSelect } from './formShortcuts';
import { DownOutlined, UpOutlined, CheckOutlined, DeleteOutlined  } from '@ant-design/icons';
import { CosmosContainerPicker } from './containerPicker';

export const ValidateRoute = Yup.object().shape({
  Name: Yup.string().required('Name is required'),
  Mode: Yup.string().required('Mode is required'),
  Target: Yup.string().required('Target is required').when('Mode', {
    is: 'SERVAPP',
    then: Yup.string().matches(/:[0-9]+$/, 'Invalid Target, must have a port'),
  }),

  Host: Yup.string().when('UseHost', {
    is: true,
    then: Yup.string().required('Host is required')
      .matches(/[\.|\:]/, 'Host must be full domain ([sub.]domain.com) or an IP')
  }),

  PathPrefix: Yup.string().when('UsePathPrefix', {
    is: true,
    then: Yup.string().required('Path Prefix is required').matches(/^\//, 'Path Prefix must start with / (e.g. /api). Do not include a domain/subdomain in it, use the Host for this.')
  }),
  
  UseHost: Yup.boolean().when('UsePathPrefix', 
  {
    is: false,
    then: Yup.boolean().oneOf([true], 'Source must at least be either Host or Path Prefix')
  }),
})

const RouteManagement = ({ routeConfig, TargetContainer, noControls=false, lockTarget=false, setRouteConfig, up, down, deleteRoute }) => {
  const [confirmDelete, setConfirmDelete] = React.useState(false);
  
  return <div style={{ maxWidth: '1000px', margin: '' }}>
    {routeConfig && <>
      <Formik
        initialValues={{
          Name: routeConfig.Name,
          Description: routeConfig.Description,
          Mode: routeConfig.Mode || "SERVAPP",
          Target: routeConfig.Target,
          UseHost: routeConfig.UseHost,
          AuthEnabled: routeConfig.AuthEnabled,
          Host: routeConfig.Host,
          UsePathPrefix: routeConfig.UsePathPrefix,
          PathPrefix: routeConfig.PathPrefix,
          StripPathPrefix: routeConfig.StripPathPrefix,
          Timeout: routeConfig.Timeout,
          ThrottlePerMinute: routeConfig.ThrottlePerMinute,
          CORSOrigin: routeConfig.CORSOrigin,
        }}
        validateOnChange={false}
        validationSchema={ValidateRoute}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          return false;  
        }}
        validate={(values) => {
            //setRouteConfig(values);
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <MainCard title={
              noControls ? 'New Route' :
              <div>{routeConfig.Name} &nbsp;
                <Chip label={<UpOutlined />} onClick={() => up()}/> &nbsp;
                <Chip label={<DownOutlined />} onClick={() => down()}/> &nbsp;
                {!confirmDelete && (<Chip label={<DeleteOutlined />} onClick={() => setConfirmDelete(true)}/>)}
                {confirmDelete && (<Chip label={<CheckOutlined />} onClick={() => deleteRoute()}/>)} &nbsp;
              </div>
            }>
              <Grid container spacing={2}>
                {formik.errors.submit && (
                  <Grid item xs={12}>
                    <FormHelperText error>{formik.errors.submit}</FormHelperText>
                  </Grid>
                )}

                <CosmosInputText
                  name="Name"
                  label="Name"
                  placeholder="Name"
                  formik={formik}
                />

                <CosmosInputText
                  name="Description"
                  label="Description"
                  placeholder="Description"
                  formik={formik}
                />

                <CosmosCollapse title="Settings">
                  <Grid container spacing={2}>
                    <CosmosFormDivider title={'Target Type'}/>
                    <Grid item xs={12}>
                    <Alert color='info'>What are you trying to access with this route?</Alert>
                    </Grid>

                    <CosmosSelect
                      name="Mode"
                      label="Mode"
                      formik={formik}
                      disabled={lockTarget}
                      options={[
                        ["SERVAPP", "ServApp - Docker Container"],
                        ["PROXY", "Proxy"],
                        ["STATIC", "Static Folder"],
                        ["SPA", "Single Page Application"],
                        ["REDIRECT", "Redirection"]
                      ]}
                    />

                    <CosmosFormDivider title={'Target Settings'}/>

                    {
                      (formik.values.Mode === "SERVAPP")? 
                      <CosmosContainerPicker
                        formik={formik}
                        lockTarget={lockTarget} 
                        TargetContainer={TargetContainer}
                      />
                      :  <CosmosInputText
                        name="Target"
                        label={formik.values.Mode == "PROXY" ? "Target URL" : "Target Folder Path"}
                        placeholder={formik.values.Mode == "PROXY" ? "localhost:8080" : "/path/to/my/app"}
                        formik={formik}
                      />
                    }
                    
                    <CosmosFormDivider title={'Source'}/>
                    <Grid item xs={12}>
                    <Alert color='info'>What URL do you want to access your target from?</Alert>
                    </Grid>
                    <CosmosCheckbox
                      name="UseHost"
                      label="Use Host"
                      formik={formik}
                    />

                    {formik.values.UseHost && <CosmosInputText
                      name="Host"
                      label="Host"
                      placeholder="Host"
                      formik={formik}
                      style={{paddingLeft: '20px'}}
                    />}

                    <CosmosCheckbox
                      name="UsePathPrefix"
                      label="Use Path Prefix"
                      formik={formik}
                    />

                    {formik.values.UsePathPrefix && <CosmosInputText
                      name="PathPrefix"
                      label="Path Prefix"
                      placeholder="Path Prefix"
                      formik={formik}
                      style={{paddingLeft: '20px'}}
                    />}

                    {formik.values.UsePathPrefix && <CosmosCheckbox
                      name="StripPathPrefix"
                      label="Strip Path Prefix"
                      formik={formik}
                      style={{paddingLeft: '20px'}}
                    />}

                    <CosmosFormDivider title={'Security'}/>
                    
                    <Grid item xs={12}>
                    <Alert color='info'>Additional security settings. MFA and Captcha are not yet implemented.</Alert>
                    </Grid>

                    <CosmosCheckbox
                      name="AuthEnabled"
                      label="Authentication Required"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="Timeout"
                      label="Timeout in milliseconds (0 for no timeout, at least 30000 or less recommended)"
                      placeholder="Timeout"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="ThrottlePerMinute"
                      label="Maximum number of requests Per Minute (0 for no limit, at least 100 or less recommended)"
                      placeholder="Throttle Per Minute"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="CORSOrigin"
                      label="Custom CORS Origin (Recommended to leave blank)"
                      placeholder="CORS Origin"
                      formik={formik}
                    />
                  </Grid>
                </CosmosCollapse>
              </Grid>
            </MainCard>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteManagement;