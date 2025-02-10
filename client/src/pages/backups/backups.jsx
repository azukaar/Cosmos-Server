import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { CloudOutlined, CloudServerOutlined, DeleteOutlined, EditOutlined, LoadingOutlined, ReloadOutlined } from "@ant-design/icons";
import { Checkbox, CircularProgress, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { crontabToText } from "../../utils/indexs";
import MenuButton from "../../components/MenuButton";
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from "react-i18next";
import { ConfirmModalDirect } from "../../components/confirmModal";
import BackupDialog, { BackupDialogInternal } from "./backupDialog";

export const Backups = ({pathFilters}) => {
  const { t } = useTranslation();
  // const [isAdmin, setIsAdmin] = useState(false);
  let isAdmin = true;
  const [config, setConfig] = useState(null);
  const [backups, setBackups] = useState([]);
  const [loading, setLoading] = useState(false);
  const [deleteBackup, setDeleteBackup] = useState(null);
  const [editOpened, setEditOpened] = useState(null);
  
  const refresh = async () => {
    setLoading(true);
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    let b = Object.values(configAsync.data.Backup.Backups || {});
    b = pathFilters ? b.filter((b) => {
      let found = false;
      for (let i = 0; i < pathFilters.length; i++) {
        if (b.Source.startsWith(pathFilters[i])) {
          found = true;
          break;
        }
      }
      return found;
    }) : b;
    setBackups(b || []);
    setLoading(false);
  };
  
  const apiDeleteBackup = async (name) => {
    setLoading(name);
    try {
      await API.backups.removeBackup(name, true);
    } catch(e) {
      setLoading(false);
      setDeleteBackup(null);
    }
    refresh();
  }

  const tryDeleteBackup = async (name) => {
    setDeleteBackup(name);
  }

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config) ? <>
      {deleteBackup && <ConfirmModalDirect
        title="Delete Backup"
        content={t('mgmt.backup.confirmBackupDeletion')}
        callback={() => apiDeleteBackup(deleteBackup)}
        onClose={() => setDeleteBackup(null)}
      />}
      <Stack spacing={2}>
        {!pathFilters && <Stack direction="row" spacing={2} justifyContent="flex-start">  
          <BackupDialog refresh={refresh}/>
          <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={refresh}>
            {t('global.refresh')}
          </ResponsiveButton>
        </Stack>}
      <div>
      {editOpened && <BackupDialogInternal 
        refresh={refresh} 
        open={editOpened} 
        setOpen={setEditOpened} 
        data={editOpened} 
      />}
      
      {backups && <PrettyTableView 
        linkTo={r => '/cosmos-ui/backups/' + r.Name}
        data={backups}
        getKey={(r) => r.Name}
        columns={[
          { 
            title: '', 
            field: () => <CloudServerOutlined />,
            style: {
              textAlign: 'right',
              width: '64px',
            },
          },
          { 
            title: t('global.name'), 
            field: (r) => r.Name,
            underline: true,
          },
          {
            title: t('mgmt.backup.sourceTitle'),
            field: (r) => (loading && r.Name == loading) ? <div><LoadingOutlined /> Deleting...</div> : r.Source,
            screenMin: 'sm',
            underline: true,
          },
          {
            title: t('mgmt.backup.repositoryTitle'),
            field: (r) => (loading && r.Name == loading) ? '' : r.Repository,
            underline: true,
          },
          {
            title: t('mgmt.backup.scheduleTitle'),
            screenMin: 'md',
            field: (r) => (loading && r.Name == loading) ? '' : crontabToText(r.Crontab, t),
            underline: true,
          },
          {
            title: '',
            clickable:true, 
            field: (r) => {
              return <div style={{position: 'relative'}}>
                <MenuButton>
                  <MenuItem disabled={loading} onClick={() => {
                    window.open('/cosmos-ui/backups/' + r.Name, '_blank');
                  }}>
                    <ListItemIcon>
                      <EditOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('global.open')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading} onClick={() => setEditOpened(r)}>
                    <ListItemIcon>
                      <EditOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText>{t('global.edit')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading} onClick={() => tryDeleteBackup(r.Name)}>
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
      {!backups && 
        <div style={{textAlign: 'center'}}>
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