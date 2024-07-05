import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DeleteOutlined, DesktopOutlined, EditOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, PlusCircleOutlined, ReloadOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, InputLabel, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal, { ConfirmModalDirect } from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import SnapRAIDDialog from "./snapRaidDialog";
import MenuButton from "../../components/MenuButton";
import MountDialog, { MountDialogInternal } from "./mountDialog";
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from 'react-i18next';

export const StorageMounts = () => {
  const { t } = useTranslation();
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [mounts, setMounts] = useState([]);
  const [mountDialog, setMountDialog] = useState(null);
  const [loading, setLoading] = useState(false);

  const refresh = async () => {
    setLoading(true);
    let mountsData = await API.storage.mounts.list();
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setMounts(mountsData.data);
    setLoading(false);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {mountDialog && <MountDialogInternal data={mountDialog.data} refresh={refresh} unmount={mountDialog.unmount} open={mountDialog} setOpen={setMountDialog} />}
    {(config) ? <>
      <PrettyTableView 
        data={mounts}
        getKey={(r) => `${r.device} - ${refresh.path}`}
        buttons={[
          <ResponsiveButton startIcon={<PlusCircleOutlined />} variant="contained" onClick={() => setMountDialog({data: null, unmount: false})}>{t('NewMount')}</ResponsiveButton>,
          <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
            refresh();
          }}>{t('Refresh')}</ResponsiveButton>
        ]}
        columns={[
          {
            title: t('Device'),
            field: (r) => <><FolderOutlined/>  {r.device}</>,
          },
          { 
            title: t('Path'),
            field: (r) => r.path,
          },
          { 
            title: t('Type'),
            field: (r) => r.type,
          },
          { 
            title: t('Options'),
            field: (r) => JSON.stringify(r.opts),
          },
          {
            title: '',
            field: (r) => <>
              <div style={{position: 'relative'}}>
                <MenuButton>
                  <MenuItem disabled={!r.device.startsWith('/dev/') || loading} onClick={() => setMountDialog({data: r, unmount: false})}>
                    <ListItemIcon>
                      <EditOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText >{t('Edit')}</ListItemText>
                  </MenuItem>
                  <MenuItem disabled={loading} onClick={() => setMountDialog({data: r, unmount: true})}>
                    <ListItemIcon>
                      <DeleteOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText >{t('Unmount')}</ListItemText>
                  </MenuItem>
                </MenuButton>
              </div>
            </>
          },
        ]}
      />
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};