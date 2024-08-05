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

const RouteSecurity = ({ routeConfig, config }) => {
  const [openModal, setOpenModal] = React.useState(false);

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
              snackit('Route updated successfully', 'success');
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
              <MainCard name={routeConfig.Name} title={'Security'}>
                <Grid container spacing={2}>
                    <CosmosFormDivider title={'Authentication'} />

                    <CosmosCheckbox
                      name="AuthEnabled"
                      label="Authentication Required"
                      formik={formik}
                    />

                    <CosmosCheckbox
                      name="AdminOnly"
                      label="Admin only"
                      formik={formik}
                    />

                    <CosmosFormDivider title={'Headers'} />

                    <CosmosCheckbox
                      name="DisableHeaderHardening"
                      label="Disable Header Hardening"
                      formik={formik}
                    />

                    <CosmosFormDivider title={'Smart Shield'} />

                    <CosmosCheckbox
                      name="_SmartShield_Enabled"
                      label="Smart Shield Protection"
                      formik={formik}
                    />

                    <CosmosSelect
                      name="_SmartShield_PolicyStrictness"
                      label="Policy Strictness"
                      placeholder="Policy Strictness"
                      options={[
                        [0, 'Default'],
                        [1, 'Strict'],
                        [2, 'Normal'],
                        [3, 'Lenient'],
                      ]}
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserTimeBudget"
                      label="Per User Time Budget in milliseconds (0 for default)"
                      placeholder="Per User Time Budget"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserRequestLimit"
                      label="Per User Request Limit (0 for default)"
                      placeholder="Per User Request Limit"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserByteLimit"
                      label="Per User Byte Limit (0 for default)"
                      placeholder="Per User Byte Limit"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_PerUserSimultaneous"
                      label="Per User Simultaneous Connections Limit (0 for default)"
                      placeholder="Per User Simultaneous Connections Limit"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="_SmartShield_MaxGlobalSimultaneous"
                      label="Max Global Simultaneous Connections Limit (0 for default)"
                      placeholder="Max Global Simultaneous Connections Limit"
                      type="number"
                      formik={formik}
                    />

                    <CosmosSelect
                      name="_SmartShield_PrivilegedGroups"
                      label="Privileged Groups "
                      placeholder="Privileged Group"
                      options={[
                        [0, 'Default'],
                        [1, 'Users & Admins'],
                        [2, 'Admin Only'],
                      ]}
                      formik={formik}
                    />
                    <CosmosFormDivider title={'Limits'} />

                    <CosmosInputText
                      name="Timeout"
                      label="Timeout in milliseconds (0 for no timeout, at least 60000 or less recommended)"
                      placeholder="Timeout"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="MaxBandwith"
                      label="Maximum Bandwith limit per user in bytes per seconds (0 for no limit)"
                      placeholder="Maximum Bandwith"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="ThrottlePerMinute"
                      label="Maximum number of requests Per Minute (0 for no limit, at least 2000 or less recommended)"
                      placeholder="Throttle Per Minute"
                      type="number"
                      formik={formik}
                    />

                    <CosmosInputText
                      name="CORSOrigin"
                      label="Custom CORS Origin (Recommended to leave blank)"
                      placeholder="CORS Origin"
                      formik={formik}
                    />

                    <CosmosCheckbox
                      name="BlockCommonBots"
                      label="Block Common Bots (Recommended)"
                      formik={formik}
                    />

                    <CosmosCheckbox
                      name="BlockAPIAbuse"
                      label="Block requests without Referer header"
                      formik={formik}
                    />
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
                Save
              </Button></MainCard>
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default RouteSecurity;