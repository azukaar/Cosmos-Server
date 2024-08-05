import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton, DeleteIconButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DeleteOutlined, DesktopOutlined, EditOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, ReloadOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, InputLabel, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import MergerDialog, { MergerDialogInternal } from "./mergerDialog";
import ResponsiveButton from "../../components/responseiveButton";
import MenuButton from "../../components/MenuButton";
import { useTranslation } from 'react-i18next';
import VMWarning from "./vmWarning";

export const StorageMerges = () => {
  const { t } = useTranslation();
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [mounts, setMounts] = useState([]);
  const [mergeDialog, setMergeDialog] = useState(null);
  const [loading, setLoading] = useState(false);
  const [containerized, setContainerized] = useState(false);

  const refresh = async () => {
    setLoading(true);
    let mountsData = await API.storage.mounts.list();
    let configAsync = await API.config.get();
    let status = await API.getStatus();

    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setMounts(mountsData.data);
    setContainerized(status.data.containerized);
    setLoading(false);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config) ? <>
      {mergeDialog && <MergerDialogInternal data={mergeDialog.data} refresh={refresh} unmount={mergeDialog.unmount} open={mergeDialog} setOpen={setMergeDialog} />}
      <Stack spacing={2}>
      {containerized && <VMWarning />}
      <div>
        <PrettyTableView 
          data={mounts.filter((mount) => mount.type === 'fuse.mergerfs')}
          getKey={(r) => `${r.device} - ${refresh.path}`}
          buttons={[
            <MergerDialog disk={{name: '/dev/sda'}} refresh={refresh} disabled={containerized}/>,
            <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
              refresh();
            }}>{t('global.refresh')}</ResponsiveButton>
          ]}
          columns={[
            {
              title: t('mgmt.storage.deviceTitle'),
              field: (r) => r.device,
            },
            { 
              title: t('mgmt.storage.pathTitle'),
              field: (r) => r.path,
            },
            { 
              title: t('mgmt.storage.typeTitle'),
              field: (r) => r.type,
            },
            { 
              title: t('mgmt.storage.optionsTitle'),
              screenMin: 'md',
              field: (r) => JSON.stringify(r.opts),

            },
            {
              title: '',
              clickable:true, 
              field: (r) => <>
                <DeleteIconButton disabled={containerized} onDelete={() => {
                  API.storage.mounts.unmount({ mountPoint: r.path, permanent: true }).then(() => {
                    refresh();
                  });
                }} />
              </>
            },
          ]}
        />

      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};