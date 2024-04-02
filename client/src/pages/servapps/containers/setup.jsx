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
  const [pullRequest, setPullRequest] = useState(null);
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("sm"));
  const padding = isMobile ? "6px 4px" : "12px 10px";
  const [latestImage, setLatestImage] = useState(containerInfo.Config.Image);

  const wrapCard = (children) => {
    if (noCard) return children;
    return <MainCard title="Docker Container Setup">{children}</MainCard>;
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
        errors.image = "Required";
      }
      if (!values.name && newContainer) {
        errors.name = "Required";
      }
      // env keys and labels key mustbe unique
      const envKeys = values.envVars.map((envVar) => envVar.key);
      const labelKeys = values.labels.map((label) => label.key);
      const uniqueEnvKeysKeys = [...new Set(envKeys)];
      const uniqueLabelKeys = [...new Set(labelKeys)];
      if (uniqueEnvKeysKeys.length !== envKeys.length) {
        errors.submit = "Environment Variables must be unique";
      }
      if (uniqueLabelKeys.length !== labelKeys.length) {
        errors.submit = "Labels must be unique";
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
                title="Pulling New Image..."
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
                        This container is not running. Editing any settings will
                        cause the container to start again.
                      </Alert>
                    )}
                  <Grid container spacing={4}>
                    {!installer && (
                      <>
                        {newContainer && (
                          <CosmosInputText
                            name="name"
                            label="Name"
                            placeholder="Name"
                            formik={formik}
                          />
                        )}
                        <CosmosInputText
                          name="image"
                          label="Image"
                          placeholder="Image"
                          formik={formik}
                        />
                        <CosmosSelect
                          name="restartPolicy"
                          label="Restart Policy"
                          placeholder="Restart Policy"
                          options={restartPolicies}
                          formik={formik}
                        />
                        <CosmosInputText
                          name="user"
                          label="User"
                          placeholder="User"
                          formik={formik}
                        />
                        <CosmosCheckbox
                          name="interactive"
                          label="Interactive Mode"
                          formik={formik}
                        />
                        {OnForceSecure && (
                          <Grid item xs={12}>
                            <Checkbox
                              type="checkbox"
                              as={FormControlLabel}
                              control={<Checkbox size="large" />}
                              label={"Force secure container"}
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

                    <CosmosFormDivider title={"Environment Variables"} />
                    <Grid item xs={12}>
                      {formik.values.envVars.map((envVar, idx) => (
                        <Grid container key={idx}>
                          <Grid item xs={5} style={{ padding }}>
                            <TextField
                              name={`envVars.${idx}.key`}
                              onChange={formik.handleChange}
                              label="Key"
                              fullWidth
                              value={envVar.key}
                            />
                          </Grid>
                          <Grid item xs={6} style={{ padding }}>
                            <TextField
                              name={`envVars.${idx}.value`}
                              onChange={formik.handleChange}
                              fullWidth
                              label="Value"
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
                        Add
                      </ResponsiveButton>
                    </Grid>

                    <CosmosFormDivider title={"Labels"} />
                    <Grid item xs={12}>
                      {formik.values.labels.map((label, idx) => (
                        <Grid container key={idx}>
                          <Grid item xs={5} style={{ padding }}>
                            <TextField
                              name={`labels.${idx}.key`}
                              fullWidth
                              label="Key"
                              value={label.key}
                              onChange={formik.handleChange}
                            />
                          </Grid>
                          <Grid item xs={6} style={{ padding }}>
                            <TextField
                              name={`labels.${idx}.value`}
                              label="Value"
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
                        Add
                      </ResponsiveButton>
                    </Grid>

                    <CosmosFormDivider title={"Devices"} />
                    <Grid item xs={12}>
                      {formik.values.devices.map((device, idx) => (
                        <Grid container key={idx}>
                          <Grid item xs={5} style={{ padding }}>
                            <TextField
                              name={`devices.${idx}.key`}
                              fullWidth
                              label="Host Path"
                              value={device.key}
                              onChange={formik.handleChange}
                            />
                          </Grid>
                          <Grid item xs={6} style={{ padding }}>
                            <TextField
                              name={`devices.${idx}.value`}
                              label="Container Path"
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
                        Add
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
                        You have updated the image. Clicking the button below
                        will pull the new image, and then only can you update
                        the container.
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
                        ? "Pull New Image"
                        : "Update Container"}
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
