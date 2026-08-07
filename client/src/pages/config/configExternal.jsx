import * as React from 'react';
import MainCard from '../../components/MainCard';
import {
  Grid,
  InputLabel,
  OutlinedInput,
  Stack,
  FormHelperText,
} from '@mui/material';
import { CosmosCheckbox, CosmosCollapse, CosmosInputPassword, CosmosInputText } from './users/formShortcuts';
import { useTranslation } from 'react-i18next';
import { FilePickerButton } from '../../components/filePicker';

const ConfigExternal = ({ formik }) => {
  const { t } = useTranslation();

  return (
    <Stack spacing={3}>
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

          <Stack direction={"row"} spacing={2} alignItems="flex-end">
            <FilePickerButton onPick={(path) => {
              if(path)
                formik.setFieldValue('DefaultDataPath', path);
            }} size="150%" select="folder" />
            <CosmosInputText
              label={t('mgmt.config.docker.defaultDatapathInput.defaultDatapathLabel')}
              name="DefaultDataPath"
              formik={formik}
              placeholder={'/cosmos-storage'}
            />
          </Stack>
        </Stack>
      </MainCard>

      <MainCard title="MongoDB">
        <Grid container spacing={3}>
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
        </Grid>
      </MainCard>
    </Stack>
  );
};

export default ConfigExternal;
