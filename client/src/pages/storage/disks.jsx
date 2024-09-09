import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloseCircleOutlined, CloudOutlined, CompassOutlined, DesktopOutlined, ExpandOutlined, LaptopOutlined, MenuFoldOutlined, MenuOutlined, MinusCircleFilled, MobileOutlined, NodeCollapseOutlined, PlusCircleFilled, PlusCircleOutlined, ReloadOutlined, SettingFilled, TabletOutlined, WarningFilled, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, InputLabel, LinearProgress, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton, TreeItem, TreeView } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import { useTheme } from '@mui/material/styles';

import diskIcon from '../../assets/images/icons/disk.svg';
import partIcon from '../../assets/images/icons/part.svg';
import lockIcon from '../../assets/images/icons/lock.svg';
import raidIcon from '../../assets/images/icons/database.svg';
import { simplifyNumber } from "../dashboard/components/utils";
import LogsInModal from "../../components/logsInModal";
import MountDiskDialog from "./mountDiskDialog";
import PasswordModal from "../../components/passwordModal";
import FormatModal from "./FormatModal";
import MenuButton from "../../components/MenuButton";
import ResponsiveButton from "../../components/responseiveButton";
import SMARTDialog, { CompleteDataSMARTDisk, diskChip, diskColor, getSMARTDef, temperatureChip } from "./smart";
import { useTranslation } from 'react-i18next';
import VMWarning from "./vmWarning";

const diskStyle = {
  width: "100%",
  padding: "20px",
  borderLeft: "1px solid #ccc",
  margin: "15px",
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

const FormatButton = ({disk, refresh, disabled}) => {
  const { t } = useTranslation();
  const [formatting, setFormatting] = useState(false);
  const [loading, setLoading] = useState(false);
  const [passwordConfirm, setPasswordConfirm] = useState(false);

  return <>
    <LoadingButton
      disabled={disabled}
      loading={loading}
      onClick={() => {setPasswordConfirm(true); setLoading(true)}}
      variant="outlined"
      color="error"
      size="small"
      startIcon={<CloseCircleOutlined />}
    >
      {t('mgmt.storage.formatButton')}
    </LoadingButton>
    
    {passwordConfirm && <FormatModal
      OnClose={() => {
        setPasswordConfirm(false);
        setLoading(false);
      }}
      textInfos={t('mgmt.storage.formatModalText', {disk: disk.name})}
      cb={async (values) => {
        setPasswordConfirm(false);
        setFormatting(values);
      }} 
    />}
    
    {formatting && <LogsInModal
      request={(cb) => {
        return API.storage.disks.format({
          disk: disk.name,
          format: formatting.format,
          password: formatting.password,
        }, cb)
      }}
      initialLogs={[
        t('mgmt.storage.startFormatLog', {disk: disk.name})
      ]}
      alwaysShow={true}
      OnSuccess={() => {
        setLoading(false);
      }}
      OnClose={() => {
        setFormatting(false);
        refresh && refresh();
      }}
      title={t('mgmt.storage.formattingLog')+'...'}
    />}
  </>
}

const Disk = ({disk, refresh, SetSMARTDialogOpened, containerized}) => {
  const theme = useTheme();
  const darkMode = theme.palette.mode === 'dark';

  let percent = 0;
  if (disk.usage) {
    percent = Math.round((disk.usage / disk.size) * 100);
  }

  return <>
    <TreeItem sx={{
      '&.Mui-selected': {
        color: 'red !important',
        backgroundColor: 'red'
      }
    }} style={diskStyle} nodeId={disk.name} label={
      <div style={{
        width: "100%",
        padding: "10px",
        border: "1px solid #eee",
        borderRadius: "5px",
        borderColor: darkMode ? "#555" : "#eee",
        backgroundColor: darkMode ? "#333" : "#fff",
        color: darkMode ? "#fff" : "#000",
      }}>
        <Stack direction="row" justifyContent="space-between" style={{
            borderLeft: "4px solid " + (disk.smart ? diskColor(disk) : "gray")
          }}>
          <Stack direction="row" spacing={2} alignItems="center" sx={{paddingLeft: '5px'}}>
            <div>
              {(disk.smart && disk.smart.Temperature) ? <div style={{
                  position: 'absolute',
                  top: '10px',
                  left: '25px',
                  backgroundColor: 'rgba(0,0,0,0.8)',
                  color: 'white',
                  borderRadius: '5px',
                  padding: '0px 5px',
                }}>
                <span style={{fontSize: '80%', opacity: 0.8}}>
                  {(disk.smart ? diskChip(disk) : "⚪")} {disk.smart.Temperature}°C
                </span></div> : ""}

              
              <Tooltip title={disk.type}>
                {icons[disk.type] ? <img width="64px" height={"64px"} src={icons[disk.type]} /> : <img width="64px" height={"64px"} src={icons["drive"]} />}
              </Tooltip>
            </div>
            <div>
            <Stack spacing={1}>
            <div style={{fontWeight: 'bold'}}>
              {disk.name.replace("/dev/", "")} 
              <span style={{opacity: 0.8, fontSize: '80%'}}>
                {disk.rota ? " (HDD)" : " (SSD)"}
                {disk.fstype ? ` - ${disk.fstype}` : ""}
                {disk.model ? ` - ${disk.model}` : ""} {disk.mountpoint ? ` (${disk.mountpoint})` : ""}
              </span>
            </div>
              {disk.usage ? <>
              <div><LinearProgress
                variant="determinate"
                color={percent > 95 ? 'error' : (percent > 75 ? 'warning' : 'info')}
                value={percent} /></div>
              <div>{simplifyNumber(disk.usage, 'b')} / {simplifyNumber(disk.size, 'b')} ({percent}%)</div>
              </>: simplifyNumber(disk.size, 'b')}

              </Stack>
            </div>
          </Stack>
          <Stack direction={"row"} spacing={2} style={{textAlign: 'right'}}>
            <Stack spacing={0} style={{textAlign: 'right', opacity: 0.8, fontSize: '80%'}}>
              {Object.keys(disk).filter(key => key !== "name" && key !== "children" && key !== "smart"&& key !== "fstype"&& key !== "mountpoint"&& key !== "model" && key !== "usage" &&key !== "rota" && key !== "type" && key !== "size").map((key, index) => {
                return <div key={index}>{key}: {
                  (typeof disk[key] == "object" ? JSON.stringify(disk[key]) : disk[key])
                }</div>
              })}
            </Stack>
            <Stack spacing={2} direction="column" justifyContent={"center"}>
              {(disk.type == "disk" || disk.type == "part") ? <FormatButton disabled={containerized} disk={disk} refresh={refresh}/> : ""}
              
              {disk.mountpoint ? <MountDiskDialog disabled={containerized} disk={disk} unmount={true} refresh={refresh} /> : ""}
              
              {(
                (disk.type == "part" || (disk.type == "disk" && (!disk.children || !disk.children.length))) && 
                disk.fstype &&
                disk.fstype !== "swap" &&
                disk.fstype !== "linux_raid_member" &&
                !disk.mountpoint
              ) ? <MountDiskDialog disk={disk} refresh={refresh} /> : ""}
              {disk.type == "disk" ? <Button onClick={() => SetSMARTDialogOpened(disk)} variant="outlined" size="small" startIcon={<CompassOutlined />}>S.M.A.R.T.</Button> : ""}
            </Stack>
          </Stack>
        </Stack>
      </div>
    }>
      {disk.children && disk.children.map((child, index) => {
        return <Disk containerized={containerized} disk={child} refresh={refresh} SetSMARTDialogOpened={SetSMARTDialogOpened}/>
      })}
    </TreeItem>
  </>;
}

export const StorageDisks = () => {
  const { t } = useTranslation();
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [disks, setDisks] = useState([]);
  const [containerized, setContainerized] = useState(false);
  const [SMARTDialogOpened, SetSMARTDialogOpened] = useState(false);

  const refresh = async () => {
    let disksData = await API.storage.disks.list();
    let configAsync = await API.config.get();
    let status = await API.getStatus();

    disksData = disksData.data.map((disk) => {
      return CompleteDataSMARTDisk(disk);
    });

    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setDisks(disksData);
    setContainerized(status.data.containerized);
  };

  useEffect(() => {
    (async () => {
      await getSMARTDef();
      refresh();
    })();
  }, []);

  return <>
    {(config) ? <>
      <SMARTDialog disk={SMARTDialogOpened} OnClose={() => SetSMARTDialogOpened(false)} />

      <Stack spacing={2} style={{maxWidth: "1000px"}}>
      {containerized && <VMWarning />}
      <div>
        <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
            refresh();
        }}>{t('global.refresh')}</ResponsiveButton>
      </div>
      <div>
      <TreeView
        selected={[""]}
        style={{userSelect: 'none' }}
        aria-label="Disks"
        defaultCollapseIcon={<MinusCircleFilled />}
        defaultExpandIcon={<PlusCircleFilled />}
      >
        {disks && disks.map((disk, index) => {
          return <Disk containerized={containerized} SetSMARTDialogOpened={SetSMARTDialogOpened} disk={disk} refresh={async () => {
            await new Promise(r => setTimeout(r, 1000));
            await refresh()
          }} />
        })}

      </TreeView>
      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};