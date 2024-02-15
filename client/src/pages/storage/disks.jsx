import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CompassOutlined, DesktopOutlined, ExpandOutlined, LaptopOutlined, MinusCircleFilled, MobileOutlined, NodeCollapseOutlined, PlusCircleFilled, PlusCircleOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, InputLabel, Stack, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton, TreeItem, TreeView } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";

import diskIcon from '../../assets/images/icons/disk.svg';
import partIcon from '../../assets/images/icons/part.svg';
import lockIcon from '../../assets/images/icons/lock.svg';
import raidIcon from '../../assets/images/icons/database.svg';

const diskStyle = {
  width: "100%",
  padding: "20px",
  borderLeft: "1px solid #ccc",
  margin: "20px",
};

const icons = {
  disk: diskIcon,
  part: partIcon,
  crypt: lockIcon,
  raid: raidIcon,
  raid0: raidIcon,
  raid1: raidIcon,
  raid5: raidIcon,
  raid6: raidIcon,
}

const Disk = ({disk}) => {
  return <TreeItem style={diskStyle} nodeId={disk.name} label={
    <div style={{
      width: "100%",
      padding: "10px",
      border: "2px solid #eee",
      borderLeft: "2px solid green",
      backgroundColor: "white",
      color: 'black',
    }}>
      <Stack direction="row" justifyContent="space-between">
        <Stack direction="row" spacing={2} alignItems="center">
          <div>
            <Tooltip title={disk.type}>
              {icons[disk.type] ? <img width="64px" height={"64px"} src={icons[disk.type]} /> : <img width="64px" height={"64px"} src={icons["drive"]} />}
            </Tooltip>
          </div>
          <div>
          <div style={{fontWeight: 'bold'}}>{disk.name}</div>
          <div>{parseInt(disk.size / 1000000000)} GB</div>
          </div>
        </Stack>
        <Stack spacing={0} style={{textAlign: 'right', opacity: 0.8, fontSize: '80%'}}>
          {Object.keys(disk).filter(key => key !== "name" && key !== "children" &&key !== "rota" && key !== "type" && key !== "size").map((key, index) => {
            return <div key={index}>{key}: {
              (typeof disk[key] == "object" ? JSON.stringify(disk[key]) : disk[key])
            }</div>
          })}
        </Stack>
      </Stack>
    </div>
  }>
    {disk.children && disk.children.map((child, index) => {
      return <Disk disk={child} />
    })}
  </TreeItem>
}

export const StorageDisks = () => {
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [disks, setDisks] = useState([]);

  const refresh = async () => {
    let disksData = await API.storage.disks.list();
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setDisks(disksData.data);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config) ? <>
      <Stack spacing={2} style={{maxWidth: "1000px"}}>
      <div>
      <TreeView
        aria-label="Disks"
        defaultCollapseIcon={<MinusCircleFilled />}
        defaultExpandIcon={<PlusCircleFilled />}
      >
        {disks && disks.map((disk, index) => {
          return <Disk disk={disk} />
        })}

      </TreeView>
      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};