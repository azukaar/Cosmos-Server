import React, { useCallback, useEffect, useMemo } from "react";
import { Formik } from "formik";
import {
  Button,
  Stack,
  Grid,
  MenuItem,
  TextField,
  IconButton,
  FormHelperText,
  CircularProgress,
  useTheme,
  Checkbox,
  Alert,
} from "@mui/material";
import MainCard from "../../../components/MainCard";
import { PlusCircleOutlined } from "@ant-design/icons";
import * as API from "../../../api";
import { LoadingButton } from "@mui/lab";
import PrettyTableView from "../../../components/tableView/prettyTableView";
import ResponsiveButton from "../../../components/responseiveButton";
import { useTranslation } from 'react-i18next';
import { FilePickerButton } from "../../../components/filePicker";
import { Backups } from "../../backups/backups";
import BackupDialog from "../../backups/backupDialog";

const VolumeContainerSetup = ({
  noCard,
  containerInfo,
  frozenVolumes = [],
  refresh,
  newContainer,
  OnChange,
}) => {
  const { t } = useTranslation();
  const [volumes, setVolumes] = React.useState([]);
  const theme = useTheme();

  useEffect(() => {
    API.docker.volumeList().then((res) => {
      setVolumes(res.data.Volumes);
    });
  }, []);

  const refreshAll = () => {
    setVolumes(null);
    if (refresh)
      refresh().then(() => {
        API.docker.volumeList().then((res) => {
          setVolumes(res.data.Volumes);
        });
      });
    else
      API.docker.volumeList().then((res) => {
        setVolumes(res.data.Volumes);
      });
  };

  const formatSource = (mount) => {
    if (mount.startsWith("/")) return mount;
    else return "/var/lib/docker/volumes/" + mount + "/_data";
  }

  const wrapCard = (children) => {
    if (noCard) return children;
    return <MainCard title={t('mgmt.servapps.newContainer.volumesTitle')}>{children}</MainCard>;
  };

  const initialValues = useMemo(() => {
    return {
      volumes: [
        ...(containerInfo.HostConfig.Mounts || []),
        ...(containerInfo.HostConfig.Binds || []).map((bind) => {
          const [source, destination, mode] = bind.split(":");
          return {
            Type: "bind",
            Source: source,
            Target: destination,
          };
        }),
      ],
    };
  }, [containerInfo.HostConfig.Binds, containerInfo.HostConfig.Mounts]);

  const onValidate = useCallback(
    (values) => {
      const errors = {};
      // check unique
      const volumes = values.volumes.map((volume) => {
        return `${volume.Target}`;
      });
      const unique = [...new Set(volumes)];
      if (unique.length !== volumes.length) {
        errors.submit = t('mgmt.servapps.newContainer.volumes.mountNotUniqueError');
      }
      OnChange && OnChange(values, volumes);
      return errors;
    },
    [OnChange]
  );

  const onSubmit = useCallback(
    async (values, { setErrors, setStatus, setSubmitting }) => {
      if (newContainer) return;
      setSubmitting(true);
      const realvalues = {
        Volumes: values.volumes.map((volume) => ({
          Type: volume.Type,
          Source: volume.Source,
          Target: volume.Target,
          ReadOnly: false, // TODO: add support for this
        })),
      };
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
          refresh && refresh();
        });
    },
    [containerInfo.Name, newContainer, refresh]
  );

  return (
    <Stack spacing={2}>
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
              {wrapCard(
                <>
                  {!newContainer &&
                    containerInfo.State &&
                    containerInfo.State.Status !== "running" && (
                      <Alert
                        severity="warning"
                        style={{ marginBottom: "15px" }}
                      >
                        {t('mgmt.servApps.volumes.containerNotRunningWarning')}
                      </Alert>
                    )}
                  <Grid container spacing={4}>
                    <Grid item xs={12}>
                      {volumes && (
                        <PrettyTableView
                          data={formik.values.volumes}
                          onRowClick={() => {}}
                          getKey={(r) => r.Id}
                          fullWidth
                          buttons={[
                            <ResponsiveButton
                              startIcon={<PlusCircleOutlined />}
                              variant="outlined"
                              color="primary"
                              onClick={() => {
                                formik.setFieldValue("volumes", [
                                  ...formik.values.volumes,
                                  {
                                    Type: "volume",
                                    Name: "",
                                    Driver: "local",
                                    Source: "",
                                    Destination: "",
                                    RW: true,
                                  },
                                ]);
                              }}
                            >
                              {t('mgmt.servapps.newContainer.volumes.newMountButton')}
                            </ResponsiveButton>,
                          ]}
                          columns={[
                            {
                              title: "Type",
                              field: (r, k) => (
                                <div
                                  style={{
                                    fontWeight: "bold",
                                    wordSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                    maxWidth: "100px",
                                  }}
                                >
                                  <TextField
                                    className="px-2 my-2"
                                    disabled={frozenVolumes.includes(r.Source)}
                                    variant="outlined"
                                    id="Type"
                                    select
                                    value={r.Type}
                                    name={`volumes[${k}].Type`}
                                    onChange={formik.handleChange}
                                  >
                                    <MenuItem value="bind">{t('mgmt.servapps.newContainer.volumes.bindInput')}</MenuItem>
                                    <MenuItem value="volume">{t('global.volume')}</MenuItem>
                                  </TextField>
                                </div>
                              ),
                            },
                            {
                              title: t('global.source'),
                              field: (r, k) => (
                                <div
                                  style={{
                                    fontWeight: "bold",
                                    wordSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                    maxWidth: "300px",
                                  }}
                                >
                                  {r.Type == "bind" ? (
                                    <Stack direction={"row"} spacing={2}>
                                    <FilePickerButton onPick={(path) => {
                                      if(path)
                                        formik.setFieldValue(`volumes[${k}].Source`, path);
                                    }} size="150%" select="folder" />
                                    <TextField
                                      className="px-2 my-2"
                                      variant="outlined"
                                      name={`volumes[${k}].Source`}
                                      id="Source"
                                      disabled={frozenVolumes.includes(
                                        r.Source
                                      )}
                                      style={{ minWidth: "200px" }}
                                      value={r.Source}
                                      onChange={formik.handleChange}
                                    />
                                    </Stack>
                                  ) : (
                                    <TextField
                                      className="px-2 my-2"
                                      variant="outlined"
                                      name={`volumes[${k}].Source`}
                                      id="Source"
                                      disabled={frozenVolumes.includes(
                                        r.Source
                                      )}
                                      select
                                      style={{ minWidth: "200px" }}
                                      value={r.Source}
                                      onChange={formik.handleChange}
                                    >
                                      {[...volumes, r].map((volume) => (
                                        <MenuItem
                                          key={volume.Id || "last"}
                                          value={volume.Name || volume.Source}
                                        >
                                          {volume.Name || volume.Source}
                                        </MenuItem>
                                      ))}
                                    </TextField>
                                  )}
                                </div>
                              ),
                            },
                            {
                              title: t('global.target'),
                              field: (r, k) => (
                                <div
                                  style={{
                                    fontWeight: "bold",
                                    wordSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                    maxWidth: "300px",
                                  }}
                                >
                                  <TextField
                                    className="px-2 my-2"
                                    variant="outlined"
                                    name={`volumes[${k}].Target`}
                                    id="Target"
                                    disabled={frozenVolumes.includes(r.Source)}
                                    style={{ minWidth: "200px" }}
                                    value={r.Target}
                                    onChange={formik.handleChange}
                                  />
                                </div>
                              ),
                            },
                            {
                              title: "",
                              field: (r) => {
                                return (
                                  <Stack direction="row" spacing={2}>
                                    <Button
                                    variant="outlined"
                                    color="primary"
                                    disabled={frozenVolumes.includes(r.Source)}
                                    onClick={() => {
                                      const newVolumes = [
                                        ...formik.values.volumes,
                                      ];
                                      newVolumes.splice(
                                        newVolumes.indexOf(r),
                                        1
                                      );
                                      formik.setFieldValue(
                                        "volumes",
                                        newVolumes
                                      );
                                    }}
                                  >
                                    {t('global.unmount')}
                                  </Button>
                                  {r.Target ? <BackupDialog preName={`${containerInfo.Name.replace("/", "").replace("/", "-")}-${r.Target.replace("/", "").replaceAll("/", "_")}`} preSource={formatSource(r.Source)} refresh={() => setTimeout(refreshAll, 1500)} /> : null}
                                  </Stack>
                                );
                              },
                            },
                          ]}
                        />
                      )}
                      {!volumes && (
                        <div style={{ height: "100px" }}>
                          <center>
                            <br />
                            <CircularProgress />
                          </center>
                        </div>
                      )}
                    </Grid>
                    <Grid item xs={12}>
                      <Stack direction="column" spacing={2}>
                        {formik.errors.submit && (
                          <Grid item xs={12}>
                            <FormHelperText error>
                              {formik.errors.submit}
                            </FormHelperText>
                          </Grid>
                        )}
                        {!newContainer && (
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
                            {t('mgmt.servapps.newContainer.volumes.updateVolumesButton')}
                          </LoadingButton>
                        )}
                      </Stack>
                    </Grid>
                  </Grid>
                </>
              )}
            </form>
          )}
        </Formik>
      </div>
      <div
        style={{
          maxWidth: "1000px",
          width: "100%",
          margin: "",
          position: "relative",
      }}>
        <MainCard title={t('mgmt.backup.backups')}>
          <Backups pathFilters={
            containerInfo.HostConfig.Mounts.map((r) => formatSource(r.Source))
          } />
        </MainCard>
      </div>
    </Stack>
  );
};

export default VolumeContainerSetup;
