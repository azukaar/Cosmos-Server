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
import { CosmosCheckbox, CosmosCollapse, CosmosInputText, CosmosSelect } from './users/formShortcuts';
import { DnsChallengeComp } from '../../utils/dns-challenge-comp';
import { useTranslation } from 'react-i18next';

const ConfigHTTPS = ({ formik, config }) => {
  const { t } = useTranslation();

  return (
    <MainCard title="HTTPS">
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
          label={t('mgmt.config.security.encryption.httpsCertSelection.httpsCertLabel')}
          formik={formik}
          onChange={(e) => {
            formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
          }}
          options={[
            ["LETSENCRYPT", t('mgmt.config.security.encryption.httpsCertSelection.sslLetsEncryptChoice')],
            ["SELFSIGNED", t('mgmt.config.security.encryption.httpsCertSelection.sslSelfSignedChoice')],
            ["PROVIDED", t('mgmt.config.security.encryption.httpsCertSelection.sslProvidedChoice')],
            ["DISABLED", t('mgmt.config.security.encryption.httpsCertSelection.sslDisabledChoice')],
          ]}
        />

        <CosmosCheckbox
          label={t('mgmt.config.security.encryption.wildcardCheckbox.wildcardLabel') + formik.values.Hostname}
          onChange={(e) => {
            formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
          }}
          name="UseWildcardCertificate"
          formik={formik}
        />

        {formik.values.UseWildcardCertificate && (
          <CosmosInputText
            name="OverrideWildcardDomains"
            onChange={(e) => {
              formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
            }}
            label={t('mgmt.config.security.encryption.overwriteWildcardInput.overwriteWildcardLabel')}
            formik={formik}
            placeholder={"example.com,*.example.com"}
          />
        )}

        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && (
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

        {
          formik.values.HTTPSCertificateMode === "LETSENCRYPT" && (
            <DnsChallengeComp
              onChange={(e) => {
                formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
              }}
              label={t('mgmt.config.security.encryption.sslLetsEncryptDnsSelection.sslLetsEncryptDnsLabel')}
              name="DNSChallengeProvider"
              configName="DNSChallengeConfig"
              formik={formik}
            />
          )
        }

        {formik.values.HTTPSCertificateMode === "LETSENCRYPT" && formik.values.DNSChallengeProvider && (
          <Grid item xs={12}>
            <CosmosCollapse title={t('mgmt.config.security.encryption.dnsChallengeAdvanced.title')}>
              <Grid container spacing={3}>
                <CosmosInputText
                  name="DNSChallengeResolvers"
                  label={t('mgmt.config.security.encryption.dnsChallengeAdvanced.resolversLabel')}
                  formik={formik}
                  helperText={t('mgmt.config.security.encryption.dnsChallengeAdvanced.resolversHelperText')}
                  placeholder="1.1.1.1:53,8.8.8.8:53"
                />

                <CosmosCheckbox
                  label={t('mgmt.config.security.encryption.dnsChallengeAdvanced.disablePropagationChecksLabel')}
                  name="DisablePropagationChecks"
                  formik={formik}
                  helperText={t('mgmt.config.security.encryption.dnsChallengeAdvanced.disablePropagationChecksHelperText')}
                />
                
                <CosmosInputText
                  name="DNSChallengePropagationWait"
                  label={t('mgmt.config.security.encryption.dnsChallengeAdvanced.propagationWaitLabel')}
                  formik={formik}
                  helperText={t('mgmt.config.security.encryption.dnsChallengeAdvanced.propagationWaitHelperText')}
                  placeholder="30"
                />
              </Grid>
            </CosmosCollapse>
          </Grid>
        )}

        <Grid item xs={12}>
          <h4>{t('mgmt.config.security.encryption.authPubKeyTitle')}</h4>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
            <pre className='code'>
              {config.HTTPConfig.AuthPublicKey}
            </pre>
          </Stack>
        </Grid>

        <Grid item xs={12}>
          <h4>{t('mgmt.config.security.encryption.rootHttpsPubKeyTitle')}</h4>
          <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
            <pre className='code'>
              {config.HTTPConfig.TLSCert}
            </pre>
          </Stack>
        </Grid>

        <Grid item xs={12}>
          <CosmosCheckbox
            label={t('mgmt.config.security.encryption.sslCertForceRenewCheckbox.sslCertForceRenewLabel')}
            name="ForceHTTPSCertificateRenewal"
            formik={formik}
          />
        </Grid>
      </Grid>
    </MainCard>
  );
};

export default ConfigHTTPS;
