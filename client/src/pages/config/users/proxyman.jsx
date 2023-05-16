import * as React from 'react';
import IsLoggedIn from '../../../isLoggedIn';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik, Field } from 'formik';
import * as Yup from 'yup';
import { useTheme } from '@mui/material/styles';
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
  Chip,
  CircularProgress,

} from '@mui/material';
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import AnimateButton from '../../../components/@extended/AnimateButton';
import RestartModal from './restart';
import RouteManagement from '../routes/routeman';
import { map } from 'lodash';
import { getFaviconURL, sanitizeRoute, ValidateRoute } from '../../../utils/routes';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import HostChip from '../../../components/hostChip';
import {RouteActions, RouteMode, RouteSecurity} from '../../../components/routeComponents';
import { useNavigate } from 'react-router';
import NewRouteCreate from '../routes/newRoute';

const stickyButton = {
  position: 'fixed',
  bottom: '20px',
  width: '300px',
  boxShadow: '0px 0px 10px 0px rgba(0,0,0,0.50)',
  right: '20px',
}

function shorten(test) {
  if (test.length > 75) {
    return test.substring(0, 75) + '...';
  }
  return test;
}

const ProxyManagement = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [config, setConfig] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);
  const [error, setError] = React.useState(null);
  const [submitErrors, setSubmitErrors] = React.useState([]);
  const [needSave, setNeedSave] = React.useState(false);
  const [openNewModal, setOpenNewModal] = React.useState(false);
  
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

  function cleanRoutes(config) {
    config.HTTPConfig.ProxyConfig.Routes = config.HTTPConfig.ProxyConfig.Routes.map((r) => {
      delete r._hasErrors;
      return r;
    });
    return config;
  }    

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  function up(event, key) {
    event.stopPropagation();
    if (key > 0) {
      let tmp = routes[key];
      routes[key] = routes[key-1];
      routes[key-1] = tmp;
      updateRoutes(routes);
      setNeedSave(true);
    }
    return false;
  }

  function deleteRoute(event, key) {
    event.stopPropagation();
    routes.splice(key, 1);
    updateRoutes(routes);
    setNeedSave(true);
    return false;
  }

  function down(event, key) {
    event.stopPropagation();
    if (key < routes.length - 1) {
      let tmp = routes[key];
      routes[key] = routes[key+1];
      routes[key+1] = tmp;
      updateRoutes(routes);
      setNeedSave(true);
    }
    return false;
  }

  React.useEffect(() => {
    refresh();
  }, []);

  let routes = config && (config.HTTPConfig.ProxyConfig.Routes || []);

  return <div style={{   }}>
    <IsLoggedIn />
    <Stack direction="row" spacing={1} style={{ marginBottom: '20px' }}>
      <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          refresh();
      }}>Refresh</Button>&nbsp;&nbsp;
      <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
        setOpenNewModal(true);
      }}>Create</Button>
    </Stack>

    {config && <>
      <RestartModal openModal={openModal} setOpenModal={setOpenModal} />
      <NewRouteCreate openNewModal={openNewModal} setOpenNewModal={setOpenNewModal} config={config}/>
      
      {routes && <PrettyTableView 
        data={routes}
        getKey={(r) => r.Name + r.Target + r.Mode}
        linkTo={(r) => '/ui/config-url/' + r.Name}
        columns={[
          { 
            title: '', 
            field: (r) => <img className="loading-image" alt="" src={getFaviconURL(r)} width="64px" />,
            style: {
              textAlign: 'center',
            },
          },
          { title: 'URL',
            search: (r) => r.Name + ' ' + r.Description,
            style: {
              textDecoration: 'inherit',
            },
            underline: true,
            field: (r) => <>
              <div style={{display:'inline-block', textDecoration: 'inherit', fontSize:'125%', color: isDark ? theme.palette.primary.light : theme.palette.primary.dark}}>{r.Name}</div><br/>
              <div style={{display:'inline-block', textDecoration: 'inherit', fontSize: '90%', opacity: '90%'}}>{r.Description}</div>
            </>
          },
          { title: 'Origin', screenMin: 'md', clickable:true, search: (r) => r.Host + ' ' + r.PathPrefix, field: (r) => <HostChip route={r} /> },
          { title: 'Target', screenMin: 'md', search: (r) => r.Target, field: (r) => <><RouteMode route={r} /> <Chip label={r.Target} /></> },
          { title: 'Security', screenMin: 'lg', field: (r) => <RouteSecurity route={r} />,
          style: {minWidth: '70px'} },
          { title: '', clickable:true, field: (r, k) =>  <RouteActions
              route={r}
              routeKey={k}
              up={(event) => up(event, k)}
              down={(event) => down(event, k)}
              deleteRoute={(event) => deleteRoute(event, k)}
            />,
            style: {
              textAlign: 'right',
            }
          },
        ]}
      />}
      {
        !routes && <div style={{textAlign: 'center'}}>
          <CircularProgress />
        </div>
      }
      
      {routes && needSave && <>
        <div>
        <br /><br /><br /><br />
        </div>
        <Stack style={{position: 'relative'}} fullWidth spacing={1}>
        <div style={stickyButton}>
        <MainCard>
          {error && (
            <Grid item xs={12}>
              <FormHelperText error>{error}</FormHelperText>
            </Grid>
          )}
            <Stack spacing={1}>
            {submitErrors.map((err) => {
              return <Alert severity="error">{err}</Alert>
            })}
            <AnimateButton>
              <Button
                className='shinyButton'
                disableElevation
                fullWidth
                onClick={() => {
                  if(routes.some((route, key) => {
                    let errors = ValidateRoute(route, config);
                    if (errors && errors.length > 0) {
                      errors = errors.map((err) => {
                        return `${route.Name}: ${err}`;
                      });
                      setSubmitErrors(errors);
                      return true;
                    }
                  })) {
                    return;
                  } else {
                    setSubmitErrors([]);
                  }
                  API.config.set(cleanRoutes(updateRoutes(routes))).then(() => {
                    setNeedSave(false);
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
                Save Changes
              </Button>
            </AnimateButton>
            </Stack>
        </MainCard>
        </div>
        </Stack>
        </>
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