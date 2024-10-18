import React, { useState, useEffect } from 'react';
import { Stack, CircularProgress, IconButton } from '@mui/material';
import * as API from '../../../api';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { useTranslation } from 'react-i18next';
import { EditOutlined, PlusCircleOutlined, ReloadOutlined, SettingOutlined } from '@ant-design/icons';
import ResponsiveButton from '../../../components/responseiveButton';
import { DeleteIconButton } from '../../../components/delete';
import VMWarning from '../vmWarning';
import RCloneNewServeConfig from './rclone_serve_new';

const RCloneServePage = ({containerized}) => {
  const { t } = useTranslation();
  const [configModal, setConfigModal] = useState(null);
  const [config, setConfig] = React.useState(null);

  const refresh = () => {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  useEffect(() => {
    refresh();
  }, []);

  if(containerized) {
    return <VMWarning />;
  }

  return (<>
    {configModal && <RCloneNewServeConfig initialValues={configModal} onClose={() => {setConfigModal(false); refresh();}} open={configModal} setOpen={setConfigModal} />}
    <br/>
    {config ? 
    <PrettyTableView 
    data={(config.RemoteStorage ? (config.RemoteStorage.Shares) : [])}
    getKey={(r, k) => `${r.name + k}`}
    buttons={[
      <ResponsiveButton startIcon={<PlusCircleOutlined />} variant="contained" onClick={() => setConfigModal(true)}>{t('mgmt.storage.rclone.create-serve')}</ResponsiveButton>,
      <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
        refresh();
      }}>{t('global.refresh')}</ResponsiveButton>,    
      <ResponsiveButton variant="outlined" startIcon={<SettingOutlined />} onClick={() => {
        API.rclone.restart()
      }}>{t('mgmt.storage.rclone.remountAll')}</ResponsiveButton>
    ]}
    columns={[
      {
        title: t('mgmt.storage.rclone.name'),
        field: (r) => <>{r.Name}</>,
      },
      { 
        title: t('mgmt.storage.rclone.protocol'),
        field: (r) => r.Protocol,
      },
      {
        title: '',
        field: (r) => <>
          <div style={{position: 'relative'}}>
            <DeleteIconButton onDelete={() => {
              return API.config.get().then(({data}) => {
                let shares = data.RemoteStorage.Shares || [];
                shares = shares.filter((s) => s.Name !== r.Name);

                return API.config.set({
                  ...data,
                  RemoteStorage: {
                    ...data.RemoteStorage,
                    Shares: shares
                  }
                })
              });
            }} />
            <IconButton variant="outlined"
             onClick={() => {
              setConfigModal(r);
            }}>
              <EditOutlined />
            </IconButton>
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

export default RCloneServePage;