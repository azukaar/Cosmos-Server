import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import AddDeviceModal from "./addDevice";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { ApiOutlined, CloudOutlined, CompassOutlined, DesktopOutlined, ExportOutlined, LaptopOutlined, MobileOutlined, SyncOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, Chip, CircularProgress, IconButton, LinearProgress, Stack, Switch, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import * as Yup from "yup";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import { isDomain } from "../../utils/indexs";
import ConfirmModal from "../../components/confirmModal";
import UploadButtons from "../../components/fileUpload";
import { useClientInfos } from "../../utils/hooks";
import { Trans, useTranslation } from 'react-i18next';
import ResyncDeviceModal from "./resyncDevice";
import VPNSalesPage from "./free";
import { autoBatchEnhancer } from "@reduxjs/toolkit";

const getDefaultConstellationHostname = (config) => {
  // if domain is set, use it
  if (isDomain(config.HTTPConfig.Hostname)) {
    return "vpn." + config.HTTPConfig.Hostname;
  } else {
    return config.HTTPConfig.Hostname;
  }
}

export const ConstellationVPN = ({ freeVersion }) => {
  const { t } = useTranslation();
  const [config, setConfig] = useState(null);
  const [users, setUsers] = useState(null);
  const [devices, setDevices] = useState(null);
  const [resynDevice, setResyncDevice] = useState(null); // [nickname, deviceName]
  const { role } = useClientInfos();
  const isAdmin = role === "2";
  const [ping, setPing] = useState(0);
  const [coStatus, setCoStatus] = React.useState(null);
  const [devicePingStatus, setDevicePingStatus] = useState({}); // {deviceName: 'loading' | 'success' | 'error'}
  const [firewallLoading, setFirewallLoading] = useState(null); // deviceName being toggled

  const refreshStatus = () => {
    API.getStatus().then((res) => {
      setCoStatus(res.data);
    });
  }

  const pingDevices = async (deviceList) => {
    if (!deviceList || deviceList.length === 0) return;

    // Initialize all devices as loading
    const initialStatus = {};
    deviceList.forEach(device => {
      initialStatus[device.deviceName] = 'loading';
    });
    setDevicePingStatus(initialStatus);

    // Ping devices 5 at a time
    const batchSize = 5;
    for (let i = 0; i < deviceList.length; i += batchSize) {
      const batch = deviceList.slice(i, i + batchSize);
      const pingPromises = batch.map(device =>
        API.constellation.pingDevice(device.deviceName)
          .then(res => {
            setDevicePingStatus(prev => ({
              ...prev,
              [device.deviceName]: res.data.reachable ? 'success' : 'error'
            }));
          })
          .catch(() => {
            setDevicePingStatus(prev => ({
              ...prev,
              [device.deviceName]: 'error'
            }));
          })
      );
      await Promise.all(pingPromises);
    }
  };

  const isFirewallBlocked = (deviceName) => {
    if (!config?.ConstellationConfig?.FirewallBlockedClients) {
      return false;
    }
    return config.ConstellationConfig.FirewallBlockedClients.includes(deviceName);
  };

  const toggleFirewallBlock = async (deviceName, isBlocked) => {
    setFirewallLoading(deviceName);

    let newConfig = { ...config };
    if (!newConfig.ConstellationConfig.FirewallBlockedClients) {
      newConfig.ConstellationConfig.FirewallBlockedClients = [];
    }

    if (isBlocked) {
      // Remove from blocked list
      newConfig.ConstellationConfig.FirewallBlockedClients =
        newConfig.ConstellationConfig.FirewallBlockedClients.filter(d => d !== deviceName);
    } else {
      // Add to blocked list
      if (!newConfig.ConstellationConfig.FirewallBlockedClients.includes(deviceName)) {
        newConfig.ConstellationConfig.FirewallBlockedClients.push(deviceName);
      }
    }

    await API.config.set(newConfig);
    setFirewallLoading(null);
  };

  let constellationEnabled = config && config.ConstellationConfig.Enabled;

  const refreshConfig = async () => {
    setPing(0);
    refreshStatus();
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    const deviceList = (await API.constellation.list()).data || [];
    setDevices(deviceList);
    if (isAdmin)
      setUsers((await API.users.list()).data || []);
    else
      setUsers([]);

    if (configAsync.data.ConstellationConfig.Enabled) {
      setPing((await API.constellation.ping()).data ? 2 : 1);
      // Ping devices after loading
      pingDevices(deviceList.filter((d) => !d.blocked));
    }
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
      <Stack spacing={2} style={{ maxWidth: "1000px", margin: freeVersion ? "auto" : 0 }}>
        <div>
          {constellationEnabled && coStatus && coStatus.ConstellationSlaveIPWarning && <Alert severity="error">
            {coStatus.ConstellationSlaveIPWarning}
          </Alert>}

          {!freeVersion && <Alert severity="info">
            <Trans i18nKey="mgmt.constellation.setupText"
              components={[<a href="https://cosmos-cloud.io/doc/61 Constellation VPN/" target="_blank"></a>, <a href="https://cosmos-cloud.io/clients" target="_blank"></a>]}
            />
          </Alert>}
          <MainCard title={t('mgmt.constellation.setupTitle')} content={config.constellationIP}>
            <Stack spacing={2}>
              {constellationEnabled && config.ConstellationConfig.SlaveMode && isAdmin && <>
                <Alert severity="info">
                  {t('mgmt.constellation.externalTextSlaveNoAdmin')}
                </Alert>
              </>}
              {constellationEnabled && config.ConstellationConfig.SlaveMode && isAdmin && <>
                <Alert severity="info">
                  {t('mgmt.constellation.externalText')}
                </Alert>
              </>}
              {!constellationEnabled && !isAdmin && <>
                <Alert severity="info">
                  {t('mgmt.constellation.setupTextNoAdmin')}
                </Alert>
              </>}
              {(isAdmin || constellationEnabled) && <Formik
                enableReinitialize
                validationSchema={Yup.object().shape({
                  DeviceName: Yup.string().required().min(3).max(32)
                    .matches(/^[a-z0-9-]+$/, t('mgmt.constellation.setup.deviceName.validationError')),
                })}
                initialValues={{
                  DeviceName: config.ConstellationConfig.ThisDeviceName || '',
                  Enabled: config.ConstellationConfig.Enabled,
                  IsRelay: config.ConstellationConfig.IsRelayNode,
                  IsExitNode: config.ConstellationConfig.IsExitNode,
                  SyncNodes: !config.ConstellationConfig.DoNotSyncNodes,
                  ConstellationHostname: (config.ConstellationConfig.ConstellationHostname && config.ConstellationConfig.ConstellationHostname != "") ? config.ConstellationConfig.ConstellationHostname :
                    getDefaultConstellationHostname(config)
                }}
                onSubmit={async (values) => {
                  const isCreating = !config.ConstellationConfig.ThisDeviceName;
                  if (isCreating) {
                    await API.constellation.create(values.DeviceName);
                    setTimeout(() => {
                      refreshConfig();
                    }, 1500);
                    return;
                  }
                  let newConfig = { ...config };
                  newConfig.ConstellationConfig.Enabled = values.Enabled;
                  newConfig.ConstellationConfig.IsRelayNode = values.IsRelay;
                  newConfig.ConstellationConfig.IsExitNode = values.IsExitNode;
                  newConfig.ConstellationConfig.ConstellationHostname = values.ConstellationHostname;
                  newConfig.ConstellationConfig.DoNotSyncNodes = !values.SyncNodes;
                  setTimeout(() => {
                    refreshConfig();
                  }, 1500);
                  return API.config.set(newConfig);
                }}
              >
                {(formik) => (
                  <form onSubmit={formik.handleSubmit}>
                    <Stack spacing={2}>
                      {isAdmin && constellationEnabled && <Stack spacing={2} direction="row">
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
                        <ApiModal 
                          callback={API.constellation.getLogs} 
                          label={t('mgmt.constellation.showLogsButton')} 
                          processContent={(logs) => {
                            const result = [];
                            
                            logs.split("\n").forEach((line, lineIndex) => {
                              if (!line.trim()) return;
                              
                              const levelMatch = line.match(/level=(\w+)/);
                              const level = levelMatch?.[1]?.toLowerCase();
                              
                              const color = {
                                error: "red",
                                warn: "orange", 
                                debug: "purple"
                              }[level] || "white";
                              
                              const elements = [];
                              let lastIndex = 0;
                              
                              line.replace(/([a-z]+=)([^\s]+)/g, (match, key, value, offset) => {
                                // Add text before this match
                                if (offset > lastIndex) {
                                  elements.push(line.slice(lastIndex, offset));
                                }
                                
                                // Style differently based on the key
                                if (key === 'msg=') {
                                  elements.push(
                                    <span key={offset}>
                                      <span style={{ fontWeight: 'bold', color: '#52a5d8ff' }}>{key}</span>
                                      <span style={{ }}>{value}</span>
                                    </span>
                                  );
                                } else {
                                  elements.push(
                                    <span key={offset}>
                                      <span style={{ color: '#81bbdfff' }}>{key}</span>
                                      {value}
                                    </span>
                                  );
                                }
                                
                                lastIndex = offset + match.length;
                                return match;
                              });
                              
                              // Add any remaining text
                              if (lastIndex < line.length) {
                                elements.push(line.slice(lastIndex));
                              }
                              
                              result.push(
                                <div key={lineIndex} style={{ color }}>
                                  ● {elements}
                                </div>
                              );
                            });
                            
                            return result;
                          }}

                        />
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

                      {constellationEnabled && <div>
                        {t('mgmt.constellation.constStatus')}: {[
                          <CircularProgress color="inherit" size={20} />,
                          <span style={{ color: "red" }}>{t('mgmt.constellation.constStatusDown')}</span>,
                          <span style={{ color: "green" }}>{t('mgmt.constellation.constStatusConnected')}</span>,
                        ][ping]}

                        <IconButton onClick={async () => {
                          setPing(0);
                          setPing((await API.constellation.ping()).data ? 2 : 1);
                        }}>
                          <SyncOutlined />
                        </IconButton>
                      </div>}

                      {!freeVersion && <>
                        <CosmosInputText
                          disabled={!isAdmin || !!config.ConstellationConfig.ThisDeviceName}
                          formik={formik}
                          name="DeviceName"
                          label={t('mgmt.constellation.setup.deviceName.label')}
                        />

                        {config.ConstellationConfig.ThisDeviceName && (
                          <CosmosCheckbox disabled={!isAdmin} formik={formik} name="Enabled" label={t('mgmt.constellation.setup.enabledCheckbox')} />
                        )}

                        {constellationEnabled && !config.ConstellationConfig.SlaveMode && <>
                          {formik.values.Enabled && <>
                            <CosmosCheckbox disabled={!isAdmin} formik={formik} name="SyncNodes" label={t('mgmt.constellation.setup.dataSync.label')} />
                            {devices.length > 0 && <Alert severity="warning">{t('mgmt.constellation.setup.deviceConnectedWarn')}</Alert>}
                            <CosmosCheckbox disabled={!isAdmin || devices.length > 0} formik={formik} name="IsRelay" label={t('mgmt.constellation.setup.relayRequests.label')} />
                            <CosmosCheckbox disabled={!isAdmin || devices.length > 0} formik={formik} name="IsExitNode" label={t('mgmt.constellation.setup.exitNode.label')} />
                            <Alert severity="info"><Trans i18nKey="mgmt.constellation.setup.hostnameInfo" /></Alert>
                            <CosmosInputText disabled={!isAdmin || devices.length > 0} formik={formik} name="ConstellationHostname" label={'Constellation ' + t('global.hostname')} />
                          </>}
                        </>}

                        {isAdmin && <><LoadingButton
                          disableElevation
                          loading={formik.isSubmitting}
                          type="submit"
                          variant="contained"
                          color="primary"
                        >
                          {config.ConstellationConfig.ThisDeviceName
                            ? t('global.saveAction')
                            : t('mgmt.constellation.setup.createConstellation')}
                        </LoadingButton>
                        </>}
                      </>}
                      {isAdmin && <><UploadButtons
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
                      /></>}
                    </Stack>
                  </form>
                )}
              </Formik>}
            </Stack>
          </MainCard>
        </div>
        {config.ConstellationConfig.Enabled && <>
          <CosmosFormDivider title={"Devices"} />

          {!config.ConstellationConfig.SlaveMode && <Stack direction="row" spacing={3} style={{ marginBottom: '20px' }}>
            <div>
              <div>{t('mgmt.constellation.deviceSeatsUsed')}: {devices ? devices.filter(d => !d.blocked).length : 0} / {coStatus ? coStatus.LicenceNumber * 10 : 0}</div>
              <LinearProgress
                style={{width: '150px'}}
                variant="determinate"
                value={(coStatus && devices) ? (devices.filter(d => !d.blocked).length / (coStatus.LicenceNumber * 10)) * 100 : 0}
                color={(coStatus && devices) ? (devices.filter(d => !d.blocked).length >= coStatus.LicenceNumber * 10 ? 'error' : 'primary') : 'primary'}
              />
            </div>

            <div>
              <div>{t('mgmt.constellation.cosmosNodeSeatsUsed')}: {devices ? devices.filter(d => !d.blocked && d.isCosmosNode).length : 0} / {coStatus ? coStatus.LicenceNodeNumber : 0}</div>
              <LinearProgress
                style={{width: '150px'}}
                variant="determinate"
                value={(coStatus && devices) ? (devices.filter(d => !d.blocked && d.isCosmosNode).length / coStatus.LicenceNodeNumber) * 100 : 0}
                color={(coStatus && devices) ? (devices.filter(d => !d.blocked && d.isCosmosNode).length >= coStatus.LicenceNodeNumber ? 'error' : 'primary') : 'primary'}
              />
            </div>
          </Stack>}

          <PrettyTableView
            data={devices.filter((d) => !d.blocked)}
            getKey={(r) => r.deviceName}
            buttons={[
              !config.ConstellationConfig.SlaveMode && (<AddDeviceModal users={users} config={config} refreshConfig={refreshConfig} devices={devices} />),
              <Button
                disableElevation
                variant="outlined"
                color="primary"
                onClick={async () => {
                  pingDevices(devices.filter((d) => !d.blocked));
                }}
              >
                {t('mgmt.constellation.setup.repingAll')}
              </Button>
            ]}
            columns={[
              {
                title: '',
                field: getIcon,
              },
              {
                title: t('mgmt.constellation.setup.deviceName.label'),
                field: (r) => {

                  const status = devicePingStatus[r.deviceName];
                  let res = "";

                  if (status === 'loading') {
                    res = <CircularProgress size={16} />;
                  } else if (status === 'success') {
                    res = "🟢";
                  } else if (status === 'error') {
                    res = "🔴";
                  }

                  return <strong>{res} {r.deviceName}</strong>;
                }
              },
              {
                title: t('mgmt.constellation.setup.owner.label'),
                field: (r) => <strong>{r.nickname}</strong>,
              },
              {
                title: t('mgmt.storage.typeTitle'),
                field: (r) => <strong>{r.isCosmosNode ? "Cosmos Node" : (r.isLighthouse ? "Lighthouse" : "Client")}</strong>,
              },
              {
                title: '',
                field: (r) => {
                  if (!r.isLighthouse) return null;
                  return (
                    <Stack direction="row" spacing={1}>
                      {r.isCosmosNode && (
                        <Tooltip title="Cosmos Node">
                          <CloudOutlined style={{ color: '#9c27b0' }} />
                        </Tooltip>
                      )}
                      {r.isRelay && (
                        <Tooltip title={t('mgmt.constellation.isRelay.label')}>
                          <ApiOutlined style={{ color: '#1976d2' }} />
                        </Tooltip>
                      )}
                      {r.isExitNode && (
                        <Tooltip title={t('mgmt.constellation.isExitNode.label')}>
                          <ExportOutlined style={{ color: '#2e7d32' }} />
                        </Tooltip>
                      )}
                    </Stack>
                  );
                },
              },
              {
                title: t('mgmt.constellation.setup.ipTitle'),
                screenMin: 'md',
                field: (r) => r.ip,
              },
              {
                title: t('mgmt.constellation.setup.firewallStatus'),
                field: (r) => {
                  const blocked = isFirewallBlocked(r.deviceName);
                  const isLoading = firewallLoading === r.deviceName;

                  if (isLoading) {
                    return <Chip
                      label={
                        <div>
                          <CircularProgress size={16} style={{ verticalAlign: "middle", marginRight: 4 }} />
                          Updating...
                        </div>
                      }
                      color="default"
                    />;
                  }

                  return blocked ? <Chip
                    label={
                      <div>
                        <Switch size="small" style={{ verticalAlign: "middle", marginRight: 4 }} />
                        Blocked
                      </div>
                    }
                    color="error"
                    onClick={() => toggleFirewallBlock(r.deviceName, blocked)}
                    style={{ cursor: 'pointer' }}
                  /> : <Chip
                    label={
                      <div>
                        <Switch
                          size="small"
                          sx={{
                            marginTop: "-3px",
                            '& .MuiSwitch-switchBase.Mui-checked': {
                              color: "white",
                            },
                            '& .MuiSwitch-switchBase.Mui-checked + .MuiSwitch-track': {
                              backgroundColor: "white",
                            },
                          }}
                          checked
                        />
                        Allowed
                      </div>
                    }
                    color="success"
                    onClick={() => toggleFirewallBlock(r.deviceName, blocked)}
                    style={{ cursor: 'pointer' }}
                  />;
                },
              },
              {
                title: '',
                clickable: true,
                field: (r) => {
                  return !config.ConstellationConfig.SlaveMode ? <>
                    <Tooltip title="Resync Device">
                      <IconButton onClick={() => setResyncDevice([r.nickname, r.deviceName])}>
                        <SyncOutlined />
                      </IconButton>
                    </Tooltip>
                    <DeleteButton onDelete={async () => {
                      await API.constellation.block(r.nickname, r.deviceName, true);
                      refreshConfig();
                    }}></DeleteButton>
                  </> : null
                }
              }
            ]}
          />
        </>}
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}

    {freeVersion && config && !constellationEnabled && <VPNSalesPage />}
  </>
};