import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { CloudOutlined, CloudServerOutlined, DeleteOutlined, EditOutlined, FolderOutlined, ReloadOutlined } from "@ant-design/icons";
import { Checkbox, CircularProgress, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { crontabToText } from "../../utils/indexs";
import MenuButton from "../../components/MenuButton";
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from "react-i18next";
import { ConfirmModalDirect } from "../../components/confirmModal";
import BackupDialog, { BackupDialogInternal } from "./backupDialog";
import {simplifyNumber} from '../dashboard/components/utils';

export const Repositories = () => {
  const { t } = useTranslation();
  // const [isAdmin, setIsAdmin] = useState(false);
  let isAdmin = true;
  const [repositories, setRepositories] = useState(null);
  const [loading, setLoading] = useState(false);
  const [deleteBackup, setDeleteBackup] = useState(null);
  const [editOpened, setEditOpened] = useState(null);
  
  const refresh = async () => {
    setLoading(true);
    let configAsync = await API.backups.listRepo();
    setRepositories(configAsync.data || []);
    setLoading(false);
  };
  
  useEffect(() => {
    refresh();
  }, []);

  if (loading) {
    return (
      <Stack direction="row" spacing={2} justifyContent="center" alignContent={"center"} alignItems={"center"}>
        <CircularProgress />
      </Stack>
    );
  }

  return  <div style={{ maxWidth: "1200px", margin: "auto" }}>
    {(repositories) ? <>
      <Stack direction="row" spacing={2} justifyContent="flex-start">  
        <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={refresh}>
          {t('global.refresh')}
        </ResponsiveButton>
      </Stack>
      <PrettyTableView 
        linkTo={r => '/cosmos-ui/backups/repo/' + r.id}
        data={Object.values(repositories)}
        getKey={(r) => r.Name}
        columns={[
          { 
            title: '', 
            field: () => <FolderOutlined />,
            style: {
              textAlign: 'right',
              width: '64px',
            },
          },
          { 
            title: t('mgmt.backup.repository'), 
            field: (r) => r.path,
            underline: true,
          },
          { 
            title: t('mgmt.backup.repository.size'), 
            field: (r) => simplifyNumber(r.stats.total_uncompressed_size, "B") + " (" + simplifyNumber(r.stats.total_size, "B") + " on disk)",
            underline: true,
          },
          { 
            title: t('mgmt.backup.repository.total_file_count'), 
            field: (r) => simplifyNumber(r.stats.total_file_count, ""),
            underline: true,
          },
          { 
            title: t('mgmt.backup.repository.snapshots_count'), 
            field: (r) => r.stats.snapshots_count,
            underline: true,
          },
        ]}
      />
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </div>;
};