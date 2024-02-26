import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, InputLabel, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import MergerDialog from "./mergerDialog";

export const StorageMerges = () => {
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
      <Stack spacing={2} style={{maxWidth: "1000px"}}>
      <div>
        <MergerDialog disk={{name: '/dev/sda'}} refresh={refresh}/>
      </div>
      <div>
        {mounts && mounts
        .filter((mount) => mount.type === 'fuse.mergerfs')
        .map((mount, index) => {
          return <div>
            <FolderOutlined/> {mount.device} - {mount.path} ({mount.type}) ({JSON.stringify(mount.opts)})
          </div>
        })}
      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};