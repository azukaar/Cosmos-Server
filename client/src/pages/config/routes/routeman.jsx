import * as React from 'react';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Grid,
  Stack,
  FormHelperText,
  TextField,
  MenuItem,
} from '@mui/material';
import RestartModal from '../users/restart';
import { CosmosCheckbox, CosmosCollapse, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../users/formShortcuts';
import { CosmosContainerPicker } from '../users/containerPicker';
import { snackit } from '../../../api/wrap';
import { IsRouteSocketProxy, ValidateRouteSchema, getHostnameFromName, sanitizeRoute } from '../../../utils/routes';
import { isDomain } from '../../../utils/indexs';
import { useTranslation } from 'react-i18next';
import { FilePickerButton } from '../../../components/filePicker';

const Hide = ({ children, h }) => {
  return h ? <div style={{ display: 'none' }}>
    {children}
  </div> : <>{children}</>
}

const debounce = (func, wait) => {
  let timeout;
  return function (...args) {
    const context = this;
    clearTimeout(timeout);
    timeout = setTimeout(() => func.apply(context, args), wait);
  };
};

const checkHost = debounce((host, setHostError) => {
  if(isDomain(host)) {
    API.checkHost(host).then((data) => {
      setHostError(null)
    }).catch((err) => {
      setHostError(err.message)
    });
  } else {
    setHostError(null);
  }
}, 500);

const RouteManagement = ({ routeConfig, routeNames, config, TargetContainer, noControls = false, lockTarget = false, title, setRouteConfig, submitButton = false, newRoute }) => {
  const { t } = useTranslation();
  const [openModal, setOpenModal] = React.useState(false);
  const [hostError, setHostError] = React.useState(null);

  const [tunnels, setTunnels] = React.useState([]);

  React.useEffect(() => {
    if(routeConfig && routeConfig.Host) {
      checkHost(routeConfig.Host, setHostError);
    }

    (async () => {
      setTunnels((await API.constellation.list()).data.filter(
        (device) => device.isLighthouse && device.isLoadBalancer
      ) || []);
    })();

  }, [])
 
  return <div style={{ maxWidth: '1000px', width: '100%', margin: 'auto', position: 'relative' }}>
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} config={config} />
    
    {routeConfig && <>
      <Formik
        initialValues={{
          Name: routeConfig.Name,
          Description: routeConfig.Description,
          Mode: routeConfig.Mode || "SERVAPP",
          Target: routeConfig.Target || 'http://',
          UseHost: routeConfig.UseHost,
          Host: routeConfig.Host,
          AcceptInsecureHTTPSTarget: routeConfig.AcceptInsecureHTTPSTarget === true,
          UseH2C: routeConfig.UseH2C === true,
          UsePathPrefix: routeConfig.UsePathPrefix,
          PathPrefix: routeConfig.PathPrefix,
          StripPathPrefix: routeConfig.StripPathPrefix,
          AuthEnabled: routeConfig.AuthEnabled,
          HideFromDashboard: routeConfig.HideFromDashboard,
          _SmartShield_Enabled: (routeConfig.SmartShield ? routeConfig.SmartShield.Enabled : false),
          RestrictToConstellation: routeConfig.RestrictToConstellation === true,
          SkipURLClean: routeConfig.SkipURLClean === true,
          OverwriteHostHeader: routeConfig.OverwriteHostHeader,
          Tunnel: routeConfig.Tunnel,
          TunneledHost: routeConfig.TunneledHost,
          WhitelistInboundIPs: routeConfig.WhitelistInboundIPs && routeConfig.WhitelistInboundIPs.join(', '),
        }}
        validationSchema={ValidateRouteSchema}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          if(!submitButton) {
            return false;
          } else {
            let commaSepIps = values.WhitelistInboundIPs;
            
            if(commaSepIps && commaSepIps.length) {
              values.WhitelistInboundIPs = commaSepIps.split(',').map((ip) => ip.trim());
            } else {
              values.WhitelistInboundIPs = [];
            }

            let fullValues = {
              ...routeConfig,
              ...values,
            }

            fullValues = sanitizeRoute(fullValues);

            if(IsRouteSocketProxy(fullValues)) {
              if(!fullValues.Host.includes(':')) {
                fullValues.Host = `0.0.0.0:${fullValues.Host}`;
              }
              fullValues.UseHost = true;
            }

            let op;
            if(newRoute) {
              op = API.config.newRoute(routeConfig.Name, fullValues)
            } else {
              op = API.config.replaceRoute(routeConfig.Name, fullValues)
            }
            
            op.then((res) => {
              if (res.status == "OK") {
                setStatus({ success: true });
                snackit(t('mgmt.urls.edit.updateSuccess'), 'success')
                setSubmitting(false);
                setOpenModal(true);
              } else {
                setStatus({ success: false });
                setErrors({ submit: res.status });
                setSubmitting(false);
              }
            });
          }
        }}
        validate={(values) => {
          let commaSepIps = values.WhitelistInboundIPs;
          
          let fullValues = {
            ...routeConfig,
            ...values,
          }

          if(commaSepIps && commaSepIps.length) {
            fullValues.WhitelistInboundIPs = commaSepIps.split(',').map((ip) => ip.trim());
          } else {
            fullValues.WhitelistInboundIPs = [];
          }

          // check name is unique
          if (newRoute && routeNames.includes(fullValues.Name)) {
            return { Name: t('navigation.market.sources.nameNotUniqueValidation')  }
          }

          if(IsRouteSocketProxy(fullValues)) {
            if(!fullValues.Host.includes(':')) {
              fullValues.Host = `0.0.0.0:${fullValues.Host}`;
            }
            fullValues.UseHost = true;
          }

          setRouteConfig && debounce(() => setRouteConfig(fullValues), 500)();
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={2}>
              <MainCard name={routeConfig.Name} title={
                noControls ? t('mgmt.urls.edit.newUrlTitle') :
                  <div>{title || routeConfig.Name}</div>
              }>
                <Grid container spacing={2}>
                  {formik.errors.submit && (
                    <Grid item xs={12}>
                      <FormHelperText error>{formik.errors.submit}</FormHelperText>
                    </Grid>
                  )}

                  <CosmosInputText
                    name="Name"
                    label={t('global.nameTitle')}
                    placeholder={t('global.nameTitle')}
                    formik={formik}
                  />

                  <CosmosInputText
                    name="Description"
                    label={t('global.description')}
                    placeholder={t('global.description')}
                    formik={formik}
                  />

                  <Hide h={lockTarget}>
                    <CosmosFormDivider title={t('mgmt.urls.edit.targetTypeTitle')} />
                    <Grid item xs={12}>
                      <Alert color='info'>{t('mgmt.urls.edit.targetTypeInfo')}</Alert>
                    </Grid>

                    <CosmosSelect
                      name="Mode"
                      label={t('mgmt.urls.edit.targetType.modeSelection.modeLabel')}
                      formik={formik}
                      disabled={lockTarget}
                      options={[
                        ["SERVAPP", t('mgmt.urls.edit.targetType.modeSelection.servAppChoice')],
                        ["PROXY", t('mgmt.urls.edit.targetType.modeSelection.proxyChoice')],
                        ["STATIC", t('mgmt.urls.edit.targetType.modeSelection.staticChoice')],
                        ["SPA", t('mgmt.urls.edit.targetType.modeSelection.spaChoice')],
                        ["REDIRECT", t('mgmt.urls.edit.targetType.modeSelection.redirectChoice')]
                      ]}
                    />
                  </Hide>
                  <CosmosFormDivider title={t('mgmt.urls.edit.targetSettingsTitle')} />

                  {
                    (formik.values.Mode === "SERVAPP") ?
                      <CosmosContainerPicker
                        formik={formik}
                        lockTarget={lockTarget}
                        TargetContainer={TargetContainer}
                        onTargetChange={() => {
                          setRouteConfig && setRouteConfig(formik.values);
                        }}
                      />
                      : <Grid item xs={12}><Stack direction="row" spacing={2} fullWidth alignItems="flex-end">
                      {(formik.values.Mode == "STATIC" || formik.values.Mode == "SPA") && (
                        <FilePickerButton onPick={(path) => {
                          if(path)
                            formik.setFieldValue('Target', path);
                        }} size="150%" select="folder" />)}

                      <CosmosInputText
                        style={{  }}
                        name="Target"
                        label={formik.values.Mode == "PROXY" ? t('mgmt.urls.edit.targetSettings.targetUrlInput.targetUrlLabel') : t('mgmt.urls.edit.targetFolderPathInput.targetFolderPathLabel')}
                        placeholder={formik.values.Mode == "PROXY" ? "http://localhost:8080" : "/path/to/my/app"}
                        formik={formik}
                      /></Stack></Grid>
                  }

                  
                  {!(formik.values.Target.startsWith('http://') || formik.values.Target.startsWith('https://')) && <Grid item xs={12}>
                    <Alert severity='warning'>{t('mgmt.urls.edit.targetSettings.targetUrlProtocolWarning')}</Alert>
                  </Grid>}

                  {formik.values.Target.startsWith('https://') && <CosmosCheckbox
                    name="AcceptInsecureHTTPSTarget"
                    label={t('mgmt.urls.edit.insecureHttpsCheckbox.insecureHttpsLabel')}
                    formik={formik}
                  />}

                  <CosmosFormDivider title={t('global.source')} />
                  
                  {IsRouteSocketProxy(formik.values) && <>
                    <Grid item xs={12}>
                      <Alert color='info'>{t('mgmt.urls.edit.sourceInfoPort2')}</Alert>
                    </Grid>
                    <CosmosInputText
                      name="Host"
                      label={t('mgmt.servapps.networks.list.host-port2')}
                      placeholder={1234}
                      formik={formik}
                      onChange={(e) => {
                      }}
                    />
                    </>}

                  {!IsRouteSocketProxy(formik.values) && <>
                  <Grid item xs={12}>
                    <Alert color='info'>{t('mgmt.urls.edit.sourceInfo')}</Alert>
                  </Grid>

                  <CosmosCheckbox
                    name="UseHost"
                    label={t('mgmt.urls.edit.useHostCheckbox.useHostLabel')}
                    formik={formik}
                  />

                  {formik.values.UseHost && (<><CosmosInputText
                    name="Host"
                    label={t('mgmt.servapps.networks.list.host')}
                    placeholder={t('mgmt.servapps.networks.list.host')}
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                    onChange={(e) => {
                      checkHost(e.target.value, setHostError)
                    }}
                  />
                  {hostError && <Grid item xs={12}>
                    <Alert color='warning'>{hostError}</Alert>
                  </Grid>}
                  </>
                  )}

                  <CosmosCheckbox
                    name="UsePathPrefix"
                    label={t('mgmt.urls.edit.usePathPrefixCheckbox.usePathPrefixLabel')}
                    formik={formik}
                  />

                  {formik.values.UsePathPrefix && <Grid item xs={12}>
                    <Alert color='warning'>{t('mgmt.urls.edit.sourceWarningPathPrefix')}</Alert>
                  </Grid>}

                  {formik.values.UsePathPrefix && <CosmosInputText
                    name="PathPrefix"
                    label={t('mgmt.urls.edit.pathPrefixInputx.pathPrefixLabel')}
                    placeholder={t('mgmt.urls.edit.pathPrefixInputx.pathPrefixPlaceholder')}
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}

                  {formik.values.UsePathPrefix && <CosmosCheckbox
                    name="StripPathPrefix"
                    label={t('mgmt.urls.edit.stripPathCheckbox.stripPathLabel')}
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}
                  </>}
                  
                  <CosmosFormDivider title={t('mgmt.urls.edit.basicSecurityTitle')} />
                  
                  {!IsRouteSocketProxy(formik.values) && <>
                  <CosmosCheckbox
                    name="AuthEnabled"
                    label={t('mgmt.urls.edit.basicSecurity.authEnabledCheckbox.authEnabledLabel')}
                    formik={formik}
                  /></>}
                  
                  <CosmosCheckbox
                    name="_SmartShield_Enabled"
                    label={t('mgmt.urls.edit.basicSecurity.smartShieldEnabledCheckbox.smartShieldEnabledLabel')}
                    formik={formik}
                  />
                  
                  <CosmosCheckbox
                    name="RestrictToConstellation"
                    label={t('mgmt.urls.edit.basicSecurity.restrictToConstellationCheckbox.restrictToConstellationLabel')}
                    formik={formik}
                  />

                  {(formik.values.Tunnel || tunnels.length > 0) && (<>
                    <CosmosSelect
                      name="Tunnel"
                      label={t('mgmt.urls.edit.tunnelSelection.tunnelLabel')}
                      formik={formik}
                      options={[
                        ['',  t('mgmt.urls.edit.tunnelSelection.noNode')],
                        ['_ANY_', t('mgmt.urls.edit.tunnelSelection.anyNode')],
                        ...tunnels.map((t) => [t.deviceName, `${t.deviceName} (${t.publicHostname})`])
                      ]}
                    />

                    {formik.values.Tunnel && <CosmosInputText
                      name="TunneledHost"
                      label={t('mgmt.urls.edit.tunneledHostInput.tunneledHostLabel2')}
                      placeholder="other-host.com"
                      formik={formik}
                    />}
                  </>)}

                  <CosmosCollapse title={t('mgmt.urls.edit.advancedSettingsTitle')}>
                    <Stack spacing={2}>
                      <CosmosCheckbox
                        name="HideFromDashboard"
                        label={t('mgmt.urls.edit.advancedSettings.hideFromDashboardCheckbox.hideFromDashboardLabel')}
                        formik={formik}
                      />

                      {!IsRouteSocketProxy(formik.values) && <>
                      <CosmosFormDivider />
                      <Alert severity='info'>{t('mgmt.urls.edit.advancedSettings.advancedSettingsInfo')}</Alert>
                      <CosmosInputText
                        name="OverwriteHostHeader"
                        label={t('mgmt.urls.edit.advancedSettings.overwriteHostHeaderInput.overwriteHostHeaderLabel')}
                        placeholder={t('mgmt.urls.edit.advancedSettings.overwriteHostHeaderInput.overwriteHostHeaderPlaceholder')}
                        formik={formik}
                      />

                      <CosmosCheckbox
                        name="SkipURLClean"
                        label={t('mgmt.urls.edit.advancedSettings.skipURLCleanCheckbox.skipURLCleanLabel')}
                        formik={formik}
                      />

                      {(formik.values.Mode === "SERVAPP" || formik.values.Mode === "PROXY") && <CosmosCheckbox
                        name="UseH2C"
                        label={t('mgmt.urls.edit.useH2CCheckbox.useH2CLabel')}
                        formik={formik}
                      />}
                      </>}

                      <Alert severity='warning'>
                        {t('mgmt.urls.edit.advancedSettings.filterIpWarning')}
                      </Alert>

                      <CosmosInputText
                        name="WhitelistInboundIPs"
                        label={t('mgmt.urls.edit.advancedSettings.whitelistInboundIpInput.whitelistInboundIpLabel')}
                        placeholder={t('mgmt.urls.edit.advancedSettings.whitelistInboundIpInput.whitelistInboundIpPlaceholder')}
                        formik={formik}
                      />
                    </Stack>
                  </CosmosCollapse>
                </Grid>
              </MainCard>
              {submitButton && <MainCard ><Button
                fullWidth
                disableElevation
                size="large"
                type="submit"
                variant="contained"
                color="primary"
              >
                {t('global.saveAction')}
              </Button></MainCard>}
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteManagement;