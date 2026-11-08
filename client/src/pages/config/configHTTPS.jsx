import * as React from 'react';
import MainCard from '../../components/MainCard';
import { Field } from 'formik';
import {
  Alert,
  Checkbox,
  FormControlLabel,
  Grid,
  Stack,
} from '@mui/material';
import { CosmosCheckbox, CosmosInputText, CosmosSelect } from './users/formShortcuts';
import ConfigZones from './configZones';
import * as API from '../../api';
import { useTranslation } from 'react-i18next';

const ConfigHTTPS = ({ formik, config, serverCertificate }) => {
  const { t } = useTranslation();
  const [addresses, setAddresses] = React.useState([]);

  React.useEffect(() => {
    API.zones.addresses().then((res) => setAddresses(res.data || [])).catch(() => setAddresses([]));
  }, []);

  const addressOptions = [{ value: "", address: "" }, ...addresses.filter((a) => a.value !== "")].map((a) => {
    const detected = a.value === "" ? (addresses.find((d) => d.value === "") || {}).address : a.address;
    const label = a.value === "" ? t('mgmt.config.zones.advertised.public') : a.value.replace("iface:", "");
    return [a.value, label + (detected ? " (" + detected + ")" : "")];
  });
  if (formik.values.AdvertisedAddress && !addressOptions.find((o) => o[0] === formik.values.AdvertisedAddress)) {
    addressOptions.push([formik.values.AdvertisedAddress, formik.values.AdvertisedAddress]);
  }

  return (
    <Stack spacing={3}>
    <ConfigZones />
    <MainCard title={t('mgmt.config.zones.generalTitle')}>
      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Alert severity="info">{t('mgmt.config.security.encryption.enryptionInfo')}</Alert>
        </Grid>

        <Grid item xs={12}>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
            <Field
              type="checkbox"
              name="GenerateMissingAuthCert"
              as={FormControlLabel}
              control={<Checkbox size="large" />}
              label={t('mgmt.config.security.encryption.genMissingAuthCheckbox.genMissingAuthLabel')}
            />
          </Stack>
        </Grid>

        <CosmosSelect
          name="HTTPSCertificateMode"
          label={t('mgmt.config.zones.serverModeLabel')}
          formik={formik}
          onChange={(e) => {
            formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
          }}
          options={[
            ["LETSENCRYPT", t('mgmt.config.zones.serverModeHTTPS')],
            ["DISABLED", t('mgmt.config.security.encryption.httpsCertSelection.sslDisabledChoice')],
          ]}
        />

        <CosmosSelect
          name="AdvertisedAddress"
          label={t('mgmt.config.zones.advertised.label')}
          helperText={t('mgmt.config.zones.advertised.helper')}
          formik={formik}
          options={addressOptions}
        />

        {formik.values.HTTPSCertificateMode !== "DISABLED" && (
            <CosmosInputText
              name="SSLEmail"
              onChange={(e) => {
                formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
              }}
              label={t('mgmt.config.security.encryption.sslLetsEncryptEmailInput.sslLetsEncryptEmailLabel')}
              formik={formik}
            />
          )
        }

        <Grid item xs={12}>
          <h4>{t('mgmt.config.security.encryption.authPubKeyTitle')}</h4>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
            <pre className='code'>
              {config.HTTPConfig.AuthPublicKey}
            </pre>
          </Stack>
        </Grid>

        {serverCertificate && <Grid item xs={12}>
          <h4>{t('mgmt.config.security.encryption.rootHttpsPubKeyTitle')}</h4>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
            <pre className='code'>
              {serverCertificate}
            </pre>
          </Stack>
        </Grid>}

        <Grid item xs={12}>
          <CosmosCheckbox
            label={t('mgmt.config.security.encryption.sslCertForceRenewCheckbox.sslCertForceRenewLabel')}
            name="ForceHTTPSCertificateRenewal"
            formik={formik}
          />
        </Grid>
      </Grid>
    </MainCard>
    </Stack>
  );
};

export default ConfigHTTPS;
