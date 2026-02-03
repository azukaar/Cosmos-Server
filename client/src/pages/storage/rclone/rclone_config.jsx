import React, { useState, useEffect } from 'react';
import { Formik, Form, Field } from 'formik';
import { TextField, Select, MenuItem, FormControl, InputLabel, Button, Typography, Stack, Alert, CircularProgress, LinearProgress, Tooltip, IconButton } from '@mui/material';
import {ProvConfig, ProvAuth} from './rclone-providers';
import { CosmosCollapse } from '../../config/users/formShortcuts';
import * as API from '../../../api';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { useTranslation } from 'react-i18next';
import { ArrowDownOutlined, ArrowUpOutlined, CheckOutlined, DownOutlined, EditOutlined, ExclamationCircleOutlined, PlusCircleOutlined, ReloadOutlined, SettingOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';
import RCloneNewConfig from './rclone_new';
import { DeleteButton, DeleteIconButton } from '../../../components/delete';
import { simplifyNumber } from '../../dashboard/components/utils';
import RCloneTransfers from './rclone-transfers';
import MiniPlotComponent from '../../dashboard/components/mini-plot';
import { FilePickerButton } from '../../../components/filePicker';
import VMWarning from '../vmWarning';

const RClonePage = ({coStatus, containerized}) => {
  const { t } = useTranslation();
  const [selectedProvider, setSelectedProvider] = useState('');
  const [providerOptions, setProviderOptions] = useState([]);
  const [providers, setProviders] = useState([]);
  const [pings, setPings] = useState({});
  const [provStats, setProvStats] = useState({});
  const [configModal, setConfigModal] = useState(null);
  const [coreStats, setCoreStats] = useState({});

  const isTransfering = coreStats.transferring && coreStats.transferring.length > 0;

  const refreshStats = () => {
    API.rclone.coreStats().then((response) => {
      setCoreStats(response);
    });
  }

  const refresh = (delay = false) => {
    API.rclone.list().then((response) => {
      setProviders(response);
      if(delay) {
        setTimeout(() => {
          refreshStats();
        }, 2000);
      } else {
        refreshStats();
      }

      Object.keys(response).forEach(provider => {
        API.rclone.stats(provider).then((stats) => {
          setProvStats(provStats => ({...provStats, [provider]: stats}));
        });

        API.rclone.pingStorage(provider).then((ping) => {
          setPings(pings => ({...pings, [provider]: {status: "OK", used: ping.used, total: ping.total}}));
        }).catch((err) => {
          setPings(pings => ({...pings, [provider]: {status: "error", error: err}}));
        });
      });
    });
  }

  useEffect(() => {
    refresh();
  }, []);

  if(containerized) {
    return <VMWarning />;
  }

  return ( <div style={{ maxWidth: "1200px", margin: "auto" }}>
    {configModal && <RCloneNewConfig initialValues={configModal} onClose={() => {setConfigModal(false); refresh(true);}} open={configModal} setOpen={setConfigModal} />}
    {isTransfering && <Alert severity="warning">{t('mgmt.storage.rclone.transferWarning')}</Alert>}
    <MiniPlotComponent  metrics={[
      "cosmos.system.rclone.all.bytes",
      "cosmos.system.rclone.all.errors",
    ]} labels={{
      "cosmos.system.rclone.all.bytes": "Bytes",
      "cosmos.system.rclone.all.errors": "Errors",
    }} />
    <br/>
    {providers ? 
    <PrettyTableView 
    data={Object.keys(providers).map(key => ({name: key, ...providers[key]}))}
    getKey={(r, k) => `${r.name + k}`}
    buttons={[
      <ResponsiveButton startIcon={<PlusCircleOutlined />} variant="contained" onClick={() => setConfigModal(true)}>{t('mgmt.storage.rclone.create')}</ResponsiveButton>,
      <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
        refresh();
      }}>{t('global.refresh')}</ResponsiveButton>,    
      <RCloneTransfers refreshStats={refreshStats} transferring={coreStats.transferring ? coreStats.transferring : []}/>,  
      <ResponsiveButton variant="outlined" startIcon={<SettingOutlined />} onClick={() => {
        API.rclone.restart()
      }}>{t('mgmt.storage.rclone.remountAll')}</ResponsiveButton>
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
        screenMin: 'md',
        field: (r) => '/mnt/cosmos-storage-' + r.name,
      },
      { 
        title: t('mgmt.storage.rclone.status'),
        screenMin: 'md',
        field: (r) => {
          let percent = pings[r.name] ? pings[r.name].used / pings[r.name].total * 100 : 0;
          return pings[r.name] ?
           (pings[r.name].status === "OK" ? 
           <Stack direction="row" alignItems="center" spacing={2}>
            <CheckOutlined style={{color: "green"}}/>
            {(pings[r.name].hasOwnProperty('used') && pings[r.name].hasOwnProperty('total')) ?
            <Tooltip title={simplifyNumber(pings[r.name].used, 'B') + ' / ' + simplifyNumber(pings[r.name].total, 'B')}>
              <LinearProgress
              variant="determinate"
              color={percent > 95 ? 'error' : (percent > 75 ? 'warning' : 'info')}
              value={percent}
              style={{width: 'calc(100% - 50px)'}} />
              </Tooltip> : null}
            {provStats[r.name] ? <>
              {provStats[r.name].hasOwnProperty('diskCache') ? <>
                <Tooltip title={t('mgmt.storage.rclone.uploadsInProgress') + ': ' + provStats[r.name].diskCache.uploadsInProgress + '\n' + t('mgmt.storage.rclone.uploadsQueued') + ': ' + provStats[r.name].diskCache.uploadsQueued}>  
                  <Stack direction={'row'} alignItems="center" spacing={1}>
                    <div>
                      <ArrowUpOutlined />
                    </div>
                    <div>
                      {provStats[r.name].diskCache.uploadsInProgress + provStats[r.name].diskCache.uploadsQueued}
                    </div>
                  </Stack>
                </Tooltip>
                {provStats[r.name].diskCache.outOfSpace	&& <Tooltip title={t('mgmt.storage.rclone.outOfSpace')}>
                  <Stack direction={'row'} alignItems="center" spacing={1}>
                    <div>
                      <ExclamationCircleOutlined style={{color: 'red'}} />
                    </div>
                  </Stack>  
                </Tooltip>}
                {provStats[r.name].diskCache.erroredFiles	? <Tooltip title={t('mgmt.storage.rclone.erroredFiles', provStats[r.name].diskCache.erroredFiles)}>
                  <Stack direction={'row'} alignItems="center" spacing={1}>
                    <div>
                      <ExclamationCircleOutlined style={{color: 'red'}} />
                    </div>
                  </Stack>  
                </Tooltip> : ''}
              </> : null}
            </> : null}
           </Stack>  : <Alert severity="error">{pings[r.name].error.message}</Alert>)
           : <CircularProgress />
        },
      },
      {
        title: '',
        field: (r) => <>
          <div style={{position: 'relative'}}>
            <DeleteIconButton onDelete={() => {
              API.rclone.deleteRemote(r.name).then(() => {
                refresh();
              });
            }} />
            <IconButton variant="outlined"
             onClick={() => {
              setConfigModal(r);
            }}>
              <EditOutlined />
            </IconButton>
            <FilePickerButton storage={r.name} />
          </div>
        </>,
      }
    ]}
  />
    : <Stack spacing={2} direction="row" justifyContent="center">
      <CircularProgress />
    </Stack>}
  </div>);
};

export default RClonePage;