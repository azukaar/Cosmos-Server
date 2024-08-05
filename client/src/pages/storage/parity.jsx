import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { AlertFilled, CheckCircleOutlined, CloudOutlined, CloudServerOutlined, CompassOutlined, DeleteOutlined, DesktopOutlined, EditOutlined, ExclamationCircleOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, ReloadOutlined, TabletOutlined, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, Checkbox, CircularProgress, InputLabel, ListItemIcon, ListItemText, Menu, MenuItem, MenuList, Paper, Stack, Typography } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal, { ConfirmModalDirect } from "../../components/confirmModal";
import { crontabToText, isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import SnapRAIDDialog, { SnapRAIDDialogInternal } from "./snapRaidDialog";
import MenuButton from "../../components/MenuButton";
import diskIcon from '../../assets/images/icons/disk.svg';
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from 'react-i18next';
import VMWarning from "./vmWarning";

const getStatus = (status) => {
  if (!status) {
    return "error";
  }
  if (status.startsWith('Error')) {
    return "error";
  } else if (status.includes('WARNING!')) {
    return "warning";
  } else {
    return "success";
  }
}

const cleanStatus = (status) => {
  if (!status) {
    return '-'
  }

  if (status.split('..').length > 2) {
    status = status.split('..').slice(2).join('').slice(1)
  }
  
  // remove every between the first ----- and the last _____
  status = status.split('-----').shift() + status.split('_____').pop();
  status = status.replace('____', '')
  status = status.replace('SnapRAID status report:', '')
  

  status = status.split('\n').map((line) => {
    return <div>{line}</div>;
  })

  return <div style={{maxHeight: '100px'}}>
    {status}
  </div>;
}

export const Parity = () => {
  const { t } = useTranslation();
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [parities, setParities] = useState([]);
  const [loading, setLoading] = useState(false);
  const [deleteRaid, setDeleteRaid] = useState(null);
  const [editOpened, setEditOpened] = useState(null);
  const [containerized, setContainerized] = useState(false);
  
  const refresh = async () => {
    setLoading(true);
    let paritiesData = await API.storage.snapRAID.list();
    let configAsync = await API.config.get();
    let status = await API.getStatus();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setParities(paritiesData.data);
    setLoading(false);
    
    setContainerized(status.data.containerized);
  };
  
  const apiDeleteRaid = async (name) => {
    setLoading(true);
    let response = await API.storage.snapRAID.delete(name);
    setLoading(false);
    setDeleteRaid(null);
    refresh();
  }

  const tryDeleteRaid = async (name) => {
    setDeleteRaid(name);
  }

  const sync = async (name) => {
    setLoading(true);
    let response = await API.storage.snapRAID.sync(name);
    setLoading(false);
  };

  const fix = async (name) => {
    setLoading(true);
    let response = await API.storage.snapRAID.fix(name);
    setLoading(false);
  };

  const setEnabled = async (name, enable) => {
    setLoading(true);
    let response = await API.storage.snapRAID.enable(name, enable);
    setLoading(false);
    await refresh();
  };

  const scrub =async (name) => {
    setLoading(true);
    let response = await API.storage.snapRAID.scrub(name);
    setLoading(false);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config) ? <>
      {deleteRaid && <ConfirmModalDirect
        title="Delete Parity"
        content={t('mgmt.storage.confirmParityDeletion')}
        callback={() => apiDeleteRaid(deleteRaid)}
        onClose={() => setDeleteRaid(null)}
      />}
      <Stack spacing={2}>
        {containerized && <VMWarning />}
        <Stack direction="row" spacing={2} justifyContent="flex-start">  
          <SnapRAIDDialog refresh={refresh} disabled={containerized}/>
          <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
            refresh();
          }}>{t('global.refresh')}</ResponsiveButton>
        </Stack>
      <div>
      {editOpened && <SnapRAIDDialogInternal refresh={refresh} open={editOpened} setOpen={setEditOpened} data={editOpened} />}
      {parities && <PrettyTableView 
        data={parities}
        getKey={(r) => r.Name}
        columns={[
          { 
            title: '', 
            field: (r) => <img width="64px" height={"64px"} src={diskIcon} />,
            style: {
              textAlign: 'right',
              width: '64px',
            },
          },
          { 
            title: '', 
            field: (r) => r.Name,
            style: {
              textAlign: 'center',
            },
          },
          {
            title: t('global.enabled'), 
            clickable:true, 
            field: (r, k) => <Checkbox disabled={loading} size='large' color={r.Enabled ? 'success' : 'default'}
              onChange={() => setEnabled(r.Name, !r.Enabled)}
              checked={r.Enabled}
            />,
          },
          {
            title: t('mgmt.storage.parityDisksTitle'),
            field: (r) => r.Parity ? r.Parity.map(d => <div>{d}</div>) : '-'
          },
          {
            title: t('mgmt.storage.dataDisksTitle'),
            field: (r) => r.Parity ? Object.keys(r.Data).map(d => <div>
              {d}: {r.Data[d]}
            </div>) : '-'
          },
          {
            title: t('mgmt.storage.syncScrubIntervalTitle'),
            screenMin: 'sm',
            field: (r) => <div>Sync: {crontabToText(r.SyncCrontab, t)}<br/>Scrub: {crontabToText(r.ScrubCrontab, t)}</div>
          },
          {
            title: t('global.statusTitle'),
            screenMax: 'md',
            field: (r) => ({
              error: <ExclamationCircleOutlined style={{color: 'red'}}/>,
              warning: <WarningOutlined style={{color: 'orange'}}/>,
              success: <CheckCircleOutlined style={{color: 'green'}}/>,
            }[getStatus(r.Status)])            
          },
          {
            title: '',
            screenMin: 'md',
            field: (r) => <div style={{maxWidth: '500px'}}>
              {{
                error: <Alert severity="error">{cleanStatus(r.Status)}</Alert>,
                warning: <Alert severity="warning">{cleanStatus(r.Status)}</Alert>,
                success: <Alert severity="success">{cleanStatus(r.Status)}</Alert>,
              }[getStatus(r.Status)]}
            </div>
          },
          {
            title: '',
            field: (r) => {
              return <div style={{position: 'relative'}}>
                <MenuButton>
                  <MenuItem disabled={loading || containerized} onClick={() => setEditOpened(r)}>
                    <ListItemIcon>
                      <EditOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('global.edit')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading || containerized} onClick={() => sync(r.Name)}>
                    <ListItemIcon>
                      <CloudOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('mgmt.storage.list.syncText')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading || containerized} onClick={() => scrub(r.Name)}>
                    <ListItemIcon>
                      <CompassOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('mgmt.storage.list.scrubText')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading || containerized} onClick={() => fix(r.Name)}>
                    <ListItemIcon>
                      <CloudOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('mgmt.storage.list.fixText')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading || containerized} onClick={() => tryDeleteRaid(r.Name)}>
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
        !parities && <div style={{textAlign: 'center'}}>
          <CircularProgress />
        </div>
      }
      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};