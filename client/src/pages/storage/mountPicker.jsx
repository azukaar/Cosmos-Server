import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, FolderOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, Checkbox, CircularProgress, FormControl, InputLabel, Select, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import { useTheme } from '@mui/material/styles';
import { Trans, useTranslation } from 'react-i18next';

export const MountPicker = ({diskMode, multiselect, value, onChange}) => {
  const { t } = useTranslation();
  const [mounts, setMounts] = useState([]);
  const [selectedMounts, setSelectedMounts] = useState(value || []);
  const theme = useTheme();
  const darkMode = theme.palette.mode === 'dark';

  useEffect(() => {
    onChange && onChange(selectedMounts);
  }, [selectedMounts]);

  const handleClick = (e, path) => {
    e.preventDefault();
    setSelectedMounts((prev) => {
      if (prev.includes(path)) {
        return prev.filter((p) => p !== path);
      } else {
        console.log([...prev, path]);
        return [...prev, path];
      }
    });
    return false;
  };

  const refresh = async () => {
    let mountsData = await API.storage.mounts.list();
    setMounts(mountsData.data);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(mounts) ? <>
      <Stack spacing={2}>
      <style>{`
        .native-multiselect option:checked,
        .native-multiselect option:hover,
        .native-multiselect option:active,
        .native-multiselect option:focus,
        .native-multiselect option {
          outline: none;
          background-color: transparent;
          color: ${darkMode ? 'white' : 'black'};
        }
      `}</style>
      <div> 
        <FormControl sx={{ width: '100%' }}>
        <InputLabel shrink htmlFor="select-multiple-native">
          {t('mgmt.storage.mountPicker')}
        </InputLabel>
        <Select
          multiple
          native
          className={'native-multiselect'}
          value={selectedMounts}
          label={t('mgmt.storage.mountPicker')}
          inputProps={{
            id: 'select-multiple-native',
          }}
        >
          {mounts.map((mount) => (
            <option key={`${mount.device} - ${mount.path}`} value={mount.path} onClick={(e) => {handleClick(e, mount.path)}}>
              <Checkbox checked={selectedMounts.indexOf(mount.path) > -1} />
              <FolderOutlined/> {mount.device} - {mount.path} ({mount.type})
            </option>
          ))}
        </Select>
      </FormControl>

      </div>  
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};