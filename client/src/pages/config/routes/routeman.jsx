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
import { ValidateRouteSchema, getHostnameFromName, sanitizeRoute } from '../../../utils/routes';
import { isDomain } from '../../../utils/indexs';

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
}, 500)

const RouteManagement = ({ routeConfig, routeNames, config, TargetContainer, noControls = false, lockTarget = false, title, setRouteConfig, submitButton = false, newRoute }) => {
  const [openModal, setOpenModal] = React.useState(false);
  const [hostError, setHostError] = React.useState(null);

  const [tunnels, setTunnels] = React.useState([]);

  React.useEffect(() => {
    if(routeConfig && routeConfig.Host) {
      checkHost(routeConfig.Host, setHostError);
    }

    (async () => {
      setTunnels((await API.constellation.list()).data.filter(
        (device) => device.isLighthouse
      ) || []);
    })();

  }, [])
 
  return <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} config={config} />
    
    {routeConfig && <>
      <Formik
        initialValues={{
          Name: routeConfig.Name,
          Description: routeConfig.Description,
          Mode: routeConfig.Mode || "SERVAPP",
          Target: routeConfig.Target,
          UseHost: routeConfig.UseHost,
          Host: routeConfig.Host,
          AcceptInsecureHTTPSTarget: routeConfig.AcceptInsecureHTTPSTarget === true,
          UsePathPrefix: routeConfig.UsePathPrefix,
          PathPrefix: routeConfig.PathPrefix,
          StripPathPrefix: routeConfig.StripPathPrefix,
          AuthEnabled: routeConfig.AuthEnabled,
          HideFromDashboard: routeConfig.HideFromDashboard,
          _SmartShield_Enabled: (routeConfig.SmartShield ? routeConfig.SmartShield.Enabled : false),
          RestrictToConstellation: routeConfig.RestrictToConstellation === true,
          OverwriteHostHeader: routeConfig.OverwriteHostHeader,
          TunnelVia: routeConfig.TunnelVia || '',
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

            let op;
            if(newRoute) {
              op = API.config.newRoute(routeConfig.Name, fullValues)
            } else {
              op = API.config.replaceRoute(routeConfig.Name, fullValues)
            }
            
            op.then((res) => {
              if (res.status == "OK") {
                setStatus({ success: true });
                snackit('Route updated successfully', 'success')
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
            return { Name: 'Name must be unique' }
          }

          setRouteConfig && debounce(() => setRouteConfig(fullValues), 500)();
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={2}>
              <MainCard name={routeConfig.Name} title={
                noControls ? 'New URL' :
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
                    label="Name"
                    placeholder="Name"
                    formik={formik}
                  />

                  <CosmosInputText
                    name="Description"
                    label="Description"
                    placeholder="Description"
                    formik={formik}
                  />

                  <Hide h={lockTarget}>
                    <CosmosFormDivider title={'Target Type'} />
                    <Grid item xs={12}>
                      <Alert color='info'>What are you trying to access with this route?</Alert>
                    </Grid>

                    <CosmosSelect
                      name="Mode"
                      label="Mode"
                      formik={formik}
                      disabled={lockTarget}
                      options={[
                        ["SERVAPP", "ServApp - Docker Container"],
                        ["PROXY", "Proxy"],
                        ["STATIC", "Static Folder"],
                        ["SPA", "Single Page Application"],
                        ["REDIRECT", "Redirection"]
                      ]}
                    />
                  </Hide>
                  <CosmosFormDivider title={'Target Settings'} />

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
                      : <CosmosInputText
                        name="Target"
                        label={formik.values.Mode == "PROXY" ? "Target URL" : "Target Folder Path"}
                        placeholder={formik.values.Mode == "PROXY" ? "http://localhost:8080" : "/path/to/my/app"}
                        formik={formik}
                      />
                  }

                  {formik.values.Target.startsWith('https://') && <CosmosCheckbox
                    name="AcceptInsecureHTTPSTarget"
                    label="Accept Insecure HTTPS Target (not recommended)"
                    formik={formik}
                  />}

                  <CosmosFormDivider title={'Source'} />

                  <Grid item xs={12}>
                    <Alert color='info'>What URL do you want to access your target from?</Alert>
                  </Grid>

                  <CosmosCheckbox
                    name="UseHost"
                    label="Use Host"
                    formik={formik}
                  />

                  {formik.values.UseHost && (<><CosmosInputText
                    name="Host"
                    label="Host"
                    placeholder="Host"
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                    onChange={(e) => {
                      checkHost(e.target.value, setHostError)
                    }}
                  />
                  {hostError && <Grid item xs={12}>
                    <Alert color='error'>{hostError}</Alert>
                  </Grid>}
                  </>
                  )}

                  <CosmosCheckbox
                    name="UsePathPrefix"
                    label="Use Path Prefix"
                    formik={formik}
                  />

                  {formik.values.UsePathPrefix && <CosmosInputText
                    name="PathPrefix"
                    label="Path Prefix"
                    placeholder="Path Prefix"
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}

                  {formik.values.UsePathPrefix && <CosmosCheckbox
                    name="StripPathPrefix"
                    label="Strip Path Prefix"
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}
                  
                  <CosmosFormDivider title={'Basic Security'} />
                  
                  <CosmosCheckbox
                    name="AuthEnabled"
                    label="Authentication Required"
                    formik={formik}
                  />
                  
                  <CosmosCheckbox
                    name="_SmartShield_Enabled"
                    label="Smart Shield Protection"
                    formik={formik}
                  />
                  
                  <CosmosCheckbox
                    name="RestrictToConstellation"
                    label="Restrict access to Constellation VPN"
                    formik={formik}
                  />

                  {(formik.values.TunnelVia || tunnels.length > 0) && (<>
                    <CosmosSelect
                      name="TunnelVia"
                      label="Tunnel via another Constellation Cosmos node"
                      formik={formik}
                      onChange={(e) => {
                        const newHostname = tunnels.find((t) => t.deviceName === e.target.value)?.publicHostname;
                        formik.setFieldValue('TunneledHost', getHostnameFromName(formik.values.Name, null, config, newHostname));
                      }}
                      options={[
                        ['', 'None'],
                        ...tunnels.map((t) => [t.deviceName, `${t.deviceName} (${t.publicHostname})`])
                      ]}
                    />

                    {formik.values.TunnelVia && <CosmosInputText
                      name="TunneledHost"
                      label="Hostname to tunnel from (what is the user facing hostname of the tunnel)" 
                      placeholder="other-host.com"
                      formik={formik}
                    />}
                  </>)}

                  <CosmosCollapse title={'Advanced Settings'}>
                    <Stack spacing={2}>
                      <CosmosCheckbox
                        name="HideFromDashboard"
                        label="Hide from Dashboard"
                        formik={formik}
                      />

                      <CosmosFormDivider />
                      <Alert severity='info'>These settings are for advanced users only. Please do not change these unless you know what you are doing.</Alert>
                      <CosmosInputText
                        name="OverwriteHostHeader"
                        label="Overwrite Host Header (use this to chain resolve request from another server/ip)"
                        placeholder="Overwrite Host Header"
                        formik={formik}
                      />

                      <Alert severity='warning'>
                        This setting will filter out all requests that do not come from the specified IPs.
                        This requires your setup to report the true IP of the client. By default it will, but some exotic setup (like installing docker/cosmos on Windows, or behind Cloudlfare)
                        will prevent Cosmos from knowing what is the client's real IP. If you used "Restrict to Constellation" above, Constellation IPs will always be allowed regardless of this setting.
                      </Alert>

                      <CosmosInputText
                        name="WhitelistInboundIPs"
                        label="Whitelist Inbound IPs and/or IP ranges (comma separated)"
                        placeholder="Whitelist Inbound IPs"
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
                Save
              </Button></MainCard>}
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteManagement;