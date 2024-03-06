import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { AlertFilled, CheckCircleOutlined, CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, ExclamationCircleOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, TabletOutlined, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, Checkbox, CircularProgress, InputLabel, ListItemIcon, ListItemText, Menu, MenuItem, MenuList, Paper, Stack, Typography } from "@mui/material";
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
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [parities, setParities] = useState([]);
  const [loading, setLoading] = useState(false);
  
  const setEnabled = () => {}
  const refresh = async () => {
    let paritiesData = await API.storage.snapRAID.list();
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setParities(paritiesData.data);
  };

  const sync = async (name) => {
    setLoading(true);
    let response = await API.storage.snapRAID.sync(name);
    setLoading(false);
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
      <Stack spacing={2}>
      <SnapRAIDDialog refresh={refresh} />
      <div>
      {parities && <PrettyTableView 
        data={parities}
        getKey={(r) => r.Name}
        columns={[
          { 
            title: '', 
            field: (r) => r.Name,
            style: {
              textAlign: 'center',
            },
          },
          {
            title: 'Enabled', 
            clickable:true, 
            field: (r, k) => <Checkbox disabled={loading} size='large' color={!r.Disabled ? 'success' : 'default'}
              onChange={setEnabled(parities.indexOf(r))}
              checked={!r.Disabled}
            />,
          },
          {
            title: 'Parity Disks',
            field: (r) => r.Parity ? r.Parity.map(d => <div>{d}</div>) : '-'
          },
          {
            title: 'Data Disks',
            field: (r) => r.Parity ? r.Data.map(d => <div>{d}</div>) : '-'
          },
          {
            title: 'Sync/Scrub Intervals',
            screenMin: 'sm',
            field: (r) => r.SyncInterval + 'h / ' +   r.ScrubInterval + 'h'
          },
          {
            title: 'Status',
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
                  <MenuItem>
                    <ListItemIcon>
                      <CloudOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText disabled={loading} onClick={() => sync(r.Name)}>Sync</ListItemText>
                  </MenuItem>
                  <MenuItem>
                    <ListItemIcon>
                      <CompassOutlined fontSize="small" />
                    </ListItemIcon>
                    <ListItemText disabled={loading} onClick={() => scrub(r.Name)}>Scrub</ListItemText>
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