import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, Checkbox, CircularProgress, FormControl, InputLabel, Select, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import { useTheme } from '@mui/material/styles';

export const MountPickerEntry = ({value, name, onChange, error, label, touched, formik}) => {
  const [mounts, setMounts] = useState([]);
  const [selectedMounts, setSelectedMounts] = useState(value || []);

  const refresh = async () => {
    let mountsData = await API.storage.mounts.list();
    setMounts(mountsData.data);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(mounts) ?
        // <Select
        //   name={name}
        //   label={label}
        // >
        //   {mounts.map((mount) => (
        //     <option key={`${mount.device} - ${mount.path}`} value={mount.path}>
        //       <FolderOutlined/> {mount.device} - {mount.path} ({mount.type})
        //     </option>
        //   ))}
        // </Select>
        <FormControl
          fullWidth
          variant="outlined"
          error={touched && Boolean(error)}
          style={{ marginBottom: '16px' }}
        >
          <InputLabel 
            shrink={value !== ''}
          >{label}</InputLabel>
          <CosmosSelect
            id={name}
            name={name}
            value={value}
            onChange={onChange}
            // label={name}
            formik={formik}
            options={mounts.map((mount) => ([
              mount.path,
              <span><FolderOutlined/> {mount.device} - {mount.path} ({mount.type})</span>,
            ]))}
          >
            {/* {mounts.map((mount) => (
              <option key={`${mount.device} - ${mount.path}`} value={mount.path}>
                <FolderOutlined/> {mount.device} - {mount.path} ({mount.type})
              </option>
            ))} */}
          </CosmosSelect>
        </FormControl>
    : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};