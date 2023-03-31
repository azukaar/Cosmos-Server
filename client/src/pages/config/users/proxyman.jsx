import * as React from 'react';
import isLoggedIn from '../../../isLoggedIn';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik, Field } from 'formik';
import * as Yup from 'yup';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined } from '@ant-design/icons';
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

} from '@mui/material';
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import AnimateButton from '../../../components/@extended/AnimateButton';
import RestartModal from './restart';
import RouteManagement from './routeman';


const ProxyManagement = () => {
  isLoggedIn();
  const [config, setConfig] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);
  const [error, setError] = React.useState(null);

  function updateRoutes(routes) {
    let con = {
      ...config,
      HTTPConfig: {
        ...config.HTTPConfig,
        ProxyConfig: {
          ...config.HTTPConfig.ProxyConfig,
          Routes: routes,
        },
      },
    };
    setConfig(con);
    return con;
  }

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  function up(key) {
    if (key > 0) {
      let tmp = routes[key];
      routes[key] = routes[key-1];
      routes[key-1] = tmp;
      updateRoutes(routes);
    }
  }

  function deleteRoute(key) {
    routes.splice(key, 1);
    updateRoutes(routes);
  }

  function down(key) {
    if (key < routes.length - 1) {
      let tmp = routes[key];
      routes[key] = routes[key+1];
      routes[key+1] = tmp;
      updateRoutes(routes);
    }
  }

  React.useEffect(() => {
    refresh();
  }, []);

  let routes = config && (config.HTTPConfig.ProxyConfig.Routes || []);

  return <div style={{ maxWidth: '1000px', margin: '' }}>
    <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
        refresh();
    }}>Refresh</Button>&nbsp;&nbsp;
    <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
        routes.push({
          Name: 'New Route',
          Description: 'New Route',
          Mode: "SERVAPP",
          UseHost: false,
          Host: '',
          UsePathPrefix: false,
          PathPrefix: '',
          Timeout: 30000,
          ThrottlePerMinute: 100,
          CORSOrigin: '',
          StripPathPrefix: false,
          AuthEnabled: false,
        });
        updateRoutes(routes);
    }}>Create</Button>
    
    <br /><br />
    
    {config && <>
      <RestartModal openModal={openModal} setOpenModal={setOpenModal} />
      {routes && routes.map((route,key) => (<>
        <RouteManagement routeConfig={route}
          setRouteConfig={(newRoute) => {
            routes[key] = newRoute;
          }}
          up={() => up(key)}
          down={() => down(key)}
          deleteRoute={() => deleteRoute(key)}
        />
        <br /><br />
      </>))}

      {routes && 
        <MainCard>
          {error && (
            <Grid item xs={12}>
              <FormHelperText error>{error}</FormHelperText>
            </Grid>
          )}
          <Grid item xs={12}>
            <AnimateButton>
              <Button
                disableElevation
                disabled={false}
                fullWidth
                onClick={() => {
                  API.config.set(updateRoutes(routes)).then(() => {
                    setOpenModal(true);
                  }).catch((err) => {
                    console.log(err);
                    setError(err.message);
                  });
                }}
                size="large"
                type="submit"
                variant="contained"
                color="primary"
              >
                Save
              </Button>
            </AnimateButton>
          </Grid>
        </MainCard>
      }
      {!routes && <>
        <Typography variant="h6" gutterBottom component="div">
          No routes configured.
        </Typography>
      </>
      }
    </>}
  </div>;
}

export default ProxyManagement;