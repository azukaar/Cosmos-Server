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
import PermissionGuard from '../../../components/permissionGuard';
import { PERM_RESOURCES } from '../../../utils/permissions';

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
    if (!mount) return null;
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
        ...(containerInfo.HostConfig.Mounts || []).map((m) => ({
          type: m.Type || m.type,
          source: m.Source || m.source,
          target: m.Target || m.target || m.Destination || m.destination,
          subpath: (m.VolumeOptions && m.VolumeOptions.Subpath) || m.SubPath || m.Subpath || m.subpath || "",
          readOnly: m.ReadOnly || m.readOnly || false,
          noCopy: (m.VolumeOptions && m.VolumeOptions.NoCopy) || m.NoCopy || m.noCopy || m.nocopy || false,
        })),
        ...(containerInfo.HostConfig.Binds || []).map((bind) => {
          const [source, destination, mode] = bind.split(":");
          return {
            type: "bind",
            source: source,
            target: destination,
            subpath: "",
            readOnly: mode && (mode.split(",").includes("ro") || mode.split(",").includes("readonly")),
            noCopy: false,
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
        return `${volume.target}`;
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
        volumes: values.volumes.map((volume) => ({
          type: volume.type,
          source: volume.source,
          target: volume.target,
          subpath: volume.subpath || "",
          readOnly: !!volume.readOnly,
          noCopy: !!volume.noCopy,
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
    <Stack spacing={2} style={{ alignItems: "center" }}>
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
                  <Grid container spacing={4}>
                    <Grid item xs={12}>
                      {volumes && (
                        <PrettyTableView
                          data={formik.values.volumes}
                          onRowClick={() => {}}
                          getKey={(r) => r.Id}
                          fullWidth
                          buttons={[
                            <PermissionGuard permission={PERM_RESOURCES}><ResponsiveButton
                              startIcon={<PlusCircleOutlined />}
                              variant="outlined"
                              color="primary"
                              onClick={() => {
                                formik.setFieldValue("volumes", [
                                  ...formik.values.volumes,
                                  {
                                    type: "volume",
                                    name: "",
                                    driver: "local",
                                    source: "",
                                    destination: "",
                                    target: "",
                                    subpath: "",
                                    readOnly: false,
                                                            rw: true,
                                  },
                                ]);
                              }}
                            >
                              {t('mgmt.servapps.newContainer.volumes.newMountButton')}
                            </ResponsiveButton></PermissionGuard>,
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
                                    disabled={frozenVolumes.includes(r.source)}
                                    variant="outlined"
                                    id="Type"
                                    select
                                    value={r.type}
                                    name={`volumes[${k}].type`}
                                    onChange={formik.handleChange}
                                  >
                                    <MenuItem value="bind">{t('mgmt.servapps.newContainer.volumes.bindInput')}</MenuItem>
                                    <MenuItem value="volume">{t('global.volume')}</MenuItem>
                                    <MenuItem value="tmpfs">tmpfs</MenuItem>
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
                                  {r.type == "bind" ? (
                                    <Stack direction={"row"} spacing={2}>
                                    <FilePickerButton onPick={(path) => {
                                      if(path)
                                        formik.setFieldValue(`volumes[${k}].source`, path);
                                    }} size="150%" select="any" />
                                    <TextField
                                      className="px-2 my-2"
                                      variant="outlined"
                                      name={`volumes[${k}].source`}
                                      id="Source"
                                      disabled={frozenVolumes.includes(
                                        r.source
                                      )}
                                      style={{ minWidth: "200px" }}
                                      value={r.source}
                                      onChange={formik.handleChange}
                                    />
                                    </Stack>
                                  ) : r.type == "tmpfs" ? (
                                    <TextField
                                      className="px-2 my-2"
                                      variant="outlined"
                                      disabled
                                      style={{ minWidth: "200px" }}
                                      value="(memory)"
                                    />
                                  ) : (
                                    <TextField
                                      className="px-2 my-2"
                                      variant="outlined"
                                      name={`volumes[${k}].source`}
                                      id="Source"
                                      disabled={frozenVolumes.includes(
                                        r.source
                                      )}
                                      select
                                      style={{ minWidth: "200px" }}
                                      value={r.source}
                                      onChange={formik.handleChange}
                                    >
                                      {[...volumes, r].map((volume) => (
                                        <MenuItem
                                          key={volume.Id || "last"}
                                          value={volume.Name || volume.source}
                                        >
                                          {volume.Name || volume.source}
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
                                    name={`volumes[${k}].target`}
                                    id="Target"
                                    disabled={frozenVolumes.includes(r.source)}
                                    style={{ minWidth: "200px" }}
                                    value={r.target}
                                    onChange={formik.handleChange}
                                  />
                                </div>
                              ),
                            },
                            {
                              title: t('mgmt.servapps.newContainer.volumes.subpathTitle'),
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
                                    name={`volumes[${k}].subpath`}
                                    id="Subpath"
                                    disabled={frozenVolumes.includes(r.source)}
                                    placeholder={t('mgmt.servapps.newContainer.volumes.subpathPlaceholder')}
                                    style={{ minWidth: "200px" }}
                                    value={r.subpath || ""}
                                    onChange={formik.handleChange}
                                  />
                                </div>
                              ),
                            },
                            {
                              title: t('mgmt.servapps.newContainer.volumes.readOnlyTitle'),
                              field: (r, k) => (
                                <div
                                  style={{
                                    fontWeight: "bold",
                                    wordSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                  }}
                                >
                                  <Checkbox
                                    className="px-2 my-2"
                                    name={`volumes[${k}].readOnly`}
                                    id="ReadOnly"
                                    disabled={frozenVolumes.includes(r.source)}
                                    checked={!!r.readOnly}
                                    onChange={formik.handleChange}
                                  />
                                </div>
                              ),
                            },
                            {
                              title: t('mgmt.servapps.newContainer.volumes.noCopyTitle'),
                              field: (r, k) => (
                                <div
                                  style={{
                                    fontWeight: "bold",
                                    wordSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                  }}
                                >
                                  <Checkbox
                                    className="px-2 my-2"
                                    name={`volumes[${k}].noCopy`}
                                    id="NoCopy"
                                    disabled={frozenVolumes.includes(r.source) || r.type !== "volume"}
                                    checked={!!r.noCopy}
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
                                    <PermissionGuard permission={PERM_RESOURCES}><Button
                                    variant="outlined"
                                    color="primary"
                                    disabled={frozenVolumes.includes(r.source)}
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
                                  </Button></PermissionGuard>
                                  {!newContainer && containerInfo.Name && (r.target ? <BackupDialog preName={`${containerInfo.Name.replace("/", "").replace("/", "-")}-${r.target.replace("/", "").replaceAll("/", "_")}`} preSource={formatSource(r.source)} refresh={() => setTimeout(refreshAll, 1500)} /> : null)}
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
                          <PermissionGuard permission={PERM_RESOURCES}>
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
                          </PermissionGuard>
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
      {!newContainer && <div
        style={{
          maxWidth: "1000px",
          width: "100%",
          margin: "",
          position: "relative",
      }}>
        {containerInfo && containerInfo.HostConfig && containerInfo.HostConfig.Mounts && <MainCard title={t('mgmt.backup.backups')}>
          <Backups pathFilters={
            containerInfo.HostConfig.Mounts.map((r) => formatSource(r.source || r.Source)).filter(Boolean)
          } />
        </MainCard>}
      </div>}
    </Stack>
  );
};

export default VolumeContainerSetup;
