import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { CloudOutlined, CloudServerOutlined, DeleteOutlined, EditOutlined, ReloadOutlined } from "@ant-design/icons";
import { Checkbox, CircularProgress, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { crontabToText } from "../../utils/indexs";
import MenuButton from "../../components/MenuButton";
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from "react-i18next";
import { ConfirmModalDirect } from "../../components/confirmModal";
import BackupDialog, { BackupDialogInternal } from "./backupDialog";

export const Backups = () => {
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
    setBackups(configAsync.data.Backup.Backups || []);
    setLoading(false);
  };
  
  const apiDeleteBackup = async (name) => {
    setLoading(true);
    await API.backups.removeBackup(name, true);
    setLoading(false);
    setDeleteBackup(null);
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
        <Stack direction="row" spacing={2} justifyContent="flex-start">  
          <BackupDialog refresh={refresh}/>
          <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={refresh}>
            {t('global.refresh')}
          </ResponsiveButton>
        </Stack>
      <div>
      {editOpened && <BackupDialogInternal 
        refresh={refresh} 
        open={editOpened} 
        setOpen={setEditOpened} 
        data={editOpened} 
      />}
      {backups && <PrettyTableView 
        linkTo={r => '/cosmos-ui/backups/' + r.Name}
        data={Object.values(backups)}
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
            field: (r) => r.Source,
            underline: true,
          },
          {
            title: t('mgmt.backup.repositoryTitle'),
            field: (r) => r.Repository,
            underline: true,
          },
          {
            title: t('mgmt.backup.scheduleTitle'),
            screenMin: 'sm',
            field: (r) => crontabToText(r.Crontab, t),
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
                  <MenuItem disabled={loading} onClick={() => tryDeleteBackup(r.name)}>
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