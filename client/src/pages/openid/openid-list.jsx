import * as React from 'react';
import * as API from '../../api';
import MainCard from '../../components/MainCard';
import { Formik, Field } from 'formik';
import * as Yup from 'yup';
import { useTheme } from '@mui/material/styles';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined, SyncOutlined, UserOutlined, KeyOutlined, ArrowRightOutlined } from '@ant-design/icons';
import {
  Alert,
  Button,
  Checkbox,
  Stack,
  Typography,
  FormHelperText,
  Chip,
  CircularProgress,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  IconButton,

} from '@mui/material';
import { EyeOutlined, EyeInvisibleOutlined } from '@ant-design/icons';
import AnimateButton from '../../components/@extended/AnimateButton';
import RestartModal from '../config/users/restart';
import RouteManagement from '../config/routes/routeman';
import { getFaviconURL, sanitizeRoute, ValidateRoute } from '../../utils/routes';
import PrettyTableView from '../../components/tableView/prettyTableView';
import HostChip from '../../components/hostChip';
import { RouteActions, RouteMode, RouteSecurity } from '../../components/routeComponents';
import { useNavigate } from 'react-router';
import NewRouteCreate from '../config/routes/newRoute';
import { DeleteButton } from '../../components/delete';
import OpenIdEditModal from './openid-edit';
import bcrypt from 'bcryptjs';
import { useTranslation } from 'react-i18next';

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

const OpenIdList = () => {
  const { t } = useTranslation();
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const [config, setConfig] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);
  const [error, setError] = React.useState(null);
  const [submitErrors, setSubmitErrors] = React.useState([]);
  const [needSave, setNeedSave] = React.useState(false);
  const [clientId, setClientId] = React.useState(false);
  const [openNewModal, setOpenNewModal] = React.useState(false);
  const [newSecret, setNewSecret] = React.useState(null);

  function updateRoutes(routes) {
    let con = {
      ...config,
      OpenIDClients: routes,
    };
    setConfig(con);
    return con;
  }

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  function save(con) {
    API.config.set(con).then((res) => {
      refresh();
    });
  }

  function getClientIcon(client) {
    try {
      let hostname = new URL(client.redirect).hostname;
      let port = new URL(client.redirect).port;
      let protocol = new URL(client.redirect).protocol + '//';
      return getFaviconURL({
        Mode: 'PROXY',
        Name: client.id,
        Target: protocol + hostname + (port ? ':' + port : ''),
      })
    } catch (e) {
      return getFaviconURL()
    }
  }

  function deleteClient(event, key) {
    event.stopPropagation();
    clients.splice(key, 1);
    save(updateRoutes(clients));
    return false;
  }

  React.useEffect(() => {
    refresh();
  }, []);

  const generateNewSecret = (clientIdToUpdate) => {
    let newSecretSeed = window.crypto.getRandomValues(new Uint32Array(4));
    let newSecret = "";
    newSecretSeed.forEach((r) => {
      newSecret += r.toString(36);
    });
    let encryptedSecret = bcrypt.hashSync(newSecret, 10);
    let index = clients.findIndex((r) => r.id === clientIdToUpdate);
    clients[index].secret = encryptedSecret;
    save(updateRoutes(clients));
    setNewSecret(newSecret);
  }

  let clients = config && (config.OpenIDClients || []);

  return <div style={{}}>
    <Stack direction="row" spacing={1} style={{ marginBottom: '20px' }}>
      <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
        refresh();
      }}>{t('global.refresh')}</Button>&nbsp;&nbsp;
      <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
        setClientId(null);
        setOpenNewModal(true);
      }}>{t('global.createAction')}</Button>
    </Stack>

    {config && <>
      <OpenIdEditModal
        clientId={clientId}
        openNewModal={openNewModal}
        setOpenNewModal={setOpenNewModal}
        config={config}
        onSubmit={(values) => {
          if (clientId) {
            let index = clients.findIndex((r) => r.id === clientId);
            clients[index] = values;
          } else {
            clients.push(values);
          }

          save(updateRoutes(clients));
          setOpenNewModal(false);
          setClientId(null);
        }}
      />

      {newSecret && <Dialog open={newSecret} onClose={() => setNewSecret(false)}>
        <DialogTitle>{t('mgmt.openId.newSecret')}</DialogTitle>
        <DialogContent>
          <DialogContentText>
            {t('mgmt.openId.secretUpdated')}
            
            <Stack direction="row" spacing={2} style={{ marginTop: '10px', width: '100%', maxWidth: '100%' }}>
              <div style={{overflowX: 'scroll', float: 'left', width: '100%', padding: '5px', background:'rgba(0,0,0,0.15)', whiteSpace: 'nowrap', wordBreak: 'keep-all', overflow: 'auto', fontStyle: 'italic'}}>
                {newSecret}
              </div>
              <IconButton size="large" style={{float: 'left'}} aria-label="copy" onClick={
                  () => {
                      navigator.clipboard.writeText(newSecret);
                  }
              }>
                  <CopyOutlined  />
              </IconButton>
            </Stack>
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setNewSecret(false)}>{t('global.close')}</Button>
        </DialogActions>
      </Dialog>}

      
      <Alert severity="warning" icon={<WarningOutlined />}>
        {t('mgmt.openId.experimentalWarning')}
      </Alert>

      {clients && <PrettyTableView
        data={clients}
        getKey={(r) => r.id}
        onRowClick={(r) => {
          setClientId(r.id);
          setOpenNewModal(true);
        }}
        columns={[
          {
            title: '',
            field: (r) => <img className="loading-image" alt="" src={getClientIcon(r)} width="64px" />,
            style: {
              textAlign: 'center',
              width: '100px',
            },
          },
          {
            title: 'Client ID',
            search: (r) => r.id,
            style: {
              textDecoration: 'inherit',
            },
            underline: true,
            field: (r) => <>
              <div style={{ display: 'inline-block', textDecoration: 'inherit', fontSize: '125%', color: isDark ? theme.palette.primary.light : theme.palette.primary.dark }}>{r.id}</div><br />
            </>
          },
          {
            title: t('mgmt.openId.redirectUri'),
            screenMin: 'sm',
            search: (r) => r.redirect,
            field: (r) => r.redirect,
          },
          {
            title: '', clickable: true, field: (r, k) => <>
              <Button variant="contained" color="primary" startIcon={<ArrowRightOutlined />} onClick={() => {
                generateNewSecret(r.id)
              }}>{t('mgmt.openId.resetSecret')}</Button>&nbsp;&nbsp;
              <DeleteButton onDelete={(event) => deleteClient(event, k)} />
            </>,
          },
        ]}
      />}
      {
        !clients && <div style={{ textAlign: 'center' }}>
          <CircularProgress />
        </div>
      }
    </>}
  </div>;
}

export default OpenIdList;