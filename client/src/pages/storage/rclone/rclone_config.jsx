import React, { useState, useEffect } from 'react';
import { Formik, Form, Field } from 'formik';
import { TextField, Select, MenuItem, FormControl, InputLabel, Button, Typography, Stack, Alert, CircularProgress } from '@mui/material';
import {ProvConfig, ProvAuth} from './rclone-providers';
import { CosmosCollapse } from '../../config/users/formShortcuts';
import * as API from '../../../api';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { useTranslation } from 'react-i18next';
import { CheckOutlined, PlusCircleOutlined, ReloadOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';
import RCloneNewConfig from './rclone_new';
import { DeleteButton } from '../../../components/delete';

const RClonePage = () => {
  const { t } = useTranslation();
  const [selectedProvider, setSelectedProvider] = useState('');
  const [providerOptions, setProviderOptions] = useState([]);
  const [providers, setProviders] = useState([]);
  const [pings, setPings] = useState({});
  const [configModal, setConfigModal] = useState(null);

  const refresh = () => {
    API.rclone.list().then((response) => {
      console.log(response);
      setProviders(response);

      Object.keys(response).forEach(provider => {
        API.rclone.pingStorage(provider).then((ping) => {
          setPings(pings => ({...pings, [provider]: {status: "OK"}}));
        }).catch((err) => {
          setPings(pings => ({...pings, [provider]: {status: "error", error: err}}));
        });
      });
    });
  }

  useEffect(() => {
    refresh();
  }, []);

  return (<>
    {configModal && <RCloneNewConfig onClose={() => {setConfigModal(false); refresh();}} open={configModal} setOpen={setConfigModal} />}

    {providers ? 
    <PrettyTableView 
    data={Object.keys(providers).map(key => ({name: key, ...providers[key]}))}
    getKey={(r, k) => `${providers.Name + k}`}
    buttons={[
      <ResponsiveButton startIcon={<PlusCircleOutlined />} variant="contained" onClick={() => setConfigModal(true)}>{t('mgmt.storage.rclone.create')}</ResponsiveButton>,
      <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
        refresh();
      }}>{t('global.refresh')}</ResponsiveButton>
    ]}
    columns={[
      {
        title: t('mgmt.storage.rclone.name'),
        field: (r) => <>{r.name}</>,
      },
      { 
        title: t('mgmt.storage.rclone.type'),
        field: (r) => r.type,
      },
      { 
        title: t('mgmt.storage.rclone.fullPath'),
        field: (r) => '/cosmos-storage/' + r.name,
      },
      { 
        title: t('mgmt.storage.rclone.status'),
        field: (r) => pings[r.name] ?
           (pings[r.name].status === "OK" ? <CheckOutlined style={{color: "green"}}/> : <Alert severity="error">{pings[r.name].error.message}</Alert>)
           : <CircularProgress />,
      },
      {
        title: '',
        field: (r) => <>
          <div style={{position: 'relative'}}>
            <DeleteButton onDelete={() => {
              API.rclone.deleteRemote(r.name).then(() => {
                refresh();
              });
            }} />
          </div>
        </>,
      }
    ]}
  />
    : <Stack spacing={2} direction="row" justifyContent="center">
      <CircularProgress />
    </Stack>}
  </>);
};

export default RClonePage;