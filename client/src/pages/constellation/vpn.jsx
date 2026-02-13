import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import AddDeviceModal from "./addDevice";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { ApiOutlined, CloudOutlined, CloudServerOutlined, CompassOutlined, DesktopOutlined, ExportOutlined, LaptopOutlined, MobileOutlined, NodeIndexOutlined, QuestionCircleOutlined, SyncOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Box, Button, Chip, CircularProgress, IconButton, LinearProgress, Skeleton, Stack, Switch, Tooltip, Typography } from "@mui/material";
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
import StatusDot from "../../components/statusDot";

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
  const [currentDeviceName, setCurrentDeviceName] = useState('');
  const [enableLoading, setEnableLoading] = useState(false);

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
    const listRes = await API.constellation.list();
    const deviceList = listRes.data || [];
    setDevices(deviceList);
    setCurrentDeviceName(listRes.currentDeviceName || '');
    if (isAdmin)
      setUsers((await API.users.list()).data || []);
    else
      setUsers([]);

    if (configAsync.data.ConstellationConfig.Enabled) {
      // Initialize device ping status to loading immediately so spinners show while waiting
      const nonBlockedDevices = deviceList.filter((d) => !d.blocked);
      const initialStatus = {};
      nonBlockedDevices.forEach(device => {
        initialStatus[device.deviceName] = 'loading';
      });
      setDevicePingStatus(initialStatus);

      setPing((await API.constellation.ping()).data ? 2 : 1);
      // Ping devices after loading
      pingDevices(nonBlockedDevices);
    }
  };

  useEffect(() => {
    refreshConfig();
  }, []);

  const getIcon = (r) => {
    const name = r.deviceName.toLowerCase();
    let icon, label;
    if (name.includes("mobile") || name.includes("phone")) {
      icon = <MobileOutlined />; label = "Mobile";
    } else if (name.includes("laptop") || name.includes("computer")) {
      icon = <LaptopOutlined />; label = "Laptop";
    } else if (name.includes("desktop")) {
      icon = <DesktopOutlined />; label = "Desktop";
    } else if (name.includes("tablet")) {
      icon = <TabletOutlined />; label = "Tablet";
    } else if (name.includes("lighthouse") || name.includes("server")) {
      icon = <CompassOutlined />; label = "Server";
    } else {
      icon = <CloudOutlined />; label = "Device";
    }
    return <Tooltip title={label}>{icon}</Tooltip>;
  }

  const currentDevice = devices && devices.find(d => d.deviceName === currentDeviceName);

  return <>
    {(devices && config && users) ? <>
      {resynDevice &&
        <ResyncDeviceModal nickname={resynDevice[0]} deviceName={resynDevice[1]} OnClose={
          () => setResyncDevice(null)
        } />
      }
      <Stack spacing={2} style={{ maxWidth: "1000px", margin: "auto" }}>

        {config.ConstellationConfig.ThisDeviceName && <MainCard>
          <Stack spacing={1.5}>
            <Stack direction="row" alignItems="center" justifyContent="space-between">
              <Stack direction="row" alignItems="center" spacing={1.5}>
                <Box sx={{
                  width: 40, height: 40, borderRadius: '50%',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  backgroundColor: !constellationEnabled ? 'action.disabledBackground' : ping === 2 ? 'success.main' : ping === 1 ? 'error.main' : 'action.disabledBackground',
                  color: 'white', fontSize: 20,
                  transition: 'background-color 0.3s',
                }}>
                  <CompassOutlined />
                </Box>
                <Stack spacing={0}>
                  <Typography variant="h6">{config.ConstellationConfig.ThisDeviceName}</Typography>
                  <Stack direction="row" spacing={2}>
                    {currentDevice && <Typography variant="body2" color="textSecondary"><strong>{t('mgmt.constellation.privateIP')}:</strong> {currentDevice.ip}</Typography>}
                    {config.ConstellationConfig.ConstellationHostname && <Typography variant="body2" color="textSecondary"><strong>{t('mgmt.constellation.publicHostname')}:</strong> {config.ConstellationConfig.ConstellationHostname}</Typography>}
                  </Stack>
                </Stack>
              </Stack>

              {currentDevice && <>
                <Stack direction="row" spacing={1} alignItems="center" sx={{ color: 'text.secondary' }}>
                  {currentDevice.isLighthouse
                    ? <Tooltip title={t('mgmt.constellation.publicLighthouseTooltip')}><CompassOutlined style={{ fontSize: 16 }} /></Tooltip>
                    : <Tooltip title={t('mgmt.constellation.privateServerTooltip')}><DesktopOutlined style={{ fontSize: 16, opacity: 0.5 }} /></Tooltip>
                  }
                  {currentDevice.isRelay && <Tooltip title={t('mgmt.constellation.isRelay.label')}><ApiOutlined style={{ fontSize: 16, color: '#1976d2' }} /></Tooltip>}
                  {currentDevice.isExitNode && <Tooltip title={t('mgmt.constellation.isExitNode.label')}><ExportOutlined style={{ fontSize: 16, color: '#2e7d32' }} /></Tooltip>}
                  {currentDevice.isLoadBalancer && <Tooltip title={t('mgmt.constellation.isLoadBalancer.label')}><DesktopOutlined style={{ fontSize: 16, color: '#ff9800' }} /></Tooltip>}

                  {currentDevice.cosmosNode === 2
                      ? <Tooltip title={t('mgmt.constellation.cosmosNodeManagerTooltip')}><span style={{ display: 'flex', alignItems: 'center',
                  gap: 4 }}><CloudServerOutlined style={{ fontSize: 16, color: '#9c27b0' }} /> <Typography
                  variant="body2">{t('mgmt.constellation.cosmosNodeManager')}</Typography></span></Tooltip>
                      : currentDevice.cosmosNode === 1
                      ? <Tooltip title={t('mgmt.constellation.cosmosNodeAgentTooltip')}><span style={{ display: 'flex', alignItems: 'center',
                  gap: 4 }}><NodeIndexOutlined style={{ fontSize: 16, color: '#e65100' }} /> <Typography
                  variant="body2">{t('mgmt.constellation.cosmosNodeAgent')}</Typography></span></Tooltip>
                      : currentDevice.isLighthouse
                      ? <Tooltip title={t('mgmt.constellation.publicLighthouseTooltip')}><span style={{ display: 'flex', alignItems: 'center',
                  gap: 4 }}><Typography
                  variant="body2">{t('mgmt.constellation.publicLighthouse')}</Typography></span></Tooltip>
                      : <Tooltip title={t('mgmt.constellation.privateServerTooltip')}><span style={{ display: 'flex', alignItems: 'center',
                  gap: 4 }}><Typography
                  variant="body2">{t('mgmt.constellation.privateServer')}</Typography></span></Tooltip>
                  }          
                </Stack>
              </>}
            </Stack>

            <Stack direction="row" alignItems="center" justifyContent="space-between"
              sx={{ px: 2, py: 1, borderRadius: 1, backgroundColor: 'action.hover' }}
            >
              <Stack direction="row" spacing={1} alignItems="center">
                <Typography variant="body2">{constellationEnabled ? t('mgmt.constellation.setup.enabledCheckbox') : t('mgmt.constellation.disabled')}</Typography>
                {constellationEnabled && <>
                  <Typography variant="body2" color="textSecondary">-</Typography>
                  <Typography variant="body2" color={ping === 2 ? 'success.main' : ping === 1 ? 'error.main' : 'textSecondary'}>
                    {[
                      <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4 }}><CircularProgress color="inherit" size={12} /> {t('mgmt.constellation.constStatus')}</span>,
                      t('mgmt.constellation.constStatusDown'),
                      t('mgmt.constellation.constStatusConnected'),
                    ][ping]}
                  </Typography>
                  <IconButton size="small" onClick={async () => {
                    setPing(0);
                    setPing((await API.constellation.ping()).data ? 2 : 1);
                  }}>
                    <SyncOutlined style={{ fontSize: 14 }} />
                  </IconButton>
                </>}
              </Stack>
              {enableLoading
                ? <CircularProgress size={24} sx={{ mr: 1 }} />
                : <Switch
                  disabled={!isAdmin}
                  checked={!!constellationEnabled}
                  onChange={async (e) => {
                    setEnableLoading(true);
                    let newConfig = { ...config };
                    newConfig.ConstellationConfig.Enabled = e.target.checked;
                    await API.config.set(newConfig);
                    setTimeout(() => {
                      refreshConfig().then(() => setEnableLoading(false));
                    }, 1500);
                  }}
                  color="success"
                />
              }
            </Stack>
          </Stack>

          {isAdmin && constellationEnabled && <>
            <Box sx={{ borderTop: '1px solid', borderColor: 'divider', mt: 2, pt: 2 }}>
              <Stack spacing={1} direction="row" flexWrap="wrap">
                <Button
                  disableElevation
                  size="small"
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
                        if (offset > lastIndex) {
                          elements.push(line.slice(lastIndex, offset));
                        }

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
              </Stack>
            </Box>
          </>}
        </MainCard>}

        {(isAdmin || !config.ConstellationConfig.ThisDeviceName) && <div>
          <MainCard title={config.ConstellationConfig.ThisDeviceName ? t('mgmt.constellation.settingsTitle') : t('mgmt.constellation.setupTitle')} content={config.constellationIP}>
            <Stack spacing={2}>
              {constellationEnabled && !config.ConstellationConfig.ThisDeviceName && <div>
                <Alert severity="error">
                  {t('mgmt.constellation.corruptedConstellation')} <br /><br />
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
                </Alert>
              </div>}
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
                  DeviceName: config.ConstellationConfig.ThisDeviceName || 'cosmos-0',
                  ConstellationHostname: config.ConstellationConfig.ConstellationHostname || '',
                  IPRange: config.ConstellationConfig.IPRange || '192.168.201.0/24',
                  IsLighthouse: currentDevice ? currentDevice.isLighthouse : true,
                  IsRelay: currentDevice ? currentDevice.isRelay : false,
                  IsExitNode: currentDevice ? currentDevice.isExitNode : false,
                  IsLoadBalancer: currentDevice ? currentDevice.isLoadBalancer : false,
                }}
                onSubmit={async (values) => {
                  const isCreating = !config.ConstellationConfig.ThisDeviceName;
                  if (isCreating) {
                    await API.constellation.create(values.DeviceName, values.IsLighthouse, values.ConstellationHostname, values.IPRange);
                    setTimeout(() => {
                      refreshConfig();
                    }, 1500);
                    return;
                  }

                  await API.constellation.editDevice({
                    isLighthouse: values.IsLighthouse,
                    isRelay: values.IsRelay,
                    isExitNode: values.IsExitNode,
                    isLoadBalancer: values.IsLoadBalancer,
                  });

                  setTimeout(() => {
                    refreshConfig();
                  }, 1500);
                }}
              >
                {(formik) => (
                  <form onSubmit={formik.handleSubmit}>
                    <Stack spacing={2}>
                      {!freeVersion && !config.ConstellationConfig.ThisDeviceName && <CosmosInputText
                        disabled={!isAdmin}
                        formik={formik}
                        name="DeviceName"
                        label={t('mgmt.constellation.setup.deviceName.label')}
                      />}

                      {!freeVersion && !config.ConstellationConfig.ThisDeviceName && <CosmosInputText disabled={!isAdmin} formik={formik} name="ConstellationHostname" label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                        <span>{'Constellation ' + t('global.hostname')}</span>
                        <Tooltip title={<Trans i18nKey="mgmt.constellation.setup.hostnameInfo" />}>
                          <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                        </Tooltip>
                      </Stack>} />}

                      {!freeVersion && !config.ConstellationConfig.ThisDeviceName && <CosmosInputText disabled={!isAdmin} formik={formik} name="IPRange" label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                        <span>{t('mgmt.constellation.setup.ipRange.label')}</span>
                        <Tooltip title={t('mgmt.constellation.setup.ipRange.tooltip')}>
                          <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                        </Tooltip>
                      </Stack>} />}

                      {!freeVersion && !config.ConstellationConfig.ThisDeviceName && <CosmosCheckbox disabled={!isAdmin} formik={formik} name="IsLighthouse" label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                        <span>{t('mgmt.constellation.setup.isLighthouse.label')}</span>
                        <Tooltip title={t('mgmt.constellation.setup.isLighthouse.tooltip')}>
                          <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                        </Tooltip>
                      </Stack>} />}
                      {constellationEnabled && formik.values.IsLighthouse && <>
                        <CosmosCheckbox disabled={!isAdmin} formik={formik} name="IsRelay" label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                          <span>{t('mgmt.constellation.setup.relayRequests.label')}</span>
                          <Tooltip title={t('mgmt.constellation.setup.relayRequests.tooltip')}>
                            <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                          </Tooltip>
                        </Stack>} />
                        <CosmosCheckbox disabled={!isAdmin} formik={formik} name="IsExitNode" label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                          <span>{t('mgmt.constellation.setup.exitNode.label')}</span>
                          <Tooltip title={t('mgmt.constellation.setup.exitNode.tooltip')}>
                            <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                          </Tooltip>
                        </Stack>} />
                        <CosmosCheckbox disabled={!isAdmin} formik={formik} name="IsLoadBalancer" label={<Stack direction="row" spacing={0.5} alignItems="center" component="span">
                          <span>{t('mgmt.constellation.setup.loadBalancer.label')}</span>
                          <Tooltip title={t('mgmt.constellation.setup.loadBalancer.tooltip')}>
                            <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                          </Tooltip>
                        </Stack>} />
                      </>}

                      {isAdmin && (!freeVersion || config.ConstellationConfig.ThisDeviceName) && <><LoadingButton
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

                      {isAdmin && <><UploadButtons
                        accept=".yml,.yaml"
                        label={t('mgmt.constellation.setup.externalConfig.label')}
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
        </div>}
        {config.ConstellationConfig.Enabled && (() => {
          const managers = devices ? devices.filter(d => !d.blocked && d.cosmosNode === 2).length : 0;
          const agents = devices ? devices.filter(d => !d.blocked && d.cosmosNode === 1).length : 0;
          const totalNodes = managers + agents;
          const limit = coStatus ? coStatus.LicenceNodeNumber : 1;
          let canCreateManager = true;
          let canCreateAgent = true;
          if (totalNodes >= limit) {
            canCreateAgent = totalNodes === limit;
            canCreateManager = totalNodes === limit && agents >= 1;
          }

          return <>
          <CosmosFormDivider title={t('mgmt.constellation.devices')} />

          <Stack direction="row" spacing={3} style={{ marginBottom: '20px' }}>
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
              <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                {t('mgmt.constellation.cosmosNodeSeatsUsed')}: {devices ? devices.filter(d => !d.blocked && d.cosmosNode > 0).length : 0} / {coStatus ? coStatus.LicenceNodeNumber+1 : 0}
                <Tooltip title={t('mgmt.constellation.cosmosNodeSeatsTooltip')}>
                  <QuestionCircleOutlined style={{ fontSize: 14, cursor: 'help', opacity: 0.6 }} />
                </Tooltip>
              </div>
              <LinearProgress
                style={{width: '150px'}}
                variant="determinate"
                value={(coStatus && devices) ? (devices.filter(d => !d.blocked && d.cosmosNode > 0).length / (coStatus.LicenceNodeNumber+1)) * 100 : 0}
                color={(coStatus && devices) ? (devices.filter(d => !d.blocked && d.cosmosNode > 0).length >= coStatus.LicenceNodeNumber+1 ? 'error' : 'primary') : 'primary'}
              />
            </div>
          </Stack>

          <PrettyTableView
            data={devices.filter((d) => !d.blocked)}
            getKey={(r) => r.deviceName}
            buttons={[
              (<AddDeviceModal users={users} config={config} refreshConfig={refreshConfig} devices={devices} canCreateManager={canCreateManager} canCreateAgent={canCreateAgent} />),
              <Button
                disableElevation
                variant="outlined"
                color="primary"
                onClick={async () => {
                  await refreshConfig();
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
                title: t('mgmt.constellation.setup.device.label'),
                field: (r) => {
                  const status = devicePingStatus[r.deviceName];
                  let res = "";

                  if (status === 'loading') {
                    res = <CircularProgress size={16} />;
                  } else if (status === 'success') {
                    res = <StatusDot status="success" />;
                  } else if (status === 'error') {
                    res = <StatusDot status="error" />;
                  }

                  return <div style={{display: "flex", alignItems: "center", gap: "8px"}}>
                    {res}
                    <div>
                      <div><strong>{r.deviceName}</strong></div>
                      <div style={{opacity: 0.8}}>{r.ip}</div>
                    </div>
                  </div>
                }
              },
              {
                title: t('mgmt.constellation.setup.owner.label'),
                field: (r) => <strong>{r.nickname}</strong>,
              },
              {
                title: t('mgmt.storage.typeTitle'),
                field: (r) => <strong>{r.cosmosNode === 2 ? t('mgmt.constellation.cosmosNodeManager') : r.cosmosNode === 1 ? t('mgmt.constellation.cosmosNodeAgent') : (r.isLighthouse ? t('mgmt.constellation.lighthouse') : t('mgmt.constellation.client'))}</strong>,
              },
              {
                title: '',
                field: (r) => {
                  const icons = [];
                  if (r.isLighthouse) icons.push(
                    <Tooltip key="lh" title={t('mgmt.constellation.publicLighthouse')}><CompassOutlined style={{ fontSize: 14 }} /></Tooltip>
                  );
                  if (r.cosmosNode === 2) icons.push(
                    <Tooltip key="mgr" title={t('mgmt.constellation.cosmosNodeManager')}><CloudServerOutlined style={{ fontSize: 14, color: '#9c27b0' }} /></Tooltip>
                  );
                  if (r.cosmosNode === 1) icons.push(
                    <Tooltip key="agt" title={t('mgmt.constellation.cosmosNodeAgent')}><NodeIndexOutlined style={{ fontSize: 14, color: '#e65100' }} /></Tooltip>
                  );
                  if (r.isRelay) icons.push(
                    <Tooltip key="relay" title={t('mgmt.constellation.isRelay.label')}><ApiOutlined style={{ fontSize: 14, color: '#1976d2' }} /></Tooltip>
                  );
                  if (r.isExitNode) icons.push(
                    <Tooltip key="exit" title={t('mgmt.constellation.isExitNode.label')}><ExportOutlined style={{ fontSize: 14, color: '#2e7d32' }} /></Tooltip>
                  );
                  if (r.isLoadBalancer) icons.push(
                    <Tooltip key="lb" title={t('mgmt.constellation.isLoadBalancer.label')}><DesktopOutlined style={{ fontSize: 14, color: '#ff9800' }} /></Tooltip>
                  );
                  if (icons.length === 0) return null;
                  return (
                    <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4, maxWidth: 60, justifyContent: 'center' }}>
                      {icons}
                    </div>
                  );
                },
              },
              {
                title: t('mgmt.constellation.setup.publicIpTitle'),
                screenMin: 'md',
                field: (r) => r.publicHostname,
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
                          {t('mgmt.constellation.updating')}
                        </div>
                      }
                      color="default"
                    />;
                  }

                  return blocked ? <Chip
                    label={
                      <div>
                        <Switch size="small" style={{ verticalAlign: "middle", marginRight: 4 }} />
                        {t('mgmt.constellation.blocked')}
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
                        {t('mgmt.constellation.allowed')}
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
                  return <>
                    <Tooltip title={t('mgmt.constellation.resyncDevice')}>
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
        </>;
        })()}
      </Stack>
    </> : <Stack spacing={2} style={{ maxWidth: "1000px", margin: "auto" }}>
      <Skeleton variant="rectangular" height={400} sx={{ borderRadius: 1 }} />
      <Skeleton variant="rectangular" height={500} sx={{ borderRadius: 1 }} />
    </Stack>}

    {freeVersion && config && !constellationEnabled && <VPNSalesPage />}
  </>
};