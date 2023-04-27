import * as React from 'react';
import MainCard from '../../../components/MainCard';
import { Formik } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Grid,
  FormHelperText,
  Chip,

} from '@mui/material';
import { CosmosCheckbox, CosmosCollapse, CosmosFormDivider, CosmosInputText, CosmosSelect } from './formShortcuts';
import { DownOutlined, UpOutlined, CheckOutlined, DeleteOutlined  } from '@ant-design/icons';
import { CosmosContainerPicker } from './containerPicker';
import { ValidateRoute } from './users/routeman';

const RouteConfig = ({route, key, lockTarget, TargetContainer, setRouteConfig}) => {
  return (<div style={{ maxWidth: '1000px', margin: '' }}>
    {route && <>
      <Formik
        initialValues={{
          Name: route.Name,
          Description: route.Description,
          Mode: route.Mode || "SERVAPP",
          Target: route.Target,
          UseHost: route.UseHost,
          AuthEnabled: route.AuthEnabled,
          Host: route.Host,
          UsePathPrefix: route.UsePathPrefix,
          PathPrefix: route.PathPrefix,
          StripPathPrefix: route.StripPathPrefix,
          Timeout: route.Timeout,
          ThrottlePerMinute: route.ThrottlePerMinute,
          CORSOrigin: route.CORSOrigin,
        }}
        validationSchema={ValidateRoute}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          return false;  
        }}
        // validate={(values) => {
        //     setRouteConfig(values);
        // }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <MainCard name={route.Name} title={route.Name}>
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
                        onTargetChange={() => {
                          setRouteConfig(formik.values);
                        }}
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
      </>
    }
  </div>)
}

export default RouteConfig;