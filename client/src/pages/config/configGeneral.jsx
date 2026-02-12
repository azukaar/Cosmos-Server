import * as React from 'react';
import * as API from '../../api';
import MainCard from '../../components/MainCard';
import { decodeJwt, importSPKI, jwtVerify } from 'jose';
import {
  Alert,
  Button,
  Grid,
  InputLabel,
  Stack,
  FormHelperText,
  TextField,
  MenuItem,
} from '@mui/material';
import { WarningFilled } from '@ant-design/icons';
import { CosmosCheckbox, CosmosInputPassword, CosmosInputText } from './users/formShortcuts';
import { LoadingButton } from '@mui/lab';
import { useTranslation } from 'react-i18next';
import { FilePickerButton } from '../../components/filePicker';

// License issuer's public key for signature verification
const LICENSE_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE8QolLbFdVfU3XPkC01NwsS94bv1W
Ijy+/SYjyHfakFQm7JDhKpbNPC5oc+e4uM6Y9UyC0686toqpTYBSzbgaQw==
-----END PUBLIC KEY-----`;

const validateAndDecodeJWT = (token) => {
  try {
    const payload = decodeJwt(token);
    return { error: null, data: payload };
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return { error: 'parse', data: null };
  }
};

const verifyJWTSignature = async (token) => {
  try {
    const publicKey = await importSPKI(LICENSE_PUBLIC_KEY, 'ES256');
    await jwtVerify(token, publicKey);
    return true;
  } catch (error) {
    console.error('Failed to verify JWT signature:', error);
    return false;
  }
};

const ConfigGeneral = ({ formik, config, status, isAdmin }) => {
  const { t } = useTranslation();
  const [isCheckingUpdate, setIsCheckingUpdate] = React.useState(false);
  const [licenseValidation, setLicenseValidation] = React.useState({ checked: false, valid: null });

  return (
    <Stack spacing={3}>
      <MainCard title={t('mgmt.config.generalTitle')}>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Alert severity="info">{t('mgmt.config.general.configFileInfo')}</Alert>
          </Grid>

          {status && !status.containerized && <>
            <Grid item xs={3}>
              <LoadingButton loading={isCheckingUpdate} variant="outlined" color="primary" onClick={() => {
                setIsCheckingUpdate(true);
                API.forceAutoUpdate().then(() => {
                  setIsCheckingUpdate(false);
                })
              }}>{t('mgmt.config.general.forceAutoUpdateButton')}</LoadingButton>
            </Grid>

            <CosmosCheckbox
              label={t('mgmt.config.general.autoupdates')}
              name="AutoUpdate"
              formik={formik}
              helperText={t('mgmt.config.general.autoupdates')}
            />

            <CosmosCheckbox
              label={t('mgmt.config.general.betaupdate')}
              name="BetaUpdates"
              formik={formik}
              helperText={t('mgmt.config.general.betaupdate')}
            /></>}

          <CosmosCheckbox
            label={t('mgmt.config.general.forceMfaCheckbox.forceMfaLabel')}
            name="RequireMFA"
            formik={formik}
            helperText={t('mgmt.config.general.forceMfaCheckbox.forceMfaHelperText')}
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
        </Grid>
      </MainCard>

      <MainCard title={t('mgmt.config.general.backupTitle')}>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Alert severity="info">{t('mgmt.config.general.backupInfo2')}</Alert>
          </Grid>
          <Grid item xs={12}>
            <Stack direction={"row"} spacing={2} alignItems="flex-end">
                <FilePickerButton canCreate onPick={(path) => {
                  if(path)
                      formik.setFieldValue('BackupOutputDir', path);
                }} size="150%" select="folder" />
              <CosmosInputText
                label={t('mgmt.config.general.backupDirInput.backupDirLabel')}
                name="BackupOutputDir"
                formik={formik}
                helperText={t('mgmt.config.general.backupDirInput.backupDirHelperText')}
              />
            </Stack>
          </Grid>

          <Grid item xs={12}>
            <Stack direction={"row"} spacing={2} alignItems="flex-end">
                {status && status.Licence && <FilePickerButton canCreate onPick={(path) => {
                  if(path)
                      formik.setFieldValue('IncrBackupOutputDir', path);
                }} size="150%" select="folder" />}
              <CosmosInputText
                label={t('mgmt.config.general.backupDirInput.incrBackupDirLabel')}
                name="IncrBackupOutputDir"
                disabled={!status || !status.Licence}
                formik={formik}
                helperText={t('mgmt.config.general.backupDirInput.backupDirHelperText')}
              />
            </Stack>
          </Grid>
        </Grid>
      </MainCard>

      <MainCard title={t('mgmt.config.general.licenceInput.licenceLabel')}>
        <Grid container spacing={3}>
          <CosmosInputPassword
            noStrength
            label={t('mgmt.config.general.licenceInput.licenceLabel')}
            name="Licence"
            formik={formik}
            helperText="Licence Key"
            autoComplete={'licence-key'}
            onChange={(e) => {
              formik.setFieldValue("ServerToken", "");
            }}
          />
          {(() => {
            const licenseResult = React.useMemo(() => {
              if (!formik.values.Licence) return null;
              return validateAndDecodeJWT(formik.values.Licence);
            }, [formik.values.Licence]);

            React.useEffect(() => {
              if (licenseResult && !licenseResult.error) {
                setLicenseValidation({ checked: false, valid: null });
                verifyJWTSignature(formik.values.Licence).then((isValid) => {
                  setLicenseValidation({ checked: true, valid: isValid });
                });
              }
            }, [formik.values.Licence]);

            if (!licenseResult) return null;

            if (licenseResult.error === 'parse') {
              return (
                <Grid item xs={12}>
                  <Alert severity="error">
                    {t('mgmt.config.general.licenceInput.licenseInvalid')}
                  </Alert>
                </Grid>
              );
            }

            if (licenseValidation.checked && !licenseValidation.valid) {
              return (
                <Grid item xs={12}>
                  <Alert severity="error">
                    {t('mgmt.config.general.licenceInput.licenseSignatureInvalid')}
                  </Alert>
                </Grid>
              );
            }

            const licenseData = licenseResult.data;
            const isExpired = licenseData.lifetime === false &&
              licenseData.expiresAt &&
              new Date(licenseData.expiresAt * 1000) < new Date();

            return (
              <>
                {licenseData.email && (
                  <Grid item xs={12}>
                    <Alert severity="info">
                      {t('mgmt.config.general.licenceInput.licenseEmail', { email: licenseData.email })} <br/>
                      {t('mgmt.config.general.licenceInput.licenseType', { type: licenseData.subscriptionType })}
                    </Alert>
                  </Grid>
                )}

                {licenseData.lifetime === false && isExpired && (
                  <Grid item xs={12}>
                    <Alert severity="warning" icon={<WarningFilled />}>
                      {t('mgmt.config.general.licenceInput.licenseExpired', {
                        date: new Date(licenseData.expiresAt * 1000).toLocaleDateString()
                      })}
                    </Alert>
                  </Grid>
                )}

                {licenseData.lifetime === false && (
                  <Grid item xs={12}>
                    <Stack spacing={1}>
                      <Button
                        href='https://billing.stripe.com/p/login/28obMlc7X1jgeqY144'
                        target="_blank"
                        variant="outlined">
                        {t('mgmt.config.general.licenceInput.manageLicenceButton')}
                      </Button>
                    </Stack>
                  </Grid>
                )}
              </>
            );
          })()}
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
    </Stack>
  );
};

export default ConfigGeneral;
