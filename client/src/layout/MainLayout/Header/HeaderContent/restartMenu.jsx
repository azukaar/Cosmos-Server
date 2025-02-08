import { ListItemIcon, ListItemText, MenuItem } from "@mui/material";
import MenuButton from "../../../../components/MenuButton";
import { PoweroffOutlined, SyncOutlined } from "@ant-design/icons";
import { useTranslation } from "react-i18next";
import * as API  from "../../../../api";
import React from "react";
import RestartModal from "../../../../pages/config/users/restart";
import { ConfirmModalDirect } from "../../../../components/confirmModal";
import { useClientInfos } from "../../../../utils/hooks";

const RestartMenu = () => {
  const { t } = useTranslation();
  const [openResartModal, setOpenRestartModal] = React.useState(false);
  const [openRestartServerModal, setOpenRestartServerModal] = React.useState(false);
  const [status, setStatus] = React.useState({});
  const {role} = useClientInfos();
  const isAdmin = role === "2";
  
  React.useEffect(() => {
    API.getStatus().then((res) => {
      setStatus(res.data);
    });
  }, []);

  const restartServer = API.restartServer;

  return isAdmin ? <>
  <RestartModal openModal={openResartModal} setOpenModal={setOpenRestartModal} />
  <RestartModal openModal={openRestartServerModal} setOpenModal={setOpenRestartModal} isHostMachine/>

  <MenuButton size="medium" icon={<PoweroffOutlined />}>
    <MenuItem disabled={status.containerized} onClick={() => setOpenRestartServerModal(true)}>
      <ListItemText>{t('global.restartServer')}</ListItemText>
    </MenuItem>
    <MenuItem disabled={false} onClick={() => setOpenRestartModal(true)}>
      <ListItemText>{t('global.restartCosmos')}</ListItemText>
    </MenuItem>
  </MenuButton>
  </> : <></>;
};

export default RestartMenu;