import React, { useCallback, useMemo, useState } from "react";
import { Field, Formik } from "formik";
import {
  Button,
  Stack,
  Grid,
  MenuItem,
  TextField,
  IconButton,
  FormHelperText,
  useMediaQuery,
  useTheme,
  Alert,
  FormControlLabel,
  Checkbox,
} from "@mui/material";
import MainCard from "../../../components/MainCard";
import {
  CosmosCheckbox,
  CosmosFormDivider,
  CosmosInputText,
  CosmosSelect,
} from "../../config/users/formShortcuts";
import { DeleteOutlined, PlusCircleOutlined } from "@ant-design/icons";
import * as API from "../../../api";
import { LoadingButton } from "@mui/lab";
import LogsInModal from "../../../components/logsInModal";
import ResponsiveButton from "../../../components/responseiveButton";
import { useTranslation } from 'react-i18next';

const containerInfoFrom = (values) => {
  const labels = {};
  values.labels.forEach((label) => {
    labels[label.key] = label.value;
  });

  const devices = values.devices.map((device) => {
    return `${device.key}:${device.value}`;
  });

  const envVars = values.envVars.map((envVar) => {
    return `${envVar.key}=${envVar.value}`;
  });

  const realvalues = {
    ...values,
    envVars: envVars,
    labels: labels,
    devices: devices,
  };

  realvalues.interactive = realvalues.interactive ? 2 : 1;

  return realvalues;
};

const restartPolicies = [
  ["no", "No Restart"],
  ["always", "Always Restart"],
  ["on-failure", "Restart On Failure"],
  ["unless-stopped", "Restart Unless Stopped"],
];

const DockerContainerSetup = ({
  noCard,
  containerInfo,
  installer,
  OnChange,
  refresh,
  newContainer,
  OnForceSecure,
}) => {
  const { t } = useTranslation();
  const [pullRequest, setPullRequest] = useState(null);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("sm"));
  const padding = isMobile ? "6px 4px" : "12px 10px";
  const [latestImage, setLatestImage] = useState(containerInfo.Config.Image);

  const wrapCard = (children) => {
    if (noCard) return children;
    return <MainCard title={t('mgmt.servapps.newContainerTitle')}>{children}</MainCard>;
  };

  const initialValues = useMemo(() => {
    return {
      name: containerInfo.Name.replace("/", ""),
      image: containerInfo.Config.Image,
      restartPolicy: containerInfo.HostConfig.RestartPolicy.Name,
      user: containerInfo.Config.User,
      envVars: containerInfo.Config.Env.map((envVar) => {
        const [key, value] = envVar.split(/=(.*)/s);
        return { key, value };
      }),
      labels: Object.keys(containerInfo.Config.Labels).map((key) => {
        return { key, value: containerInfo.Config.Labels[key] };
      }),
      devices: containerInfo.HostConfig.Devices
        ? containerInfo.HostConfig.Devices.map((device) => {
            return typeof device == "string"
              ? {
                  key: device.split(":")[0],
                  value: device.split(":")[1] || device.split(":")[0],
                }
              : { key: device.PathOnHost, value: device.PathInContainer };
          })
        : [],
      interactive: containerInfo.Config.Tty && containerInfo.Config.OpenStdin,
    };
  }, [
    containerInfo.Config.Env,
    containerInfo.Config.Image,
    containerInfo.Config.Labels,
    containerInfo.Config.OpenStdin,
    containerInfo.Config.Tty,
    containerInfo.Config.User,
    containerInfo.HostConfig.Devices,
    containerInfo.HostConfig.RestartPolicy.Name,
    containerInfo.Name,
  ]);

  const onValidate = useCallback(
    (values) => {
      const errors = {};
      if (!values.image) {
        errors.image = t('global.required');
      }
      if (!values.name && newContainer) {
        errors.name = t('global.required');
      }
      // env keys and labels key mustbe unique
      const envKeys = values.envVars.map((envVar) => envVar.key);
      const labelKeys = values.labels.map((label) => label.key);
      const uniqueEnvKeysKeys = [...new Set(envKeys)];
      const uniqueLabelKeys = [...new Set(labelKeys)];
      if (uniqueEnvKeysKeys.length !== envKeys.length) {
        errors.submit = t('mgmt.servapps.newContainer.env.keyNotUniqueError');
      }
      if (uniqueLabelKeys.length !== labelKeys.length) {
        errors.submit = t('mgmt.servapps.newContainer.label.labelNotUniqueError');
      }
      OnChange && OnChange(containerInfoFrom(values));
      return errors;
    },
    [OnChange, newContainer]
  );

  const onSubmit = useCallback(
    async (values, { setErrors, setStatus, setSubmitting }) => {
      if (values.image !== latestImage) {
        setPullRequest(
          () => (cb) => API.docker.pullImage(values.image, cb, true)
        );
        return;
      }

      if (newContainer) return false;
      delete values.name;

      setSubmitting(true);

      let realvalues = containerInfoFrom(values);

      return API.docker
        .updateContainer(containerInfo.Name.replace("/", ""), realvalues)
        .then((res) => {
          setStatus({ success: true });
          setSubmitting(false);
          refresh && refresh();
        })
        .catch((err) => {
          setStatus({ success: false });
          setErrors({ submit: err.message });
          setSubmitting(false);
        });
    },
    [containerInfo.Name, latestImage, newContainer, refresh]
  );

  return (
    <div
      style={{
        maxWidth: "1000px",
        width: "100%",
        margin: "",
        position: "relative",
      }}
    >
      <Formik
        initialValues={initialValues}
        enableReinitialize
        validate={onValidate}
        onSubmit={onSubmit}
      >
        {(formik) => (
          <form noValidate onSubmit={formik.handleSubmit}>
            {pullRequest && (
              <LogsInModal
                request={pullRequest}
                title={t('mgmt.servapps.newContainer.pullingImageStatus')}
                OnSuccess={() => {
                  setPullRequest(null);
                  setLatestImage(formik.values.image);
                }}
                OnClose={() => {
                  setPullRequest(null);
                }}
              />
            )}
            <Stack spacing={2}>
              {wrapCard(
                <>
                  {containerInfo.State &&
                    containerInfo.State.Status !== "running" && (
                      <Alert
                        severity="warning"
                        style={{ marginBottom: "15px" }}
                      >
                        {t('mgmt.servApps.volumes.containerNotRunningWarning')}
                      </Alert>
                    )}
                  <Grid container spacing={4}>
                    {!installer && (
                      <>
                        {newContainer && (
                          <CosmosInputText
                            name="name"
                            label={t('global.nameTitle')}
                            placeholder={t('global.nameTitle')}
                            formik={formik}
                          />
                        )}
                        <CosmosInputText
                          name="image"
                          label={t('mgmt.servApps.container.overview.imageTitle')}
                          placeholder={t('mgmt.servApps.container.overview.imageTitle')}
                          formik={formik}
                        />
                        <CosmosSelect
                          name="restartPolicy"
                          label={t('mgmt.servapps.newContainer.restartPolicyInput.restartPolicyLabel')}
                          placeholder={t('mgmt.servapps.newContainer.restartPolicyInput.restartPolicyPlaceholder')}
                          options={restartPolicies}
                          formik={formik}
                        />
                        <CosmosInputText
                          name="user"
                          label={t('global.user')}
                          placeholder={t('global.user')}
                          formik={formik}
                        />
                        <CosmosCheckbox
                          name="interactive"
                          label={t('mgmt.servapps.newContainer.interactiveCheckbox.interactiveLabel')}
                          formik={formik}
                        />
                        {OnForceSecure && (
                          <Grid item xs={12}>
                            <Checkbox
                              type="checkbox"
                              as={FormControlLabel}
                              control={<Checkbox size="large" />}
                              label={t('mgmt.servapps.newContainer.forceSecureCheckbox.forceSecureLabel')}
                              checked={
                                containerInfo.Config.Labels.hasOwnProperty(
                                  "cosmos-force-network-secured"
                                ) &&
                                containerInfo.Config.Labels[
                                  "cosmos-force-network-secured"
                                ] === "true"
                              }
                              onChange={(e) => {
                                OnForceSecure(e.target.checked);
                              }}
                            />
                          </Grid>
                        )}
                      </>
                    )}

                    <CosmosFormDivider title={t('mgmt.servapps.newContainer.envTitle')} />
                    <Grid item xs={12}>
                      {formik.values.envVars.map((envVar, idx) => (
                        <Grid container key={idx}>
                          <Grid item xs={5} style={{ padding }}>
                            <TextField
                              name={`envVars.${idx}.key`}
                              onChange={formik.handleChange}
                              label={t('mgmt.servapps.newContainer.env.envKeyInput.envKeyLabel')}
                              fullWidth
                              value={envVar.key}
                            />
                          </Grid>
                          <Grid item xs={6} style={{ padding }}>
                            <TextField
                              name={`envVars.${idx}.value`}
                              onChange={formik.handleChange}
                              fullWidth
                              label={t('mgmt.servapps.newContainer.env.envValueInput.envValueLabel')}
                              value={envVar.value}
                            />
                          </Grid>
                          <Grid item xs={1} style={{ padding }}>
                            <IconButton
                              fullWidth
                              variant="outlined"
                              color="error"
                              onClick={() => {
                                const newEnvVars = [...formik.values.envVars];
                                newEnvVars.splice(idx, 1);
                                formik.setFieldValue("envVars", newEnvVars);
                              }}
                            >
                              <DeleteOutlined />
                            </IconButton>
                          </Grid>
                        </Grid>
                      ))}

                      <ResponsiveButton
                        variant="outlined"
                        color="primary"
                        size="large"
                        onClick={() => {
                          const newEnvVars = [...formik.values.envVars];
                          newEnvVars.push({ key: "", value: "" });
                          formik.setFieldValue("envVars", newEnvVars);
                        }}
                        startIcon={<PlusCircleOutlined />}
                      >
                        {t('global.addAction')}
                      </ResponsiveButton>
                    </Grid>

                    <CosmosFormDivider title={t('mgmt.servapps.newContainer.labelsTitle')} />
                    <Grid item xs={12}>
                      {formik.values.labels.map((label, idx) => (
                        <Grid container key={idx}>
                          <Grid item xs={5} style={{ padding }}>
                            <TextField
                              name={`labels.${idx}.key`}
                              fullWidth
                              label={t('mgmt.servapps.newContainer.env.envKeyInput.envKeyLabel')}
                              value={label.key}
                              onChange={formik.handleChange}
                            />
                          </Grid>
                          <Grid item xs={6} style={{ padding }}>
                            <TextField
                              name={`labels.${idx}.value`}
                              label={t('mgmt.servapps.newContainer.env.envValueInput.envValueLabel')}
                              fullWidth
                              value={label.value}
                              onChange={formik.handleChange}
                            />
                          </Grid>
                          <Grid item xs={1} style={{ padding }}>
                            <IconButton
                              fullWidth
                              variant="outlined"
                              color="error"
                              onClick={() => {
                                const newLabels = [...formik.values.labels];
                                newLabels.splice(idx, 1);
                                formik.setFieldValue("labels", newLabels);
                              }}
                            >
                              <DeleteOutlined />
                            </IconButton>
                          </Grid>
                        </Grid>
                      ))}
                      <ResponsiveButton
                        variant="outlined"
                        color="primary"
                        size="large"
                        onClick={() => {
                          const newLabels = [...formik.values.labels];
                          newLabels.push({ key: "", value: "" });
                          formik.setFieldValue("labels", newLabels);
                        }}
                        startIcon={<PlusCircleOutlined />}
                      >
                        {t('global.addAction')}
                      </ResponsiveButton>
                    </Grid>

                    <CosmosFormDivider title={t('mgmt.servapps.newContainer.devicesTitle')} />
                    <Grid item xs={12}>
                      {formik.values.devices.map((device, idx) => (
                        <Grid container key={idx}>
                          <Grid item xs={5} style={{ padding }}>
                            <TextField
                              name={`devices.${idx}.key`}
                              fullWidth
                              label={t('mgmt.servapps.newContainer.devices.hostPathInput.hostPathLabel')}
                              value={device.key}
                              onChange={formik.handleChange}
                            />
                          </Grid>
                          <Grid item xs={6} style={{ padding }}>
                            <TextField
                              name={`devices.${idx}.value`}
                              label={t('mgmt.servapps.newContainer.devices.containerPathInput.containerPathLabel')}
                              fullWidth
                              value={device.value}
                              onChange={formik.handleChange}
                            />
                          </Grid>
                          <Grid item xs={1} style={{ padding }}>
                            <IconButton
                              fullWidth
                              variant="outlined"
                              color="error"
                              onClick={() => {
                                const newDevices = [...formik.values.devices];
                                newDevices.splice(idx, 1);
                                formik.setFieldValue("devices", newDevices);
                              }}
                            >
                              <DeleteOutlined />
                            </IconButton>
                          </Grid>
                        </Grid>
                      ))}
                      <ResponsiveButton
                        variant="outlined"
                        color="primary"
                        size="large"
                        onClick={() => {
                          const newDevices = [...formik.values.devices];
                          newDevices.push({ key: "", value: "" });
                          formik.setFieldValue("devices", newDevices);
                        }}
                        startIcon={<PlusCircleOutlined />}
                      >
                        {t('global.addAction')}
                      </ResponsiveButton>
                    </Grid>
                  </Grid>
                </>
              )}
              {!newContainer && (
                <MainCard>
                  <Stack direction="column" spacing={2}>
                    {formik.errors.submit && (
                      <Grid item xs={12}>
                        <FormHelperText error>
                          {formik.errors.submit}
                        </FormHelperText>
                      </Grid>
                    )}
                    {formik.values.image !== latestImage && (
                      <Alert
                        severity="warning"
                        style={{ marginBottom: "15px" }}
                      >
                        {t('mgmt.servapps.newContainer.imageUpdateWarning')}
                      </Alert>
                    )}
                    <LoadingButton
                      fullWidth
                      disableElevation
                      disabled={formik.errors.submit}
                      loading={formik.isSubmitting}
                      size="large"
                      type="submit"
                      variant="contained"
                      color="primary"
                    >
                      {formik.values.image !== latestImage
                        ? t('mgmt.servapps.newContainer.pullImageButton')
                        : t('mgmt.servapps.newContainer.updateContainerButton')}
                    </LoadingButton>
                  </Stack>
                </MainCard>
              )}
            </Stack>
          </form>
        )}
      </Formik>
    </div>
  );
};

export default DockerContainerSetup;
