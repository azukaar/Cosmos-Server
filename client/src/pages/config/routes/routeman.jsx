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
} from '@mui/material';
import RestartModal from '../users/restart';
import { CosmosCheckbox, CosmosCollapse, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../users/formShortcuts';
import { CosmosContainerPicker } from '../users/containerPicker';
import { snackit } from '../../../api/wrap';
import { ValidateRouteSchema, sanitizeRoute } from '../../../utils/routes';
import { isDomain } from '../../../utils/indexs';
import { useTranslation } from 'react-i18next';

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
  const { t } = useTranslation();
  const [openModal, setOpenModal] = React.useState(false);
  const [hostError, setHostError] = React.useState(null);

  React.useEffect(() => {
    if(routeConfig && routeConfig.Host) {
      checkHost(routeConfig.Host, setHostError);
    }
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
                    label={t('Name')}
                    placeholder={t('Name')}
                    formik={formik}
                  />

                  <CosmosInputText
                    name="Description"
                    label={t('Description')}
                    placeholder={t('Description')}
                    formik={formik}
                  />

                  <Hide h={lockTarget}>
                    <CosmosFormDivider title={t('TargetType')} />
                    <Grid item xs={12}>
                      <Alert color='info'>{t('TargetTypeInfo')}</Alert>
                    </Grid>

                    <CosmosSelect
                      name="Mode"
                      label={t('Mode')}
                      formik={formik}
                      disabled={lockTarget}
                      options={[
                        ["SERVAPP", "ServApp - Docker Container"],
                        ["PROXY", "Proxy"],
                        ["STATIC", t('StaticFolder')],
                        ["SPA", t('SinglePageApplication')],
                        ["REDIRECT", t('Redirection')]
                      ]}
                    />
                  </Hide>
                  <CosmosFormDivider title={t('TargetSettings')} />

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
                        label={formik.values.Mode == "PROXY" ? t('TargetURL') : t('TargetFolderPath')}
                        placeholder={formik.values.Mode == "PROXY" ? "http://localhost:8080" : "/path/to/my/app"}
                        formik={formik}
                      />
                  }

                  {formik.values.Target.startsWith('https://') && <CosmosCheckbox
                    name="AcceptInsecureHTTPSTarget"
                    label={t('AcceptInsecureHTTPSTarget')}
                    formik={formik}
                  />}

                  <CosmosFormDivider title={t('Source')} />

                  <Grid item xs={12}>
                    <Alert color='info'>{t('SourceInfo')}</Alert>
                  </Grid>

                  <CosmosCheckbox
                    name="UseHost"
                    label={t('UseHost')}
                    formik={formik}
                  />

                  {formik.values.UseHost && (<><CosmosInputText
                    name="Host"
                    label={t('Host')}
                    placeholder={t('Host')}
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
                    label={t('UsePathPrefix')}
                    formik={formik}
                  />

                  {formik.values.UsePathPrefix && <CosmosInputText
                    name="PathPrefix"
                    label={t('PathPrefix')}
                    placeholder={t('PathPrefix')}
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}

                  {formik.values.UsePathPrefix && <CosmosCheckbox
                    name="StripPathPrefix"
                    label={t('StripPathPrefix')}
                    formik={formik}
                    style={{ paddingLeft: '20px' }}
                  />}
                  
                  <CosmosFormDivider title={'Basic Security'} />
                  
                  <CosmosCheckbox
                    name="AuthEnabled"
                    label={t('AuthEnabled')}
                    formik={formik}
                  />
                  
                  <CosmosCheckbox
                    name="_SmartShield_Enabled"
                    label={t('SmartShieldEnabled')}
                    formik={formik}
                  />
                  
                  <CosmosCheckbox
                    name="RestrictToConstellation"
                    label={t('RestrictToConstellation')}
                    formik={formik}
                  />

                  <CosmosCollapse title={'Advanced Settings'}>
                    <Stack spacing={2}>
                      <CosmosCheckbox
                        name="HideFromDashboard"
                        label={t('HideFromDashboard')}
                        formik={formik}
                      />

                      <CosmosFormDivider />
                      <Alert severity='info'>{t('AdvancedUsersOnlyInfo')}</Alert>
                      <CosmosInputText
                        name="OverwriteHostHeader"
                        label={t('OverwriteHostHeader')}
                        placeholder={t('OverwriteHostHeaderplaceholder')}
                        formik={formik}
                      />

                      <Alert severity='warning'>
                        {t('warningFilterIP')}
                      </Alert>

                      <CosmosInputText
                        name="WhitelistInboundIPs"
                        label={t('WhitelistInboundIPs')}
                        placeholder={t('WhitelistInboundIPs')}
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
                {t('Save')}
              </Button></MainCard>}
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteManagement;