import * as React from 'react';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik } from 'formik';
import {
  Alert,
  Button,
  Divider,
  Grid,
  Stack,

} from '@mui/material';
import RestartModal from '../users/restart';
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText, CosmosSelect } from '../users/formShortcuts';
import { snackit } from '../../../api/wrap';
import { IsRouteSocketProxy } from '../../../utils/routes';
import { useTranslation } from 'react-i18next';

const RouteSecurity = ({ routeConfig, config }) => {
  const { t } = useTranslation();
  const [openModal, setOpenModal] = React.useState(false);
  const isNotSocketProxy = !IsRouteSocketProxy(routeConfig);

  return <div style={{ maxWidth: '1000px', width: '100%', margin: '', position: 'relative' }}>
    <RestartModal openModal={openModal} setOpenModal={setOpenModal} config={config} />

    {routeConfig && <>
      <Formik
        initialValues={{
          AuthEnabled: routeConfig.AuthEnabled,
          Timeout: routeConfig.Timeout,
          ThrottlePerMinute: routeConfig.ThrottlePerMinute,
          CORSOrigin: routeConfig.CORSOrigin,
          MaxBandwith: routeConfig.MaxBandwith,
          BlockAPIAbuse: routeConfig.BlockAPIAbuse,
          BlockCommonBots: routeConfig.BlockCommonBots,
          AdminOnly: routeConfig.AdminOnly,
          SpoofHostname: routeConfig.SpoofHostname,
          DisableHeaderHardening: routeConfig.DisableHeaderHardening,
          _SmartShield_Enabled: (routeConfig.SmartShield ? routeConfig.SmartShield.Enabled : false),
          _SmartShield_PolicyStrictness: (routeConfig.SmartShield ? routeConfig.SmartShield.PolicyStrictness : 0),
          _SmartShield_PerUserTimeBudget: (routeConfig.SmartShield ? routeConfig.SmartShield.PerUserTimeBudget : 0),
          _SmartShield_PerUserRequestLimit: (routeConfig.SmartShield ? routeConfig.SmartShield.PerUserRequestLimit : 0),
          _SmartShield_PerUserByteLimit: (routeConfig.SmartShield ? routeConfig.SmartShield.PerUserByteLimit : 0),
          _SmartShield_PerUserSimultaneous: (routeConfig.SmartShield ? routeConfig.SmartShield.PerUserSimultaneous : 0),
          _SmartShield_MaxGlobalSimultaneous: (routeConfig.SmartShield ? routeConfig.SmartShield.MaxGlobalSimultaneous : 0),
          _SmartShield_PrivilegedGroups: (routeConfig.SmartShield ? routeConfig.SmartShield.PrivilegedGroups : []),
        }}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          const fullValues = {
            ...routeConfig,
            ...values,
          }

          if(!fullValues.SmartShield) {
            fullValues.SmartShield = {};
          }

          fullValues.SmartShield.Enabled = values._SmartShield_Enabled;
          delete fullValues._SmartShield_Enabled;
          fullValues.SmartShield.PolicyStrictness = values._SmartShield_PolicyStrictness;
          delete fullValues._SmartShield_PolicyStrictness;
          fullValues.SmartShield.PerUserTimeBudget = values._SmartShield_PerUserTimeBudget;
          delete fullValues._SmartShield_PerUserTimeBudget;
          fullValues.SmartShield.PerUserRequestLimit = values._SmartShield_PerUserRequestLimit;
          delete fullValues._SmartShield_PerUserRequestLimit;
          fullValues.SmartShield.PerUserByteLimit = values._SmartShield_PerUserByteLimit;
          delete fullValues._SmartShield_PerUserByteLimit;
          fullValues.SmartShield.PerUserSimultaneous = values._SmartShield_PerUserSimultaneous;
          delete fullValues._SmartShield_PerUserSimultaneous;
          fullValues.SmartShield.MaxGlobalSimultaneous = values._SmartShield_MaxGlobalSimultaneous;
          delete fullValues._SmartShield_MaxGlobalSimultaneous;
          fullValues.SmartShield.PrivilegedGroups = values._SmartShield_PrivilegedGroups;
          delete fullValues._SmartShield_PrivilegedGroups;

          API.config.replaceRoute(routeConfig.Name, fullValues).then((res) => {
            if (res.status == "OK") {
              setStatus({ success: true });
              snackit(t('mgmt.urls.edit.updateSuccess'), 'success');
              setSubmitting(false);
              setOpenModal(true);
            } else {
              setStatus({ success: false });
              setErrors({ submit: res.status });
              setSubmitting(false);
            }
          });
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <Stack spacing={2}>
            <MainCard name={routeConfig.Name} title={t('global.securityTitle')}>
            <Grid container spacing={2}>
                    {isNotSocketProxy && <>
                      <CosmosFormDivider title={t('auth.authTitle')} />

                      <CosmosCheckbox
                        name="AuthEnabled"
                      label={t('mgmt.urls.edit.basicSecurity.authEnabledCheckbox.authEnabledLabel')}
                        formik={formik}
                      />

                      <CosmosCheckbox
                        name="AdminOnly"
                        label={t('mgmt.urls.edit.basicSecurity.adminOnlyCheckbox.adminOnlyLabel')}
                        formik={formik}
                      />

                      <CosmosFormDivider title={t('mgmt.urls.edit.basicSecurity.headers.headersTitle')} />
  
                      <CosmosCheckbox
                        name="DisableHeaderHardening"
                        label={t('mgmt.urls.edit.basicSecurity.disableHeaderHardeningCheckbox.disableHeaderHardeningLabel')}
                        formik={formik}
                      />
                    </>}

                    <CosmosFormDivider title={t('mgmt.urls.edit.basicSecurity.smartShield.smartShieldTitle')} />

                    <CosmosCheckbox
                      name="_SmartShield_Enabled"
                      label={t('mgmt.urls.edit.basicSecurity.smartShieldEnabledCheckbox.smartShieldEnabledLabel')}
                      formik={formik}
                    />

                    <CosmosSelect
                      name="_SmartShield_PolicyStrictness"
                      label={t('mgmt.urls.edit.basicSecurity.smartShieldPolicyStrictness.smartShieldPolicySelectTitle')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShieldPolicyStrictness.smartShieldPolicySelectTitle')}
                      options={[
                        [0, t('global.default')],
                        [1, t('mgmt.urls.edit.basicSecurity.smartShieldPolicyStrictness.smartShieldPolicySelectChoice.strict')],
                        [2, t('global.normal')],
                        [3, t('mgmt.urls.edit.basicSecurity.smartShieldPolicyStrictness.smartShieldPolicySelectChoice.lenient')],
                      ]}
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserTimeBudget"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.timeBudget.timeBudgetLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.timeBudget.timeBudgetPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserRequestLimit"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.requestLimit.requestLimitLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.requestLimit.requestLimitPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserByteLimit"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.byteLimit.byteLimitLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.byteLimit.byteLimitPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserSimultaneous"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.connectionLimitUser.connectionLimitUserLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.connectionLimitUser.connectionLimitUserPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_MaxGlobalSimultaneous"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.connectionLimitGlobal.connectionLimitGlobalLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.connectionLimitGlobal.connectionLimitGlobalPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosSelect
                      name="_SmartShield_PrivilegedGroups"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.privilegedGroups.privilegedGroupsLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.privilegedGroups.privilegedGroupsPlaceholder')}
                      options={[
                        [0, t('global.default')],
                        [1, t('mgmt.urls.edit.basicSecurity.smartShield.privilegedGroups.privilegedGroupsSelection.usersAdmins')],
                        [2, t('mgmt.urls.edit.basicSecurity.smartShield.privilegedGroups.privilegedGroupsSelection.adminsOnly')],
                      ]}
                      formik={formik}
                    />
                    
                    {isNotSocketProxy && <>
                    <CosmosFormDivider title={t('mgmt.urls.edit.basicSecurity.smartShield.limits.limitsTitle')} />

                    <CosmosInputText
                      name="Timeout"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.limits.timeout.timeoutLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.limits.timeout.timeoutPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="MaxBandwith"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.limits.bandwidth.bandwidthLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.limits.bandwidth.bandwidthPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="ThrottlePerMinute"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.limits.requests.requestsPlaceholder')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.limits.requests.requestsPlaceholder')}
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="CORSOrigin"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.cors.corsLabel')}
                      placeholder={t('mgmt.urls.edit.basicSecurity.smartShield.cors.corsPlaceholder')}
                      formik={formik}
                    />

                    <CosmosCheckbox
                      name="BlockCommonBots"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.commonBotsCheckbox.commomBotsLabel')}
                      formik={formik}
                    />

                    <CosmosCheckbox
                      name="BlockAPIAbuse"
                      label={t('mgmt.urls.edit.basicSecurity.smartShield.noReffererCheckbox.noReffererLabel')}
                      formik={formik}
                    />
                  </>}
                </Grid>
              </MainCard>
              <MainCard ><Button
                fullWidth
                disableElevation
                size="large"
                type="submit"
                variant="contained"
                color="primary"
              >
                {t('global.saveAction')}
              </Button></MainCard>
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteSecurity;