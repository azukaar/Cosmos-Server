import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { AlertFilled, CheckCircleOutlined, CloudOutlined, DeleteOutlined, EditOutlined, ExclamationCircleOutlined, ReloadOutlined, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { LoadingButton } from "@mui/lab";
import ConfirmModal, { ConfirmModalDirect } from "../../components/confirmModal";
import MenuButton from "../../components/MenuButton";
import diskIcon from '../../assets/images/icons/disk.svg';
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from 'react-i18next';
import VMWarning from "./vmWarning";
import { RaidDialog, RaidDialogInternal } from "./raidDialog";

const getRAIDStatus = (details) => {
  if (!details || !details["Array State"]) {
    return "error";
  }
  const state = details["Array State"].toLowerCase();
  if (state.includes("active") || state.includes("clean")) {
    return "success";
  } else if (state.includes("degraded")) {
    return "warning";
  }
  return "error";
};

const cleanRAIDStatus = (details) => {
  if (!details) {
    return '-';
  }

  return (
    <div style={{ maxHeight: '100px' }}>
      <div>State: {details["Array State"] || '-'}</div>
      <div>Active Devices: {details["Active Devices"] || '-'}</div>
      <div>Failed Devices: {details["Failed Devices"] || '0'}</div>
      <div>Spare Devices: {details["Spare Devices"] || '0'}</div>
    </div>
  );
};

export const RAIDArrays = () => {
  const { t } = useTranslation();
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [raids, setRaids] = useState([]);
  const [loading, setLoading] = useState(false);
  const [deleteRaid, setDeleteRaid] = useState(null);
  const [editOpened, setEditOpened] = useState(null);
  const [containerized, setContainerized] = useState(false);
  const [addDeviceRaid, setAddDeviceRaid] = useState(null);
  
  const refresh = async () => {
    setLoading(true);
    let raidsData = await API.storage.raid.list();
    let configAsync = await API.config.get();
    let status = await API.getStatus();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setRaids(raidsData.data);
    setLoading(false);
    setContainerized(status.data.containerized);
  };
  
  const apiDeleteRaid = async (name) => {
    setLoading(true);
    await API.storage.raid.delete(name);
    setLoading(false);
    setDeleteRaid(null);
    refresh();
  };

  const tryDeleteRaid = async (name) => {
    setDeleteRaid(name);
  };

  const resizeRaid = async (name) => {
    setLoading(true);
    await API.storage.raid.resize(name);
    setLoading(false);
    refresh();
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config) ? <>
      {deleteRaid && <ConfirmModalDirect
        title={t('mgmt.storage.raid.deleteTitle')}
        content={t('mgmt.storage.raid.confirmDeletion')}
        callback={() => apiDeleteRaid(deleteRaid)}
        onClose={() => setDeleteRaid(null)}
      />}
      <Stack spacing={2}>
        {containerized && <VMWarning />}
        <Stack direction="row" spacing={2} justifyContent="flex-start">  
          <RaidDialog refresh={refresh} disabled={containerized}/>
          <ResponsiveButton 
            variant="outlined" 
            startIcon={<ReloadOutlined />} 
            onClick={refresh}
          >
            {t('global.refresh')}
          </ResponsiveButton>
        </Stack>
      <div>
      {editOpened && <RaidDialogInternal 
        refresh={refresh} 
        open={editOpened} 
        setOpen={setEditOpened} 
        data={editOpened} 
      />}
      {raids && <PrettyTableView 
        data={raids}
        getKey={(r) => r.device}
        columns={[
          { 
            title: '', 
            field: () => <img width="64px" height="64px" src={diskIcon} alt="RAID" />,
            style: {
              textAlign: 'right',
              width: '64px',
            },
          },
          { 
            title: t('mgmt.storage.raid.nameTitle'), 
            field: (r) => r.device.split('/').pop(),
            style: {
              textAlign: 'center',
            },
          },
          {
            title: t('mgmt.storage.raid.levelTitle'),
            field: (r) => `RAID ${r["Raid Level"] || '-'}`
          },
          {
            title: t('mgmt.storage.raid.sizeTitle'),
            screenMin: 'sm',
            field: (r) => r["Array Size"] || '-'
          },
          {
            title: t('mgmt.storage.raid.devicesTitle'),
            field: (r) => (r.devices || []).map(d => 
              <div key={d.path}>
                {d.path} ({d.status})
              </div>
            )
          },
          {
            title: t('global.statusTitle'),
            screenMax: 'md',
            field: (r) => ({
              error: <ExclamationCircleOutlined style={{color: 'red'}}/>,
              warning: <WarningOutlined style={{color: 'orange'}}/>,
              success: <CheckCircleOutlined style={{color: 'green'}}/>,
            }[getRAIDStatus(r)])            
          },
          {
            title: '',
            screenMin: 'md',
            field: (r) => <div style={{maxWidth: '500px'}}>
              {{
                error: <Alert severity="error">{cleanRAIDStatus(r)}</Alert>,
                warning: <Alert severity="warning">{cleanRAIDStatus(r)}</Alert>,
                success: <Alert severity="success">{cleanRAIDStatus(r)}</Alert>,
              }[getRAIDStatus(r)]}
            </div>
          },
          {
            title: '',
            field: (r) => {
              return <div style={{position: 'relative'}}>
                <MenuButton>
                  <MenuItem 
                    disabled={loading || containerized} 
                    onClick={() => resizeRaid(r.device.split('/').pop())}
                  >
                    <ListItemIcon>
                      <CloudOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('mgmt.storage.raid.resize')}</ListItemText>
                  </MenuItem>
                  <MenuItem 
                    disabled={loading || containerized} 
                    onClick={() => setEditOpened(r)}
                  >
                    <ListItemIcon>
                      <EditOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('global.edit')}</ListItemText>
                  </MenuItem>
                  <MenuItem 
                    disabled={loading || containerized} 
                    onClick={() => tryDeleteRaid(r.device.split('/').pop())}
                  >
                    <ListItemIcon>
                      <DeleteOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('global.delete')}</ListItemText>
                  </MenuItem>
                </MenuButton>
              </div>
            }
          }
        ]}
      />}
      {
        !raids && <div style={{textAlign: 'center'}}>
          <CircularProgress />
        </div>
      }
      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>;
};