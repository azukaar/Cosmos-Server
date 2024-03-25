import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DeleteOutlined, DesktopOutlined, EditOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, InputLabel, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import SnapRAIDDialog from "./snapRaidDialog";
import MenuButton from "../../components/MenuButton";

export const StorageMounts = () => {
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [mounts, setMounts] = useState([]);

  const refresh = async () => {
    let mountsData = await API.storage.mounts.list();
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setMounts(mountsData.data);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config) ? <>
      <PrettyTableView 
        data={mounts}
        getKey={(r) => `${r.device} - ${refresh.path}`}
        columns={[
          {
            title: 'Device',
            field: (r) => <><FolderOutlined/>  {r.device}</>,
          },
          { 
            title: 'Path',
            field: (r) => r.path,
          },
          { 
            title: 'Type',
            field: (r) => r.type,
          },
          { 
            title: 'Options',
            field: (r) => JSON.stringify(r.opts),
          },
          {
            title: '',
            field: (r) => <>
              <div style={{position: 'relative'}}>
                <MenuButton>
                  <MenuItem>
                    <ListItemIcon>
                      <EditOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText disabled={false} onClick={() => tryDeleteRaid(r.Name)}>Edit</ListItemText>
                  </MenuItem>
                  <MenuItem>
                    <ListItemIcon>
                      <DeleteOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText disabled={false} onClick={() => tryDeleteRaid(r.Name)}>Delete</ListItemText>
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