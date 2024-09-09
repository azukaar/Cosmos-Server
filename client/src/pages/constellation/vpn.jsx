import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import AddDeviceModal from "./addDevice";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, LaptopOutlined, MobileOutlined, SyncOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, IconButton, Stack, Tooltip } from "@mui/material";
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
  const [resynDevice, setResyncDevice] = useState(null); // [nickname, deviceName]
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
      {resynDevice &&
        <ResyncDeviceModal nickname={resynDevice[0]} deviceName={resynDevice[1]} OnClose={
          () => setResyncDevice(null)
        } />
      }
      <Stack spacing={2} style={{maxWidth: "1000px"}}>
      <div>
        <Alert severity="info">
          <Trans i18nKey="mgmt.constellation.setupText"
            components={[<a href="https://cosmos-cloud.io/doc/61 Constellation VPN/" target="_blank"></a>, <a href="https://cosmos-cloud.io/clients" target="_blank"></a>]}
          />
        </Alert>
        <MainCard title={t('mgmt.constellation.setupTitle')} content={config.constellationIP}>
          <Stack spacing={2}>
          {config.ConstellationConfig.Enabled && config.ConstellationConfig.SlaveMode && <>
            <Alert severity="info">
              {t('mgmt.constellation.externalText')}
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
                      {t('mgmt.constellation.restartButton')}
                  </Button>
                  <ApiModal callback={API.constellation.getLogs} label={t('mgmt.constellation.showLogsButton')} />
                  <ApiModal callback={API.constellation.getConfig} label={t('mgmt.constellation.showConfigButton')} />
                  <ConfirmModal
                    variant="outlined"
                    color="warning"
                    label={t('mgmt.constellation.resetLabel')}
                    content={t('mgmt.constellation.resetText')}
                    callback={async () => {
                      await API.constellation.reset();
                      refreshConfig();
                    }}
                  />
                  </Stack>}
                  <CosmosCheckbox formik={formik} name="Enabled" label={t('mgmt.constellation.setup.enabledCheckbox')} />
                  {config.ConstellationConfig.Enabled && !config.ConstellationConfig.SlaveMode && <>
                    {formik.values.Enabled && <>
                      <CosmosCheckbox formik={formik} name="IsRelay" label={t('mgmt.constellation.setup.relayRequests.label')} />
                      <CosmosCheckbox formik={formik} name="PrivateNode" label={t('mgmt.constellation.setup.privNode.label')} />
                      {!formik.values.PrivateNode && <>
                        <Alert severity="info"><Trans i18nKey="mgmt.constellation.setup.hostnameInfo" /></Alert>
                        <CosmosInputText formik={formik} name="ConstellationHostname" label={'Constellation '+t('global.hostname')} />
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
                      {t('global.saveAction')}
                  </LoadingButton>
                  <UploadButtons
                    accept=".yml,.yaml"
                    label={config.ConstellationConfig.SlaveMode ?
                      t('mgmt.constellation.setup.externalConfig.slaveMode.label')
                      : t('mgmt.constellation.setup.externalConfig.label')}
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
                  title: t('mgmt.constellation.setup.deviceName.label'),
                  field: (r) => <strong>{r.deviceName}</strong>,
              },
              {
                  title: t('mgmt.constellation.setup.owner.label'),
                  field: (r) => <strong>{r.nickname}</strong>,
              },
              {
                  title: t('mgmt.storage.typeTitle'),
                  field: (r) => <strong>{r.isLighthouse ? "Lighthouse" : "Client"}</strong>,
              },
              {
                  title: t('mgmt.constellation.setup.ipTitle'),
                  screenMin: 'md', 
                  field: (r) => r.ip,
              },
              {
                title: '',
                clickable: true,
                field: (r) => {
                  return <>
                    <Tooltip title="Resync Device">
                      <IconButton onClick={() => setResyncDevice([r.nickname, r.deviceName])}>
                        <SyncOutlined />
                      </IconButton>
                    </Tooltip>
                    <DeleteButton onDelete={async () => {
                      await API.constellation.block(r.nickname, r.deviceName, true);
                      refreshConfig();
                    }}></DeleteButton>
                  </>
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