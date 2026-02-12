import * as React from 'react';
import * as API from '../../api';
import MainCard from '../../components/MainCard';
import { Formik } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Grid,
  Stack,
  FormHelperText,
} from '@mui/material';
import RestartModal from './users/restart';
import { ArrowDownOutlined, DeleteOutlined, SyncOutlined } from '@ant-design/icons';
import { LoadingButton } from '@mui/lab';
import { useClientInfos } from '../../utils/hooks';
import ConfirmModal from '../../components/confirmModal';
import PrettyTabbedView from '../../components/tabbedView/tabbedView';

import { useTranslation } from 'react-i18next';

import ConfigGeneral from './configGeneral';
import ConfigHTTP from './configHTTP';
import ConfigHTTPS from './configHTTPS';
import ConfigExternal from './configExternal';
import ConfigAppearance from './configAppearance';

const SaveButton = ({ formik, saveLabel, isAdmin }) => {
  if (!isAdmin) return null;
  return (
    <MainCard>
      {formik.errors.submit && (
        <Grid item xs={12}>
          <FormHelperText error>{formik.errors.submit}</FormHelperText>
        </Grid>
      )}
      <Grid item xs={12}>
        <LoadingButton
          disableElevation
          loading={formik.isSubmitting}
          fullWidth
          size="large"
          type="submit"
          variant="contained"
          color="primary"
        >
          {saveLabel}
        </LoadingButton>
      </Grid>
    </MainCard>
  );
};

const ConfigManagement = () => {
  const { t } = useTranslation();
  const [config, setConfig] = React.useState(null);
  const [status, setStatus] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);
  const [openResartModal, setOpenRestartModal] = React.useState(false);
  const [saveLabel, setSaveLabel] = React.useState(t('global.saveAction'));
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });

    API.getStatus().then((res) => {
      setStatus(res.data);
    });
  }

  React.useEffect(() => {
    refresh();
  }, []);

  const wrapTab = (formik, content) => (
    <div style={{ maxWidth: "1000px", margin: "auto" }}>
      <Stack spacing={3}>
        <Stack direction="row" spacing={2}>
          <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
              refresh();
          }}>{t('mgmt.config.header.refreshButton.refreshLabel')}</Button>

          {isAdmin && <Button variant="outlined" color="primary"
            startIcon={<ArrowDownOutlined />} href={"/cosmos/_logs"}>
            {t('mgmt.config.header.downloadLogsButton.downloadLogsLabel')}
          </Button>}

          <ConfirmModal variant="outlined" color="warning" startIcon={<DeleteOutlined />} callback={() => {
              API.metrics.reset().then(() => {
                refresh();
              });
          }}
          label={t('mgmt.config.header.purgeMetricsButton.purgeMetricsLabel')}
          content={t('mgmt.config.header.purgeMetricsButton.purgeMetricsPopUp.cofirmAction')} />
        </Stack>

        {!isAdmin && <div>
          <Alert severity="warning">{t('mgmt.config.general.notAdminWarning')}
          </Alert>
        </div>}

        <SaveButton formik={formik} saveLabel={saveLabel} isAdmin={isAdmin} />
        {content}
        <SaveButton formik={formik} saveLabel={saveLabel} isAdmin={isAdmin} />
      </Stack>
    </div>
  );

  return <div>
    {config && <>
      <RestartModal openModal={openModal} setOpenModal={setOpenModal} config={config} />
      <RestartModal openModal={openResartModal} setOpenModal={setOpenRestartModal} />

      <Formik
        initialValues={{
          MongoDB: config.MongoDB,
          LoggingLevel: config.LoggingLevel,
          RequireMFA: config.RequireMFA,
          GeoBlocking: config.BlockedCountries,
          CountryBlacklistIsWhitelist: config.CountryBlacklistIsWhitelist,
          AutoUpdate: config.AutoUpdate,
          BetaUpdates: config.BetaUpdates,

          Licence: config.Licence,
          ServerToken: config.ServerToken,

          Hostname: config.HTTPConfig.Hostname,
          GenerateMissingTLSCert: config.HTTPConfig.GenerateMissingTLSCert,
          GenerateMissingAuthCert: config.HTTPConfig.GenerateMissingAuthCert,
          HTTPPort: config.HTTPConfig.HTTPPort,
          HTTPSPort: config.HTTPConfig.HTTPSPort,
          SSLEmail: config.HTTPConfig.SSLEmail,
          UseWildcardCertificate: config.HTTPConfig.UseWildcardCertificate,
          HTTPSCertificateMode: config.HTTPConfig.HTTPSCertificateMode,
          DNSChallengeProvider: config.HTTPConfig.DNSChallengeProvider,
          DNSChallengeConfig: config.HTTPConfig.DNSChallengeConfig,
          DisablePropagationChecks: config.HTTPConfig.DisablePropagationChecks,
          DNSChallengePropagationWait: config.HTTPConfig.DNSChallengePropagationWait || "",
          DNSChallengeResolvers: config.HTTPConfig.DNSChallengeResolvers,
          ForceHTTPSCertificateRenewal: config.HTTPConfig.ForceHTTPSCertificateRenewal,
          OverrideWildcardDomains: config.HTTPConfig.OverrideWildcardDomains,
          UseForwardedFor: config.HTTPConfig.UseForwardedFor,
          TrustedProxies: config.HTTPConfig.TrustedProxies && config.HTTPConfig.TrustedProxies.join(', '),
          AllowSearchEngine: config.HTTPConfig.AllowSearchEngine,
          AllowHTTPLocalIPAccess: config.HTTPConfig.AllowHTTPLocalIPAccess,
          PublishMDNS: config.HTTPConfig.PublishMDNS,

          Email_Enabled: config.EmailConfig.Enabled,
          Email_Host: config.EmailConfig.Host,
          Email_Port: config.EmailConfig.Port,
          Email_Username: config.EmailConfig.Username,
          Email_Password: config.EmailConfig.Password,
          Email_From: config.EmailConfig.From,
          Email_UseTLS : config.EmailConfig.UseTLS,
          Email_AllowInsecureTLS : config.EmailConfig.AllowInsecureTLS,
          Email_NotifyLogin: config.EmailConfig.NotifyLogin,

          SkipPruneNetwork: config.DockerConfig.SkipPruneNetwork,
          SkipPruneImages: config.DockerConfig.SkipPruneImages,
          DefaultDataPath: config.DockerConfig.DefaultDataPath || "/cosmos-storage",

          Background: config && config.HomepageConfig && config.HomepageConfig.Background,
          Expanded: config && config.HomepageConfig && config.HomepageConfig.Expanded,
          PrimaryColor: config && config.ThemeConfig && config.ThemeConfig.PrimaryColor,
          SecondaryColor: config && config.ThemeConfig && config.ThemeConfig.SecondaryColor,

          MonitoringEnabled: !config.MonitoringDisabled,

          BackupOutputDir: config.BackupOutputDir,
          IncrBackupOutputDir: config.Backup && config.Backup.Backups && config.Backup.Backups["Cosmos Internal Backup"] ? config.Backup.Backups["Cosmos Internal Backup"].Repository : "",

          AdminWhitelistIPs: config.AdminWhitelistIPs && config.AdminWhitelistIPs.join(', '),
          AdminConstellationOnly: config.AdminConstellationOnly,

          PuppetModeEnabled: config.Database.PuppetMode,
          PuppetModeHostname: config.Database.Hostname,
          PuppetModeDbVolume: config.Database.DbVolume,
          PuppetModeConfigVolume: config.Database.ConfigVolume,
          PuppetModeVersion: config.Database.Version,
          PuppetModeUsername: config.Database.Username,
          PuppetModePassword: config.Database.Password,
        }}

        validationSchema={Yup.object().shape({
          Hostname: Yup.string().max(255).required(t('mgmt.config.http.hostnameInput.HostnameValidation')),
          MongoDB: Yup.string().max(512),
          LoggingLevel: Yup.string().max(255).required(t('mgmt.config.general.logLevelInput.logLevelValidation')),
        })}

        onSubmit={async (values, { setSubmitting }) => {
          setSubmitting(true);

          if(!values.IncrBackupOutputDir && config.Backup.Backups && config.Backup.Backups["Cosmos Internal Backup"]) {
            delete config.Backup.Backups["Cosmos Internal Backup"];
          }

          let toSave = {
            ...config,
            MongoDB: values.MongoDB,
            Licence: values.Licence,
            ServerToken: values.ServerToken,
            Database: {
              ...config.Database,
              PuppetMode: values.PuppetModeEnabled,
              Hostname: values.PuppetModeHostname,
              DbVolume: values.PuppetModeDbVolume,
              ConfigVolume: values.PuppetModeConfigVolume,
              Version: values.PuppetModeVersion,
              Username: values.PuppetModeUsername,
              Password: values.PuppetModePassword,
            },
            LoggingLevel: values.LoggingLevel,
            RequireMFA: values.RequireMFA,
            AutoUpdate: values.AutoUpdate,
            BetaUpdates: values.BetaUpdates,
            BlockedCountries: values.GeoBlocking,
            CountryBlacklistIsWhitelist: values.CountryBlacklistIsWhitelist,
            MonitoringDisabled: !values.MonitoringEnabled,
            BackupOutputDir: values.BackupOutputDir,
            AdminConstellationOnly: values.AdminConstellationOnly,
            AdminWhitelistIPs: (values.AdminWhitelistIPs && values.AdminWhitelistIPs != "") ?
              values.AdminWhitelistIPs.split(',').map((x) => x.trim()) : [],
            HTTPConfig: {
              ...config.HTTPConfig,
              Hostname: values.Hostname,
              GenerateMissingAuthCert: values.GenerateMissingAuthCert,
              HTTPPort: values.HTTPPort,
              HTTPSPort: values.HTTPSPort,
              SSLEmail: values.SSLEmail,
              UseWildcardCertificate: values.UseWildcardCertificate,
              HTTPSCertificateMode: values.HTTPSCertificateMode,
              DNSChallengeProvider: values.DNSChallengeProvider,
              DNSChallengeConfig: values.DNSChallengeConfig,
              DisablePropagationChecks: values.DisablePropagationChecks,
              DNSChallengePropagationWait: values.DNSChallengePropagationWait,
              DNSChallengeResolvers: values.DNSChallengeResolvers,
              ForceHTTPSCertificateRenewal: values.ForceHTTPSCertificateRenewal,
              OverrideWildcardDomains: values.OverrideWildcardDomains.replace(/\s/g, ''),
              UseForwardedFor: values.UseForwardedFor,
              TrustedProxies: (values.TrustedProxies && values.TrustedProxies != "") ?
                values.TrustedProxies.split(',').map((x) => x.trim()) : [],
              AllowSearchEngine: values.AllowSearchEngine,
              AllowHTTPLocalIPAccess: values.AllowHTTPLocalIPAccess,
              PublishMDNS: values.PublishMDNS,
            },
            EmailConfig: {
              ...config.EmailConfig,
              Enabled: values.Email_Enabled,
              Host: values.Email_Host,
              Port: values.Email_Port,
              Username: values.Email_Username,
              Password: values.Email_Password,
              From: values.Email_From,
              UseTLS: values.Email_UseTLS,
              AllowInsecureTLS: values.Email_AllowInsecureTLS,
              NotifyLogin: values.Email_NotifyLogin,
            },
            DockerConfig: {
              ...config.DockerConfig,
              SkipPruneNetwork: values.SkipPruneNetwork,
              SkipPruneImages: values.SkipPruneImages,
              DefaultDataPath: values.DefaultDataPath
            },
            HomepageConfig: {
              ...config.HomepageConfig,
              Background: values.Background,
              Expanded: values.Expanded
            },
            ThemeConfig: {
              ...config.ThemeConfig,
              PrimaryColor: values.PrimaryColor,
              SecondaryColor: values.SecondaryColor
            },
            Backup: {
              ...config.Backup,
              Backups: values.IncrBackupOutputDir ? {
                ...(config.Backup.Backups || {}),
                "Cosmos Internal Backup": {
                  ...(config.Backup.Backups ? config.Backup.Backups["Cosmos Internal Backup"] : {}),
                  "Crontab": "0 0 4 * * *",
                  "CrontabForget": "0 0 12 * * *",
                  "Source": status && status.ConfigFolder,
                  "Password": "",
                  "Name": "Cosmos Internal Backup",
                  Repository: values.IncrBackupOutputDir
                }
              } : (config.Backup.Backups || {})
            }
          }

          return API.config.set(toSave).then(() => {
            setOpenModal(true);
            setSaveLabel(t('global.savedConfirmation'));
            setTimeout(() => {
              setSaveLabel(t('global.saveAction'));
            }, 3000);
          }).catch(() => {
            setOpenModal(true);
            setSaveLabel(t('global.savedError'));
            setTimeout(() => {
              setSaveLabel(t('global.saveAction'));
            }, 3000);
          });
        }}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            <PrettyTabbedView
              rootURL="/cosmos-ui/config-general"
              tabs={[
                {
                  title: t('mgmt.config.generalTitle'),
                  children: wrapTab(formik, <ConfigGeneral formik={formik} config={config} status={status} isAdmin={isAdmin} />),
                  url: "/general",
                },
                {
                  title: "HTTP",
                  children: wrapTab(formik, <ConfigHTTP formik={formik} />),
                  url: "/http-server",
                },
                {
                  title: "HTTPS",
                  children: wrapTab(formik, <ConfigHTTPS formik={formik} config={config} />),
                  url: "/https",
                },
                {
                  title: t('mgmt.config.appearanceTitle'),
                  children: wrapTab(formik, <ConfigAppearance formik={formik} />),
                  url: "/appearance",
                },
                {
                  title: t('mgmt.config.externalTitle'),
                  children: wrapTab(formik, <ConfigExternal formik={formik} />),
                  url: "/external",
                },
              ]}
            />
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default ConfigManagement;
