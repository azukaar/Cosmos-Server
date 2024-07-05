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
  const [saveLabel, setSaveLabel] = React.useState(t('Save'));
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
      }}>{t('refresh')}</Button>

      {isAdmin && <Button variant="outlined" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          setOpenRestartModal(true);
      }}>{t('restart')}</Button>}
      
      <ConfirmModal variant="outlined" color="warning" startIcon={<DeleteOutlined />} callback={() => {
          API.metrics.reset().then((res) => {
            refresh();
          });
      }}
      label={t('purgeMetrics')} 
      content={t('ConfirmPurgeMetrics')} />
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

          Email_Enabled: config.EmailConfig.Enabled,
          Email_Host: config.EmailConfig.Host,
          Email_Port: config.EmailConfig.Port,
          Email_Username: config.EmailConfig.Username,
          Email_Password: config.EmailConfig.Password,
          Email_From: config.EmailConfig.From,
          Email_UseTLS : config.EmailConfig.UseTLS,
          Email_AllowInsecureTLS : config.EmailConfig.AllowInsecureTLS,

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
          Hostname: Yup.string().max(255).required(t('HostnameRequired')),
          MongoDB: Yup.string().max(512),
          LoggingLevel: Yup.string().max(255).required(t('LoglevelRequired')),
        })}

        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          setSubmitting(true);
        
          let toSave = {
            ...config,
            MongoDB: values.MongoDB,
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
            setSaveLabel(t('Saved'));
            setTimeout(() => {
              setSaveLabel(t('Save'));
            }, 3000);
          }).catch((err) => {
            setOpenModal(true);
            setSaveLabel(t('SaveError'));
            setTimeout(() => {
              setSaveLabel(t('Save'));
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
                <Alert severity="warning">{t('WarningNotAdmin')} 
                </Alert>
              </div>} 

              <MainCard title={t('General')}>
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    <Alert severity="info">{t('Infobox.OverwriteConfFile')}</Alert>
                  </Grid>

                  <CosmosCheckbox
                    label={t('RequireMFA')}
                    name="RequireMFA"
                    formik={formik}
                    helperText={t('RequireMFAHelper')}
                  />
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="MongoDB-login">{t('MongoDB-Login')}</InputLabel>
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
                      <CosmosCollapse title={t('PuppetMode')}>
                        <Grid container spacing={3}>
                          <Grid item xs={12}>
                            <CosmosCheckbox
                              label={t('PuppetModeEnable')}
                              name="PuppetModeEnabled"
                              formik={formik}
                              helperText={t('PuppetModeEnableHelper')}
                            />

                            {formik.values.PuppetModeEnabled && (
                              <Grid container spacing={3}>
                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label={t('PuppetModeHostname')}
                                    name="PuppetModeHostname"
                                    formik={formik}
                                    helperText={t('PuppetModeHostname')}
                                  />
                                </Grid>

                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label={t('PuppetModeDbVolume')}
                                    name="PuppetModeDbVolume"
                                    formik={formik}
                                    helperText={t('PuppetModeDbVolume')}
                                  />

                                  <CosmosInputText
                                    label={t('PuppetModeConfigVolume')}
                                    name="PuppetModeConfigVolume"
                                    formik={formik}
                                    helperText={t('PuppetModeConfigVolume')}
                                  />
                                  
                                  <CosmosInputText
                                    label={t('PuppetModeVersion')}
                                    name="PuppetModeVersion"
                                    formik={formik}
                                    helperText={t('PuppetModeVersion')}
                                  />
                                </Grid>

                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label={t('PuppetModeUsername')}
                                    name="PuppetModeUsername"
                                    formik={formik}
                                    helperText={t('PuppetModeUsername')}
                                  />

                                  <CosmosInputPassword
                                    label={t('PuppetModePassword')}
                                    name="PuppetModePassword"
                                    autoComplete='new-password'
                                    formik={formik}
                                    helperText={t('PuppetModePassword')}
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
                    label={t('BackupDir')+"("+t('relativeToHost') +"`/`)"}
                    name="BackupOutputDir"
                    formik={formik}
                    helperText={t('BackupDirHelper')+"("+t('relativeToHost') +"`/`)"}
                  />
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="LoggingLevel-login">{t('Level of logging')} ({t('Default')}: INFO)</InputLabel>
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
                    label={t('MonitoringEnabled')}
                    name="MonitoringEnabled"
                    formik={formik}
                  />
                </Grid>
              </MainCard>
              
              <MainCard title={t('Appearance')}>
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    {!uploadingBackground && formik.values.Background && <img src=
                      {formik.values.Background} alt={t('PreviewBroken')}
                      width={285} />}
                    {uploadingBackground && <Skeleton variant="rectangular" width={285} height={140} />}
                     <Stack spacing={1} direction="row">
                      <UploadButtons
                        accept='.jpg, .png, .gif, .jpeg, .webp, .bmp, .avif, .tiff, .svg'
                        label={t('UploadWallpaper')}
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
                        {t('ResetWallpaper')}
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
                        {t('ResetColors')}
                      </Button>
                    </Stack>
                  </Grid>
                  
                  <Grid item xs={12}>
                    <CosmosCheckbox 
                      label={t('AppDetailsOnHomepage')}
                      name="Expanded"
                      formik={formik}
                    />
                  </Grid>
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel style={{marginBottom: '10px'}} htmlFor="PrimaryColor">{t('PrimaryColor')}</InputLabel>
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
                      <InputLabel style={{marginBottom: '10px'}} htmlFor="SecondaryColor">{t('SecondaryColor')}</InputLabel>
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
                      <InputLabel htmlFor="Hostname-login">{t('HostnameLogin')}</InputLabel>
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
                        label={<span>{t('AllowHTTPLocalIPAccess')} &nbsp;
                          <Tooltip title={<span style={{fontSize:'110%'}}><Trans i18nKey="AllowHTTPLocalIPAccessTitle">
                                When HTTPS is used along side a domain, depending on your networking configuration, it is possible that your server is not receiving direct local connections. <br />
                                This option allows you to also access your Cosmos admin using your local IP address, like ip:port. <br />
                                You can already create ip:port URLs for your apps, <strong>but this will make them HTTP-only</strong>.</Trans></span>}>
                                <QuestionCircleOutlined size={'large'} />
                            </Tooltip></span>}
                        name="AllowHTTPLocalIPAccess"
                        formik={formik}
                      />
                      {formik.values.allowHTTPLocalIPAccess && <Alert severity="warning"><Trans i18nKey="AllowHTTPLocalIPAccessAlert">
                      This option is not recommended as it exposes your server to security risks on your local network. <br />
                      Your local network is safer than the internet, but not safe, as devices like IoTs, smart-TVs, smartphones or even your router can be compromised. <br />
                      <strong>If you want to have a secure offline / local-only access to a server that uses a domain name and HTTPS, use Constellation instead.</strong>
                      </Trans></Alert>}
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
                      {t('Infobox.AllowIndex')}<br />
                    </Alert>
                    <CosmosCheckbox 
                      label={t('AllowIndexHelper')}
                      name="AllowSearchEngine"
                      formik={formik}
                    />
                  </Grid>
                  </Grid>
              </MainCard>
              
              <MainCard title="Emails - SMTP">
                <Stack spacing={2}>
                  <Alert severity="info">{t('Email.InfoEmail')}.</Alert>

                  <CosmosCheckbox 
                    label={t('Email.EnSMTP')}
                    name="Email_Enabled"
                    formik={formik}
                    helperText={t('Email.EnSMTP')}
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
                      label={t('Email.SMTPUsername')}
                      name="Email_Username"
                      formik={formik}
                      helperText={t('Email.SMTPUsername')}
                    />

                    <CosmosInputPassword
                      label={t('Email.SMTPPassword')}
                      name="Email_Password"
                      autoComplete='new-password'
                      formik={formik}
                      helperText={t('Email.SMTPPassword')}
                      noStrength
                    />

                    <CosmosInputText
                      label={t('Email.SMTPFrom')}
                      name="Email_From"
                      formik={formik}
                      helperText={t('Email.SMTPFrom')}
                    />

                    <CosmosCheckbox
                      label={t('Email.SMTPS')}
                      name="Email_UseTLS"
                      formik={formik}
                      helperText={t('Email.SMTPS')}
                    />

                    {formik.values.Email_UseTLS && (
                      <CosmosCheckbox
                        label={t('Email.AllowInsecureTLS')}
                        name="Email_AllowInsecureTLS"
                        formik={formik}
                        helperText={t('Email.AllowSelfSigned')}
                      />
                    )}
                  </>)}
                </Stack>
              </MainCard>

              <MainCard title="Docker">
                <Stack spacing={2}>
                  <CosmosCheckbox
                    label={t('SkipPruneNetwork')}
                    name="SkipPruneNetwork"
                    formik={formik}
                  />

                  <CosmosCheckbox
                    label={t('SkipPruneImages')}
                    name="SkipPruneImages"
                    formik={formik}
                  />

                  <CosmosInputText
                    label={t('DefaultDataPath')}
                    name="DefaultDataPath"
                    formik={formik}
                    placeholder={'/usr'}
                  />
                </Stack>
              </MainCard>


              <MainCard title={t('Security')}>
                  <Grid container spacing={3}>

                  {/* <CosmosCheckbox
                    label={"Read Client IP from X-Forwarded-For header (not recommended)"}
                    name="UseForwardedFor"
                    formik={formik}
                  /> */}

                  <CosmosFormDivider title='Geo-Blocking' />

                  <CosmosCheckbox
                    label={t('CoutryBlacklistIsWhitelist')}
                    name="CountryBlacklistIsWhitelist"
                    formik={formik}
                  />

                  <Grid item xs={12}>
                      <InputLabel htmlFor="GeoBlocking">{t('GeoBlocking')}
                      {formik.values.CountryBlacklistIsWhitelist ? t('allowed') : t('blocked')}
                      {t('GeoBlocking2')}</InputLabel>
                  </Grid>

                  <CountrySelect name="GeoBlocking" label={t('ChooseBlockOrAllow')} formik={formik} />

                  <Grid item xs={12}>
                    <Button onClick={() => {
                      formik.setFieldValue("GeoBlocking", ["CN","RU","TR","BR","BD","IN","NP","PK","LK","VN","ID","IR","IQ","EG","AF","RO",])
                      formik.setFieldValue("CountryBlacklistIsWhitelist", false)
                    }} variant="outlined">{t('GeoBlockResetToDefault')}</Button>
                  </Grid>
                  
                  <CosmosFormDivider title={t('AdminRestrictions')} />

                  <Grid item xs={12}>
                    <Alert severity="info">{t('AdminRestrictionsDescr')}</Alert>
                  </Grid>
                  
                  <CosmosInputText
                    label={t('AdminWhitelistIPs')}
                    name="AdminWhitelistIPs"
                    formik={formik}
                    helperText={t('AdminWhitelistIPsHelper')}
                  />

                  <CosmosCheckbox
                    label={t('AdminConstellationOnly')}
                    name="AdminConstellationOnly"
                    formik={formik}
                  />

                  <CosmosFormDivider title={t('Encryption')} />

                  <Grid item xs={12}>
                    <Alert severity="info">{t('WarningChangePrivateKeysFromRemote')}</Alert>
                  </Grid>

                  <Grid item xs={12}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <Field
                        type="checkbox"
                        name="GenerateMissingAuthCert"
                        as={FormControlLabel}
                        control={<Checkbox size="large" />}
                        label={t('GenerateMissingAuthCert')}
                      />
                    </Stack>
                  </Grid>

                  <CosmosSelect
                    name="HTTPSCertificateMode"
                    label={t('HTTPSCertificates')}
                    formik={formik}
                    onChange={(e) => {
                      formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
                    }}
                    options={[
                      ["LETSENCRYPT", t('SSLLetsEncrypt')],
                      ["SELFSIGNED", t('SSLSelfSigned')],
                      ["PROVIDED", t('SSLProvided')],
                      ["DISABLED", t('SSLDisabled')],
                    ]}
                  />

                  <CosmosCheckbox
                    label={t('UseWildcard') + formik.values.Hostname}
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
                      label={t('OverrideWildcardDomains')}
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
                        label={t('LetsEncryptEmail')}
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
                        label={t('DNSChallengeProvider')}
                        name="DNSChallengeProvider"
                        configName="DNSChallengeConfig"
                        formik={formik}
                      />
                    )
                  }

                  <Grid item xs={12}>
                    <h4>{t('AuthPublicKey')}</h4>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <pre className='code'>
                        {config.HTTPConfig.AuthPublicKey}
                      </pre>
                    </Stack>
                  </Grid>

                  <Grid item xs={12}>
                    <h4>{t('RootHTTPSPublicKey')}</h4>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <pre className='code'>
                        {config.HTTPConfig.TLSCert}
                      </pre>
                    </Stack>
                  </Grid>

                  <Grid item xs={12}>
                    <CosmosCheckbox
                      label={t('ForceHTTPSCertificateRenewal')}
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