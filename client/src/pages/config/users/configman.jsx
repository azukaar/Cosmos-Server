import * as React from 'react';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import { Formik, Field } from 'formik';
import * as Yup from 'yup';
import {
  Alert,
  Button,
  Checkbox,
  FormControlLabel,
  Grid,
  InputLabel,
  OutlinedInput,
  Stack,
  FormHelperText,
  TextField,
  MenuItem,
  Skeleton,
  Tooltip,
} from '@mui/material';
import RestartModal from './restart';
import { DeleteOutlined, QuestionCircleOutlined, SyncOutlined, WarningFilled } from '@ant-design/icons';
import { CosmosCheckbox, CosmosCollapse, CosmosFormDivider, CosmosInputPassword, CosmosInputText, CosmosSelect } from './formShortcuts';
import CountrySelect from '../../../components/countrySelect';
import { DnsChallengeComp } from '../../../utils/dns-challenge-comp';

import UploadButtons from '../../../components/fileUpload';
import { SliderPicker
 } from 'react-color';
 import { LoadingButton } from '@mui/lab';

 // TODO: Remove circular deps
 import {SetPrimaryColor, SetSecondaryColor} from '../../../App';
import { useClientInfos } from '../../../utils/hooks';
import ConfirmModal from '../../../components/confirmModal';
import { DownloadFile } from '../../../api/downloadButton';
import { isDomain } from '../../../utils/indexs';

import { Trans, useTranslation } from 'react-i18next';

const ConfigManagement = () => {
  const { t } = useTranslation();
  const [config, setConfig] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);
  const [openResartModal, setOpenRestartModal] = React.useState(false);
  const [uploadingBackground, setUploadingBackground] = React.useState(false);
  const [saveLabel, setSaveLabel] = React.useState(t('global.saveAction'));
  const {role} = useClientInfos();
  const isAdmin = role === "2";

  function refresh() {
    API.config.get().then((res) => {
      setConfig(res.data);
    });
  }

  function getRouteDomain(domain) {
    let parts = domain.split('.');
    return parts[parts.length - 2] + '.' + parts[parts.length - 1];
  }

  React.useEffect(() => {
    refresh();
  }, []);

  return <div style={{maxWidth: '1000px', margin: ''}}>
    <Stack direction="row" spacing={2} style={{marginBottom: '15px'}}>
      <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          refresh();
      }}>{t('mgmt.config.header.refreshButton.refreshLabel')}</Button>

      {isAdmin && <Button variant="outlined" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          setOpenRestartModal(true);
      }}>{t('mgmt.config.header.restartButton.restartLabel')}</Button>}
      
      <ConfirmModal variant="outlined" color="warning" startIcon={<DeleteOutlined />} callback={() => {
          API.metrics.reset().then((res) => {
            refresh();
          });
      }}
      label={t('mgmt.config.header.purgeMetricsButton.purgeMetricsLabel')} 
      content={t('mgmt.config.header.purgeMetricsButton.purgeMetricsPopUp.cofirmAction')} />
    </Stack>
    
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
          ForceHTTPSCertificateRenewal: config.HTTPConfig.ForceHTTPSCertificateRenewal,
          OverrideWildcardDomains: config.HTTPConfig.OverrideWildcardDomains,
          UseForwardedFor: config.HTTPConfig.UseForwardedFor,
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
          DefaultDataPath: config.DockerConfig.DefaultDataPath || "/usr",

          Background: config && config.HomepageConfig && config.HomepageConfig.Background,
          Expanded: config && config.HomepageConfig && config.HomepageConfig.Expanded,
          PrimaryColor: config && config.ThemeConfig && config.ThemeConfig.PrimaryColor,
          SecondaryColor: config && config.ThemeConfig && config.ThemeConfig.SecondaryColor,
        
          MonitoringEnabled: !config.MonitoringDisabled,

          BackupOutputDir: config.BackupOutputDir,

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

        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          setSubmitting(true);
        
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
            // AutoUpdate: values.AutoUpdate,
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
              ForceHTTPSCertificateRenewal: values.ForceHTTPSCertificateRenewal,
              OverrideWildcardDomains: values.OverrideWildcardDomains.replace(/\s/g, ''),
              UseForwardedFor: values.UseForwardedFor,
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
          }
          
          return API.config.set(toSave).then((data) => {
            setOpenModal(true);
            setSaveLabel(t('global.savedConfirmation'));
            setTimeout(() => {
              setSaveLabel(t('global.saveAction'));
            }, 3000);
          }).catch((err) => {
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
            <Stack spacing={3}>
            {isAdmin && <MainCard>
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
              </MainCard>}

              {!isAdmin && <div>
                <Alert severity="warning">{t('mgmt.config.general.notAdminWarning')} 
                </Alert>
              </div>} 

              <MainCard title={t('mgmt.config.generalTitle')}>
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    <Alert severity="info">{t('mgmt.config.general.configFileInfo')}</Alert>
                  </Grid>

                  <CosmosCheckbox
                    label={t('mgmt.config.general.forceMfaCheckbox.forceMfaLabel')}
                    name="RequireMFA"
                    formik={formik}
                    helperText={t('mgmt.config.general.forceMfaCheckbox.forceMfaHelperText')}
                  />
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="MongoDB-login">{t('mgmt.config.general.mongoDbInput')}</InputLabel>
                      <OutlinedInput
                        id="MongoDB-login"
                        type="password"
                        autoComplete='new-password'
                        value={formik.values.MongoDB}
                        name="MongoDB"
                        onBlur={formik.handleBlur}
                        onChange={formik.handleChange}
                        placeholder="MongoDB"
                        fullWidth
                        error={Boolean(formik.touched.MongoDB && formik.errors.MongoDB)}
                      />
                      {formik.touched.MongoDB && formik.errors.MongoDB && (
                        <FormHelperText error id="standard-weight-helper-text-MongoDB-login">
                          {formik.errors.MongoDB}
                        </FormHelperText>
                      )}
                      <CosmosCollapse title={t('mgmt.config.general.puppetModeTitle')}>
                        <Grid container spacing={3}>
                          <Grid item xs={12}>
                            <CosmosCheckbox
                              label={t('mgmt.config.general.puppetMode.enableCheckbox.enableLabel')}
                              name="PuppetModeEnabled"
                              formik={formik}
                              helperText={t('mgmt.config.general.puppetMode.enableCheckbox.enableHelperText')}
                            />

                            {formik.values.PuppetModeEnabled && (
                              <Grid container spacing={3}>
                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label={t('mgmt.config.general.puppetMode.hostnameInput.hostnameLabel')}
                                    name="PuppetModeHostname"
                                    formik={formik}
                                    helperText={t('mgmt.config.general.puppetMode.hostnameInput.hostnameHelperText')}
                                  />
                                </Grid>

                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label={t('mgmt.config.general.puppetMode.dbVolumeInput.dbVolumeLabel')}
                                    name="PuppetModeDbVolume"
                                    formik={formik}
                                    helperText={t('mgmt.config.general.puppetMode.dbVolumeInput.dbVolumeHelperText')}
                                  />

                                  <CosmosInputText
                                    label={t('mgmt.config.general.puppetMode.configVolumeInput.configVolumeLabel')}
                                    name="PuppetModeConfigVolume"
                                    formik={formik}
                                    helperText={t('mgmt.config.general.puppetMode.configVolumeInput.configVolumeHelperText')}
                                  />
                                  
                                  <CosmosInputText
                                    label={t('mgmt.config.general.puppetMode.versionInput.versionLabel')}
                                    name="PuppetModeVersion"
                                    formik={formik}
                                    helperText={t('mgmt.config.general.puppetMode.versionInput.versionHelperText')}
                                  />
                                </Grid>

                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label={t('mgmt.config.general.puppetMode.usernameInput.usernameLabel')}
                                    name="PuppetModeUsername"
                                    formik={formik}
                                    helperText={t('mgmt.config.general.puppetMode.usernameInput.usernameHelperText')}
                                  />

                                  <CosmosInputPassword
                                    label={t('mgmt.config.general.puppetMode.passwordInput.passwordLabel')}
                                    name="PuppetModePassword"
                                    autoComplete='new-password'
                                    formik={formik}
                                    helperText={t('mgmt.config.general.puppetMode.passwordInput.passwordHelperText')}
                                    noStrength
                                  />
                                </Grid>
                              </Grid>
                            )}
                          </Grid>
                        </Grid>
                      </CosmosCollapse>
                    </Stack>
                  </Grid>

                  <CosmosInputText
                    label={t('mgmt.config.general.backupDirInput.backupDirLabel')}
                    name="BackupOutputDir"
                    formik={formik}
                    helperText={t('mgmt.config.general.backupDirInput.backupDirHelperText')}
                  />
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="LoggingLevel-login">{t('mgmt.config.general.logLevelInput')}</InputLabel>
                      <TextField
                        className="px-2 my-2"
                        variant="outlined"
                        name="LoggingLevel"
                        id="LoggingLevel"
                        select
                        value={formik.values.LoggingLevel}
                        onChange={formik.handleChange}
                        error={
                          formik.touched.LoggingLevel &&
                          Boolean(formik.errors.LoggingLevel)
                        }
                        helperText={
                          formik.touched.LoggingLevel && formik.errors.LoggingLevel
                        }
                      >
                        <MenuItem key={"DEBUG"} value={"DEBUG"}>
                          DEBUG
                        </MenuItem>
                        <MenuItem key={"INFO"} value={"INFO"}>
                          INFO
                        </MenuItem>
                        <MenuItem key={"WARNING"} value={"WARNING"}>
                          WARNING
                        </MenuItem>
                        <MenuItem key={"ERROR"} value={"ERROR"}>
                          ERROR
                        </MenuItem>
                      </TextField>
                    </Stack>
                  </Grid>

                  <CosmosCheckbox
                    label={t('mgmt.config.general.monitoringCheckbox.monitoringLabel')}
                    name="MonitoringEnabled"
                    formik={formik}
                  />

                  <CosmosInputPassword
                    noStrength
                    label={t('mgmt.config.general.licenceInput.licenceLabel')}
                    name="Licence"
                    formik={formik}
                    helperText="Licence Key"
                    onChange={(e) => {
                      formik.setFieldValue("ServerToken", "");
                    }}
                  />
                  {formik.values.Licence && 
                  <Grid item xs={12}>
                    <Stack spacing={1}><Button 
                      href='https://billing.stripe.com/p/login/28obMlc7X1jgeqY144'
                      target="_blank"
                      variant="outlined">{t('mgmt.config.general.licenceInput.manageLicenceButton')}</Button>
                    </Stack>
                  </Grid>
                  }
                </Grid>
              </MainCard>
              
              <MainCard title={t('mgmt.config.appearanceTitle')}>
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    {!uploadingBackground && formik.values.Background && <img src=
                      {formik.values.Background} alt={t('mgmt.config.appearance.uploadWallpaperButton.previewBrokenError')}
                      width={285} />}
                    {uploadingBackground && <Skeleton variant="rectangular" width={285} height={140} />}
                     <Stack spacing={1} direction="row">
                      <UploadButtons
                        accept='.jpg, .png, .gif, .jpeg, .webp, .bmp, .avif, .tiff, .svg'
                        label={t('mgmt.config.appearance.uploadWallpaperButton.uploadWallpaperLabel')}
                        OnChange={(e) => {
                          setUploadingBackground(true);
                          const file = e.target.files[0];
                          API.uploadImage(file, "background").then((data) => {
                            formik.setFieldValue('Background', data.data.path);
                            setUploadingBackground(false);
                          });
                        }}
                      />

                      <Button
                        variant="outlined"
                        onClick={() => {
                          formik.setFieldValue('Background', "");
                        }}
                      >
                        {t('mgmt.config.appearance.resetWallpaperButton.resetWallpaperLabel')}
                      </Button>
                      <Button
                        variant="outlined"
                        onClick={() => {
                          formik.setFieldValue('PrimaryColor', "");
                          SetPrimaryColor("");
                          formik.setFieldValue('SecondaryColor', "");
                          SetSecondaryColor("");
                        }}
                      >
                        {t('mgmt.config.appearance.resetColorsButton.resetColorsLabel')}
                      </Button>
                    </Stack>
                  </Grid>
                  
                  <Grid item xs={12}>
                    <CosmosCheckbox 
                      label={t('mgmt.config.appearance.appDetailsOnHomepageCheckbox.appDetailsOnHomepageLabel')}
                      name="Expanded"
                      formik={formik}
                    />
                  </Grid>
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel style={{marginBottom: '10px'}} htmlFor="PrimaryColor">{t('mgmt.config.appearance.primaryColorSlider')}</InputLabel>
                      <SliderPicker
                        id="PrimaryColor"
                        color={formik.values.PrimaryColor}
                        onChange={color => {
                          let colorRGB = `rgba(${color.rgb.r}, ${color.rgb.g}, ${color.rgb.b}, ${color.rgb.a})`
                          formik.setFieldValue('PrimaryColor', colorRGB);
                          SetPrimaryColor(colorRGB);
                        }}
                      />
                    </Stack>
                  </Grid>
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel style={{marginBottom: '10px'}} htmlFor="SecondaryColor">{t('mgmt.config.appearance.secondaryColorSlider')}</InputLabel>
                      <SliderPicker
                        id="SecondaryColor"
                        color={formik.values.SecondaryColor}
                        onChange={color => {
                          let colorRGB = `rgba(${color.rgb.r}, ${color.rgb.g}, ${color.rgb.b}, ${color.rgb.a})`
                          formik.setFieldValue('SecondaryColor', colorRGB);
                          SetSecondaryColor(colorRGB);
                        }}
                      />
                    </Stack>
                  </Grid>
                </Grid>
              </MainCard>

              <MainCard title="HTTP">
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="Hostname-login">{t('mgmt.config.http.hostnameInput.HostnameLabel')}</InputLabel>
                      <OutlinedInput
                        id="Hostname-login"
                        type="text"
                        value={formik.values.Hostname}
                        name="Hostname"
                        onBlur={formik.handleBlur}
                        onChange={formik.handleChange}
                        placeholder="Hostname"
                        fullWidth
                        error={Boolean(formik.touched.Hostname && formik.errors.Hostname)}
                      />
                      {formik.touched.Hostname && formik.errors.Hostname && (
                        <FormHelperText error id="standard-weight-helper-text-Hostname-login">
                          {formik.errors.Hostname}
                        </FormHelperText>
                      )}
                    </Stack>
                  </Grid>

                  {(formik.values.HTTPSCertificateMode != "DISABLED" || isDomain(formik.values.Hostname)) ? (
                  <Grid item xs={12}>
                      <CosmosCheckbox 
                        label={<span>{t('mgmt.config.http.allowInsecureLocalAccessCheckbox.allowInsecureLocalAccessLabel')} &nbsp;
                          <Tooltip title={<span style={{fontSize:'110%'}}><Trans i18nKey="mgmt.config.http.allowInsecureLocalAccessCheckbox.allowInsecureLocalAccessTooltip" /></span>}>
                                <QuestionCircleOutlined size={'large'} />
                            </Tooltip></span>}
                        name="AllowHTTPLocalIPAccess"
                        formik={formik}
                      />
                      {formik.values.allowHTTPLocalIPAccess && <Alert severity="warning"><Trans i18nKey="mgmt.config.http.allowInsecureLocalAccessCheckbox.allowInsecureLocalAccessWarning" /></Alert>}
                  </Grid>) : ""}

                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="HTTPPort-login">HTTP Port (Default: 80)</InputLabel>
                      <OutlinedInput
                        id="HTTPPort-login"
                        type="text"
                        value={formik.values.HTTPPort}
                        name="HTTPPort"
                        onBlur={formik.handleBlur}
                        onChange={formik.handleChange}
                        placeholder="HTTPPort"
                        fullWidth
                        error={Boolean(formik.touched.HTTPPort && formik.errors.HTTPPort)}
                      />
                      {formik.touched.HTTPPort && formik.errors.HTTPPort && (
                        <FormHelperText error id="standard-weight-helper-text-HTTPPort-login">
                          {formik.errors.HTTPPort}
                        </FormHelperText>
                      )}
                    </Stack>
                  </Grid>

                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="HTTPSPort-login">HTTPS Port (Default: 443)</InputLabel>
                      <OutlinedInput
                        id="HTTPSPort-login"
                        type="text"
                        value={formik.values.HTTPSPort}
                        name="HTTPSPort"
                        onBlur={formik.handleBlur}
                        onChange={formik.handleChange}
                        placeholder="HTTPSPort"
                        fullWidth
                        error={Boolean(formik.touched.HTTPSPort && formik.errors.HTTPSPort)}
                      />
                      {formik.touched.HTTPSPort && formik.errors.HTTPSPort && (
                        <FormHelperText error id="standard-weight-helper-text-HTTPSPort-login">
                          {formik.errors.HTTPSPort}
                        </FormHelperText>
                      )}
                    </Stack>
                  </Grid>
                  <Grid item xs={12}>
                    <Alert severity="info">
                      {t('mgmt.config.http.allowSearchIndexCheckbox')}<br />
                    </Alert>
                    <CosmosCheckbox 
                      label={t('mgmt.config.http.allowSearchIndexCheckbox.allowSearchIndexLabel')}
                      name="AllowSearchEngine"
                      formik={formik}
                    />
                  </Grid>
                  <Grid item xs={12}>
                    <Alert severity="info">
                      {t('mgmt.config.http.publishMDNSCheckbox')}
                    </Alert>
                    <CosmosCheckbox 
                      label="Publish .local domains on your local network with mDNS"
                      name="PublishMDNS"
                      formik={formik}
                    />
                  </Grid>
                  </Grid>
              </MainCard>
              
              <MainCard title="Emails - SMTP">
                <Stack spacing={2}>
                  <Alert severity="info">{t('mgmt.config.email.inbobox.label')}.</Alert>

                  <CosmosCheckbox 
                    label={t('mgmt.config.email.enableCheckbox.enableLabel')}
                    name="Email_Enabled"
                    formik={formik}
                    helperText={t('mgmt.config.email.enableCheckbox.enableHelperText')}
                  />

                  {formik.values.Email_Enabled && (<>
                    <CosmosInputText
                      label="SMTP Host"
                      name="Email_Host"
                      formik={formik}
                      helperText="SMTP Host"
                    />

                    <CosmosInputText
                      label="SMTP Port"
                      name="Email_Port"
                      formik={formik}
                      helperText="SMTP Port"
                    />

                    <CosmosInputText
                      label={t('mgmt.config.email.usernameInput.usernameLabel')}
                      name="Email_Username"
                      formik={formik}
                      helperText={t('mgmt.config.email.usernameInput.usernameHelperText')}
                    />

                    <CosmosInputPassword
                      label={t('mgmt.config.email.passwordInput.passwordLabel')}
                      name="Email_Password"
                      autoComplete='new-password'
                      formik={formik}
                      helperText={t('mgmt.config.email.passwordInput.passwordHelperText')}
                      noStrength
                    />

                    <CosmosInputText
                      label={t('mgmt.config.email.senderInput.senderLabel')}
                      name="Email_From"
                      formik={formik}
                      helperText={t('mgmt.config.email.senderInput.senderHelperText')}
                    />

                    <CosmosCheckbox
                      label={t('mgmt.config.email.tlsCheckbox.tlsLabel')}
                      name="Email_UseTLS"
                      formik={formik}
                      helperText={t('mgmt.config.email.tlsCheckbox.tlsLabel')}
                    />

                    {formik.values.Email_UseTLS && (
                      <CosmosCheckbox
                        label={t('mgmt.config.email.selfSignedCheckbox.SelfSignedLabel')}
                        name="Email_AllowInsecureTLS"
                        formik={formik}
                        helperText={t('mgmt.config.email.selfSignedCheckbox.SelfSignedHelperText')}
                      />
                    )}
                    
                    <CosmosCheckbox
                      label={t('mgmt.config.email.notifyLoginCheckbox.notifyLoginLabel')}
                      name="Email_NotifyLogin"
                      formik={formik}
                    />
                  </>)}
                </Stack>
              </MainCard>

              <MainCard title="Docker">
                <Stack spacing={2}>
                  <CosmosCheckbox
                    label={t('mgmt.config.docker.skipPruneNetworkCheckbox.skipPruneNetworkLabel')}
                    name="SkipPruneNetwork"
                    formik={formik}
                  />

                  <CosmosCheckbox
                    label={t('mgmt.config.docker.skipPruneImageCheckbox.skipPruneImageLabel')}
                    name="SkipPruneImages"
                    formik={formik}
                  />

                  <CosmosInputText
                    label={t('mgmt.config.docker.defaultDatapathInput.defaultDatapathLabel')}
                    name="DefaultDataPath"
                    formik={formik}
                    placeholder={'/usr'}
                  />
                </Stack>
              </MainCard>


              <MainCard title={t('global.securityTitle')}>
                  <Grid container spacing={3}>

                  {/* <CosmosCheckbox
                    label={"Read Client IP from X-Forwarded-For header (not recommended)"}
                    name="UseForwardedFor"
                    formik={formik}
                  /> */}

                  <CosmosFormDivider title='Geo-Blocking' />

                  <CosmosCheckbox
                    label={t('mgmt.config.security.invertBlacklistCheckbox.invertBlacklistLabel')}
                    name="CountryBlacklistIsWhitelist"
                    formik={formik}
                  />

                  <Grid item xs={12}>
                      <InputLabel htmlFor="GeoBlocking">
                        {t('mgmt.config.security.geoBlockSelection.geoBlockLabel', {blockAllow: formik.values.CountryBlacklistIsWhitelist ? t('mgmt.config.security.geoBlockSelection.geoBlockLabel.varAllow') : t('mgmt.config.security.geoBlockSelection.geoBlockLabel.varBlock')
                          })}
                      </InputLabel>
                  </Grid>

                  <CountrySelect name="GeoBlocking" label={t('mgmt.config.security.geoBlockSelection', { blockAllow: formik.values.CountryBlacklistIsWhitelist ? t('mgmt.config.security.geoBlockSelection.varAllow') : t('mgmt.config.security.geoBlockSelection.varBlock') })} formik={formik} />

                  <Grid item xs={12}>
                    <Button onClick={() => {
                      formik.setFieldValue("GeoBlocking", ["CN","RU","TR","BR","BD","IN","NP","PK","LK","VN","ID","IR","IQ","EG","AF","RO",])
                      formik.setFieldValue("CountryBlacklistIsWhitelist", false)
                    }} variant="outlined">{t('mgmt.config.security.geoblock.resetToDefaultButton')}</Button>
                  </Grid>
                  
                  <CosmosFormDivider title={t('mgmt.config.security.adminRestrictionsTitle')} />

                  <Grid item xs={12}>
                    <Alert severity="info">{t('mgmt.config.security.adminRestrictions.adminRestrictionsInfo')}</Alert>
                  </Grid>
                  
                  <CosmosInputText
                    label={t('mgmt.config.security.adminRestrictions.adminWhitelistInput.adminWhitelistLabel')}
                    name="AdminWhitelistIPs"
                    formik={formik}
                    helperText={t('mgmt.config.security.adminRestrictions.adminWhitelistInput.adminWhitelistHelperText')}
                  />

                  <CosmosCheckbox
                    label={t('mgmt.config.security.adminRestrictions.adminConstellationCheckbox.adminConstellationLabel')}
                    name="AdminConstellationOnly"
                    formik={formik}
                  />

                  <CosmosFormDivider title={t('mgmt.config.security.encryptionTitle')} />

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

              {isAdmin && <MainCard>
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
              </MainCard>}
            </Stack>
          </form>
        )}
      </Formik>
    </>}
  </div>;
}

export default ConfigManagement;