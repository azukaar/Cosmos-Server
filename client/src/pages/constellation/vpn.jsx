import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import AddDeviceModal from "./addDevice";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, LaptopOutlined, MobileOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, Stack } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import { isDomain } from "../../utils/indexs";
import ConfirmModal from "../../components/confirmModal";
import UploadButtons from "../../components/fileUpload";
import { useClientInfos } from "../../utils/hooks";
import { Trans, useTranslation } from 'react-i18next';

const getDefaultConstellationHostname = (config) => {
  // if domain is set, use it
  if(isDomain(config.HTTPConfig.Hostname)) {
    return "vpn." + config.HTTPConfig.Hostname;
  } else {
    return config.HTTPConfig.Hostname;
  }
}

export const ConstellationVPN = () => {
  const { t } = useTranslation();
  const [config, setConfig] = useState(null);
  const [users, setUsers] = useState(null);
  const [devices, setDevices] = useState(null);
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  const refreshConfig = async () => {
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setDevices((await API.constellation.list()).data || []);
    if(isAdmin)
      setUsers((await API.users.list()).data || []);
    else 
      setUsers([]);
  };

  useEffect(() => {
    refreshConfig();
  }, []);

  const getIcon = (r) => {
    if (r.deviceName.toLowerCase().includes("mobile") || r.deviceName.toLowerCase().includes("phone")) {
      return <MobileOutlined />
    }
    else if (r.deviceName.toLowerCase().includes("laptop") || r.deviceName.toLowerCase().includes("computer")) {
      return <LaptopOutlined />
    } else if (r.deviceName.toLowerCase().includes("desktop")) {
      return <DesktopOutlined />
    } else if (r.deviceName.toLowerCase().includes("tablet")) {
      return <TabletOutlined />
    } else if (r.deviceName.toLowerCase().includes("lighthouse") || r.deviceName.toLowerCase().includes("server")) {
      return <CompassOutlined />
    } else {
      return <CloudOutlined />
    }
  }

  return <>
    {(devices && config && users) ? <>
      <Stack spacing={2} style={{maxWidth: "1000px"}}>
      <div>
        <Alert severity="info"><Trans i18nKey="ConstellationDescription">
          Constellation is a VPN that runs inside your Cosmos network. It automatically
          connects all your devices together, and allows you to access them from anywhere.
          Please refer to the <a href="https://cosmos-cloud.io/doc/61 Constellation VPN/" target="_blank">documentation</a> for more information.
          In order to connect, please use the <a href="https://cosmos-cloud.io/clients" target="_blank">Constellation App</a>.
          Constellation is currently free to use until the end of the beta, planned January 2024.
        </Trans></Alert>
        <MainCard title={t('ConstellationSetup')} content={config.constellationIP}>
          <Stack spacing={2}>
          {config.ConstellationConfig.Enabled && config.ConstellationConfig.SlaveMode && <>
            <Alert severity="info">
              {t('ConstellationExternalNetwork')}
            </Alert>
          </>}  
          <Formik
            initialValues={{
              Enabled: config.ConstellationConfig.Enabled,
              PrivateNode: config.ConstellationConfig.PrivateNode,
              IsRelay: config.ConstellationConfig.NebulaConfig.Relay.AMRelay,
              ConstellationHostname: (config.ConstellationConfig.ConstellationHostname && config.ConstellationConfig.ConstellationHostname != "") ? config.ConstellationConfig.ConstellationHostname :
                getDefaultConstellationHostname(config)
            }}
            onSubmit={(values) => {
              let newConfig = { ...config };
              newConfig.ConstellationConfig.Enabled = values.Enabled;
              newConfig.ConstellationConfig.PrivateNode = values.PrivateNode;
              newConfig.ConstellationConfig.NebulaConfig.Relay.AMRelay = values.IsRelay;
              newConfig.ConstellationConfig.ConstellationHostname = values.ConstellationHostname;
              setTimeout(() => {
                refreshConfig();
              }, 1500);
              return API.config.set(newConfig);
            }}
          >
            {(formik) => (
              <form onSubmit={formik.handleSubmit}>
                <Stack spacing={2}>        
                {formik.values.Enabled && <Stack spacing={2} direction="row">    
                  <Button
                      disableElevation
                      variant="outlined"
                      color="primary"
                      onClick={async () => {
                        await API.constellation.restart();
                      }}
                    >
                      {t('RestartVPNService')}
                  </Button>
                  <ApiModal callback={API.constellation.getLogs} label={t('ShowVPNLogs')} />
                  <ApiModal callback={API.constellation.getConfig} label={t('ShowVPNConfig')} />
                  <ConfirmModal
                    variant="outlined"
                    color="warning"
                    label={t('ResetNetwork')}
                    content={t('ResetnetworkText')}
                    callback={async () => {
                      await API.constellation.reset();
                      refreshConfig();
                    }}
                  />
                  </Stack>}
                  <CosmosCheckbox formik={formik} name="Enabled" label={t('ConstellationEnabled')} />
                  {config.ConstellationConfig.Enabled && !config.ConstellationConfig.SlaveMode && <>
                    {formik.values.Enabled && <>
                      <CosmosCheckbox formik={formik} name="IsRelay" label={t('RelayRequests')} />
                      <CosmosCheckbox formik={formik} name="PrivateNode" label={t('PrivateNode')} />
                      {!formik.values.PrivateNode && <>
                        <Alert severity="info">This is your Constellation hostname, that you will use to connect. If you are using a domain name, this needs to be different from your server's hostname. Whatever the domain you choose, it is very important that you make sure there is a A entry in your domain DNS pointing to this server. <strong>If you change this value, you will need to reset your network and reconnect all the clients!</strong></Alert>
                        <CosmosInputText formik={formik} name="ConstellationHostname" label={'Constellation '+t('Hostname')} />
                      </>}
                    </>}
                  </>}
                  <LoadingButton
                      disableElevation
                      loading={formik.isSubmitting}
                      type="submit"
                      variant="contained"
                      color="primary"
                    >
                      {t('Save')}
                  </LoadingButton>
                  <UploadButtons
                    accept=".yml,.yaml"
                    label={t('ExternalNetworkConfig')}
                    variant="outlined"
                    fullWidth
                    OnChange={async (e) => {
                      let file = e.target.files[0];
                      await API.constellation.connect(file);
                      setTimeout(() => {
                        refreshConfig();
                      }, 1000);
                    }}
                  />
                </Stack>
              </form>
            )}
          </Formik>
          </Stack>
        </MainCard>
      </div>
      {config.ConstellationConfig.Enabled && !config.ConstellationConfig.SlaveMode && <>
      <CosmosFormDivider title={"Devices"} />
      <PrettyTableView 
          data={devices.filter((d) => !d.blocked)}
          getKey={(r) => r.deviceName}
          buttons={[
            <AddDeviceModal users={users} config={config} refreshConfig={refreshConfig} devices={devices}/>,
          ]}
          columns={[
              {
                  title: '',
                  field: getIcon,
              },
              {
                  title: t('DeviceName'),
                  field: (r) => <strong>{r.deviceName}</strong>,
              },
              {
                  title: t('Owner'),
                  field: (r) => <strong>{r.nickname}</strong>,
              },
              {
                  title: t('Type'),
                  field: (r) => <strong>{r.isLighthouse ? "Lighthouse" : "Client"}</strong>,
              },
              {
                  title: t('ConstellationIP'),
                  screenMin: 'md', 
                  field: (r) => r.ip,
              },
              {
                title: '',
                clickable: true,
                field: (r) => {
                  return <DeleteButton onDelete={async () => {
                    await API.constellation.block(r.nickname, r.deviceName, true);
                    refreshConfig();
                  }}></DeleteButton>
                }
              }
          ]}
        />
      </>}
        </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};