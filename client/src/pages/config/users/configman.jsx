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

const ConfigManagement = () => {
  const [config, setConfig] = React.useState(null);
  const [openModal, setOpenModal] = React.useState(false);
  const [openResartModal, setOpenRestartModal] = React.useState(false);
  const [uploadingBackground, setUploadingBackground] = React.useState(false);
  const [saveLabel, setSaveLabel] = React.useState("Save");
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
      }}>Refresh</Button>

      {isAdmin && <Button variant="outlined" color="primary" startIcon={<SyncOutlined />} onClick={() => {
          setOpenRestartModal(true);
      }}>Restart Server</Button>}
      
      <ConfirmModal variant="outlined" color="warning" startIcon={<DeleteOutlined />} callback={() => {
          API.metrics.reset().then((res) => {
            refresh();
          });
      }}
      label={'Purge Metrics Dashboard'} 
      content={'Are you sure you want to purge all the metrics data from the dashboards?'} />
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
          Hostname: Yup.string().max(255).required('Hostname is required'),
          MongoDB: Yup.string().max(512),
          LoggingLevel: Yup.string().max(255).required('Logging Level is required'),
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
            setSaveLabel("Saved!");
            setTimeout(() => {
              setSaveLabel("Save");
            }, 3000);
          }).catch((err) => {
            setOpenModal(true);
            setSaveLabel("Error while saving, try again.");
            setTimeout(() => {
              setSaveLabel("Save");
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
                <Alert severity="warning">As you are not an admin, you can't edit the configuration.
                This page is only here for visibility. 
                </Alert>
              </div>} 

              <MainCard title="General">
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    <Alert severity="info">This page allow you to edit the configuration file. Any Environment Variable overwritting configuration won't appear here.</Alert>
                  </Grid>

                  <CosmosCheckbox
                    label="Force Multi-Factor Authentication"
                    name="RequireMFA"
                    formik={formik}
                    helperText="Require MFA for all users"
                  />
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="MongoDB-login">MongoDB connection string. It is advised to use Environment variable to store this securely instead. (Optional)</InputLabel>
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
                      <CosmosCollapse title="Puppet Mode">
                        <Grid container spacing={3}>
                          <Grid item xs={12}>
                            <CosmosCheckbox
                              label="Puppet Mode Enabled"
                              name="PuppetModeEnabled"
                              formik={formik}
                              helperText="Enable Puppet Mode"
                            />

                            {formik.values.PuppetModeEnabled && (
                              <Grid container spacing={3}>
                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label="Puppet Mode Hostname"
                                    name="PuppetModeHostname"
                                    formik={formik}
                                    helperText="Puppet Mode Hostname"
                                  />
                                </Grid>

                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label="Puppet Mode Database Volume"
                                    name="PuppetModeDbVolume"
                                    formik={formik}
                                    helperText="Puppet Mode Database Volume"
                                  />

                                  <CosmosInputText
                                    label="Puppet Mode Config Volume"
                                    name="PuppetModeConfigVolume"
                                    formik={formik}
                                    helperText="Puppet Mode Config Volume"
                                  />
                                  
                                  <CosmosInputText
                                    label="Puppet Mode Version"
                                    name="PuppetModeVersion"
                                    formik={formik}
                                    helperText="Puppet Mode Version"
                                  />
                                </Grid>

                                <Grid item xs={12}>
                                  <CosmosInputText
                                    label="Puppet Mode Username"
                                    name="PuppetModeUsername"
                                    formik={formik}
                                    helperText="Puppet Mode Username"
                                  />

                                  <CosmosInputPassword
                                    label="Puppet Mode Password"
                                    name="PuppetModePassword"
                                    autoComplete='new-password'
                                    formik={formik}
                                    helperText="Puppet Mode Password"
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
                    label="Backup Output Directory (relative to the host server `/`)"
                    name="BackupOutputDir"
                    formik={formik}
                    helperText="Directory where backups will be stored (relative to the host server `/`)"
                  />
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel htmlFor="LoggingLevel-login">Level of logging (Default: INFO)</InputLabel>
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
                    label="Monitoring Enabled"
                    name="MonitoringEnabled"
                    formik={formik}
                  />
                </Grid>
              </MainCard>
              
              <MainCard title="Appearance">
                <Grid container spacing={3}>
                  <Grid item xs={12}>
                    {!uploadingBackground && formik.values.Background && <img src=
                      {formik.values.Background} alt="preview seems broken. Please re-upload."
                      width={285} />}
                    {uploadingBackground && <Skeleton variant="rectangular" width={285} height={140} />}
                     <Stack spacing={1} direction="row">
                      <UploadButtons
                        accept='.jpg, .png, .gif, .jpeg, .webp, .bmp, .avif, .tiff, .svg'
                        label="Upload Wallpaper"
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
                        Reset Wallpaper
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
                        Reset Colors
                      </Button>
                    </Stack>
                  </Grid>
                  
                  <Grid item xs={12}>
                    <CosmosCheckbox 
                      label="Show Application Details on Homepage"
                      name="Expanded"
                      formik={formik}
                    />
                  </Grid>
                  
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <InputLabel style={{marginBottom: '10px'}} htmlFor="PrimaryColor">Primary Color</InputLabel>
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
                      <InputLabel style={{marginBottom: '10px'}} htmlFor="SecondaryColor">Secondary Color</InputLabel>
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
                      <InputLabel htmlFor="Hostname-login">Hostname: This will be used to restrict access to your Cosmos Server (Your IP, or your domain name)</InputLabel>
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
                        label={<span>Allow insecure access via local IP &nbsp;
                          <Tooltip title={<span style={{fontSize:'110%'}}>
                            When HTTPS is used along side a domain, depending on your networking configuration, it is possible that your server is not receiving direct local connections. <br />
                            This option allows you to also access your Cosmos admin using your local IP address, like ip:port. <br />
                            You can already create ip:port URLs for your apps, <strong>but this will make them HTTP-only</strong>.</span>}>
                              <QuestionCircleOutlined size={'large'} />
                          </Tooltip></span>}
                        name="AllowHTTPLocalIPAccess"
                        formik={formik}
                      />
                    {formik.values.AllowHTTPLocalIPAccess && <Alert severity="warning">
                      This option is not recommended as it exposes your server to security risks on your local network. <br />
                      Your local network is safer than the internet, but not safe, as devices like IoTs, smart-TVs, smartphones or even your router can be compromised. <br />
                      <strong>If you want to have a secure offline / local-only access to a server that uses a domain name and HTTPS, use Constellation instead.</strong>
                    </Alert>}
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
                      Enable this option if you have a public site and want to allow search engines to find it, so it appears on search results. <br />
                    </Alert>
                    <CosmosCheckbox 
                      label="Allow search engines to index your server"
                      name="AllowSearchEngine"
                      formik={formik}
                    />
                  </Grid>
                  </Grid>
              </MainCard>
              
              <MainCard title="Emails - SMTP">
                <Stack spacing={2}>
                  <Alert severity="info">This allow you to setup an SMTP server for Cosmos to send emails such as password reset emails and invites.</Alert>

                  <CosmosCheckbox 
                    label="Enable SMTP"
                    name="Email_Enabled"
                    formik={formik}
                    helperText="Enable SMTP"
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
                      label="SMTP Username"
                      name="Email_Username"
                      formik={formik}
                      helperText="SMTP Username"
                    />

                    <CosmosInputPassword
                      label="SMTP Password"
                      name="Email_Password"
                      autoComplete='new-password'
                      formik={formik}
                      helperText="SMTP Password"
                      noStrength
                    />

                    <CosmosInputText
                      label="SMTP From"
                      name="Email_From"
                      formik={formik}
                      helperText="SMTP From"
                    />

                    <CosmosCheckbox
                      label="SMTP Uses TLS"
                      name="Email_UseTLS"
                      formik={formik}
                      helperText="SMTP Uses TLS"
                    />

                    {formik.values.Email_UseTLS && (
                      <CosmosCheckbox
                        label="Allow Insecure TLS"
                        name="Email_AllowInsecureTLS"
                        formik={formik}
                        helperText="Allow self-signed certificate"
                      />
                    )}
                  </>)}
                </Stack>
              </MainCard>

              <MainCard title="Docker">
                <Stack spacing={2}>
                  <CosmosCheckbox
                    label="Do not clean up Network"
                    name="SkipPruneNetwork"
                    formik={formik}
                  />

                  <CosmosCheckbox
                    label="Do not clean up Images"
                    name="SkipPruneImages"
                    formik={formik}
                  />

                  <CosmosInputText
                    label="Default data path for installs"
                    name="DefaultDataPath"
                    formik={formik}
                    placeholder={'/usr'}
                  />
                </Stack>
              </MainCard>


              <MainCard title="Security">
                  <Grid container spacing={3}>

                  {/* <CosmosCheckbox
                    label={"Read Client IP from X-Forwarded-For header (not recommended)"}
                    name="UseForwardedFor"
                    formik={formik}
                  /> */}

                  <CosmosFormDivider title='Geo-Blocking' />

                  <CosmosCheckbox
                    label={"Use list as whitelist instead of blacklist"}
                    name="CountryBlacklistIsWhitelist"
                    formik={formik}
                  />

                  <Grid item xs={12}>
                    <InputLabel htmlFor="GeoBlocking">Geo-Blocking: (Those countries will be 
                    {formik.values.CountryBlacklistIsWhitelist ? " allowed to access " : " blocked from accessing "}
                    your server)</InputLabel>
                  </Grid>

                  <CountrySelect name="GeoBlocking" label="Choose which countries you want to block or allow" formik={formik} />

                  <Grid item xs={12}>
                    <Button onClick={() => {
                      formik.setFieldValue("GeoBlocking", ["CN","RU","TR","BR","BD","IN","NP","PK","LK","VN","ID","IR","IQ","EG","AF","RO",])
                      formik.setFieldValue("CountryBlacklistIsWhitelist", false)
                    }} variant="outlined">Reset to default (most dangerous countries)</Button>
                  </Grid>
                  
                  <CosmosFormDivider title='Admin Restrictions' />

                  <Alert severity="info">
                    Use those options to restrict access to the admin panel. Be careful, if you lock yourself out, you will need to manually edit the config file.
                    To restrict the access to your local network, you can use the "Admin Whitelist" with the IP range 192.168.0.0/16 
                  </Alert>
                  
                  <CosmosInputText
                    label="Admin Whitelist Inbound IPs and/or IP ranges (comma separated)"
                    name="AdminWhitelistIPs"
                    formik={formik}
                    helperText="Comma separated list of IPs that will be allowed to access the admin panel"
                  />

                  <CosmosCheckbox
                    label={"Only allow access to the admin panel from the constellation"}
                    name="AdminConstellationOnly"
                    formik={formik}
                  />

                  <CosmosFormDivider title='Encryption' />

                  <Grid item xs={12}>
                    <Alert severity="info">For security reasons, It is not possible to remotely change the Private keys of any certificates on your instance. It is advised to manually edit the config file, or better, use Environment Variables to store them.</Alert>
                  </Grid>

                  <Grid item xs={12}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <Field
                        type="checkbox"
                        name="GenerateMissingAuthCert"
                        as={FormControlLabel}
                        control={<Checkbox size="large" />}
                        label="Generate missing Authentication Certificates automatically (Default: true)"
                      />
                    </Stack>
                  </Grid>

                  <CosmosSelect
                    name="HTTPSCertificateMode"
                    label="HTTPS Certificates"
                    formik={formik}
                    onChange={(e) => {
                      formik.setFieldValue("ForceHTTPSCertificateRenewal", true);
                    }}
                    options={[
                      ["LETSENCRYPT", "Automatically generate certificates using Let's Encrypt (Recommended)"],
                      ["SELFSIGNED", "Locally self-sign certificates (unsecure)"],
                      ["PROVIDED", "I have my own certificates"],
                      ["DISABLED", "Do not use HTTPS (very unsecure)"],
                    ]}
                  />

                  <CosmosCheckbox
                    label={"Use Wildcard Certificate for the root domain of " + formik.values.Hostname}
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
                      label="(optional, only if you know what you are doing) Override Wildcard Domains (comma separated, need to add both wildcard AND root domain like in the placeholder)"
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
                        label="Email address for Let's Encrypt"
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
                        label="Pick a DNS provider (if you are using a DNS Challenge, otherwise leave empty)"
                        name="DNSChallengeProvider"
                        configName="DNSChallengeConfig"
                        formik={formik}
                      />
                    )
                  }

                  <Grid item xs={12}>
                    <h4>Authentication Public Key</h4>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <pre className='code'>
                        {config.HTTPConfig.AuthPublicKey}
                      </pre>
                    </Stack>
                  </Grid>

                  <Grid item xs={12}>
                    <h4>Root HTTPS Public Key</h4>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <pre className='code'>
                        {config.HTTPConfig.TLSCert}
                      </pre>
                    </Stack>
                  </Grid>

                  <Grid item xs={12}>
                    <CosmosCheckbox
                      label={"Force HTTPS Certificate Renewal On Next Save"}
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