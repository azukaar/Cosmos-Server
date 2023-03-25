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

const RouteManagement = ({ routeConfig, setRouteConfig, up, down, deleteRoute }) => {
  const [confirmDelete, setConfirmDelete] = React.useState(false);

  return <div style={{ maxWidth: '1000px', margin: '' }}>
    {routeConfig && <>
      <Formik
        initialValues={{
          Name: routeConfig.Name,
          Description: routeConfig.Description,
          Mode: routeConfig.Mode,
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
        validationSchema={Yup.object().shape({

        })}
        validate={(values) => {
          setRouteConfig(values);
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          return false;  
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <MainCard title={
              <div>{routeConfig.Name} &nbsp;
                <Chip label={<UpOutlined />} onClick={() => up()}/> &nbsp;
                <Chip label={<DownOutlined />} onClick={() => down()}/> &nbsp;
                {!confirmDelete && (<Chip label={<DeleteOutlined />} onClick={() => setConfirmDelete(true)}/>)}
                {confirmDelete && (<Chip label={<CheckOutlined />} onClick={() => deleteRoute()}/>)} &nbsp;
              </div>
            }>
              <Grid container spacing={2}>

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

                    <CosmosSelect
                      name="Mode"
                      label="Mode"
                      formik={formik}
                      options={[
                        ["SERVAPP", "ServApp - Docker Container"],
                        ["PROXY", "Proxy"],
                        ["STATIC", "Static Folder"],
                        ["SPA", "Single Page Application"],
                      ]}
                    />

                    <CosmosFormDivider title={'Target Settings'}/>

                    {
                      formik.values.Mode === "SERVAPP" ? 
                      <CosmosContainerPicker
                        formik={formik}
                      />
                      :  <CosmosInputText
                      name="Target"
                      label={formik.values.Mode == "PROXY" ? "Target URL" : "Target Folder Path"}
                      placeholder={formik.values.Mode == "PROXY" ? "localhost:8080" : "/path/to/my/app"}
                      formik={formik}
                    />
                    }
                    
                    <CosmosFormDivider title={'Source'}/>

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
                    />}

                    {formik.values.UsePathPrefix && <CosmosCheckbox
                      name="StripPathPrefix"
                      label="Strip Path Prefix"
                      formik={formik}
                    />}

                    <CosmosFormDivider title={'Security'}/>
                    
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