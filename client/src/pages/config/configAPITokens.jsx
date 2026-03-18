import React, { useEffect, useState } from "react";
import * as API from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { PlusOutlined, ReloadOutlined, CopyOutlined, DeleteOutlined, SearchOutlined } from "@ant-design/icons";
import {
  Alert,
  Button,
  Chip,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  ListItemIcon,
  ListItemText,
  MenuItem,
  Stack,
  TextField,
  Tooltip,
} from "@mui/material";
import { CosmosCheckbox, CosmosInputText } from "./users/formShortcuts";
import { FormikProvider, useFormik } from "formik";
import { LoadingButton } from "@mui/lab";
import ResponsiveButton from "../../components/responseiveButton";
import MenuButton from "../../components/MenuButton";
import { useClientInfos } from "../../utils/hooks";
import { PERM_ADMIN } from "../../utils/permissions";
import PermissionGuard from "../../components/permissionGuard";
import dayjs from "dayjs";
import { useTranslation } from "react-i18next";
import EventExplorerStandalone from "../dashboard/eventsExplorerStandalone";
import { ConfirmModalDirect } from "../../components/confirmModal";

const CreateTokenDialog = ({ open, onClose, onCreated, t }) => {
  const formik = useFormik({
    initialValues: {
      name: "",
      description: "",
      readOnly: false,
      ipWhitelist: "",
      restrictToConstellation: false,
      expiryDays: 0,
    },
    validateOnChange: false,
    onSubmit: async (values, { setErrors, setSubmitting }) => {
      setSubmitting(true);
      const request = {
        ...values,
        ipWhitelist: values.ipWhitelist
          ? values.ipWhitelist.split(",").map((s) => s.trim()).filter(Boolean)
          : [],
      };
      try {
        const res = await API.apiTokens.create(request);
        setSubmitting(false);
        onCreated(res.data);
      } catch (err) {
        setErrors({ submit: err.message });
        setSubmitting(false);
      }
    },
  });

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <FormikProvider value={formik}>
        <form onSubmit={formik.handleSubmit}>
          <DialogTitle>{t('mgmt.config.apiTokens.createTitle')}</DialogTitle>
          <DialogContent>
            <Stack spacing={2} style={{ marginTop: "10px" }}>
              <TextField
                fullWidth
                required
                id="name"
                name="name"
                label={t('mgmt.config.apiTokens.tokenName')}
                value={formik.values.name}
                onChange={formik.handleChange}
                error={formik.touched.name && Boolean(formik.errors.name)}
                helperText={formik.touched.name && formik.errors.name}
              />
              <TextField
                fullWidth
                id="description"
                name="description"
                label={t('global.description')}
                value={formik.values.description}
                onChange={formik.handleChange}
              />
              <CosmosCheckbox
                formik={formik}
                name="readOnly"
                label={t('mgmt.config.apiTokens.readOnly')}
              />
              <TextField
                fullWidth
                id="ipWhitelist"
                name="ipWhitelist"
                label={t('mgmt.config.apiTokens.ipWhitelist')}
                value={formik.values.ipWhitelist}
                onChange={formik.handleChange}
                helperText={t('mgmt.config.apiTokens.ipWhitelistHelper')}
              />
              <CosmosCheckbox
                formik={formik}
                name="restrictToConstellation"
                label={t('mgmt.config.apiTokens.restrictToConstellation')}
              />
              <TextField
                fullWidth
                select
                id="expiryDays"
                name="expiryDays"
                label={t('mgmt.config.apiTokens.expiry')}
                value={formik.values.expiryDays}
                onChange={formik.handleChange}
              >
                <MenuItem value={0}>{t('mgmt.config.apiTokens.expiryNever')}</MenuItem>
                <MenuItem value={7}>{t('mgmt.config.apiTokens.expiry7Days')}</MenuItem>
                <MenuItem value={30}>{t('mgmt.config.apiTokens.expiry30Days')}</MenuItem>
                <MenuItem value={90}>{t('mgmt.config.apiTokens.expiry90Days')}</MenuItem>
              </TextField>
              {formik.errors.submit && (
                <Alert severity="error">{formik.errors.submit}</Alert>
              )}
            </Stack>
          </DialogContent>
          <DialogActions>
            <Button onClick={onClose}>{t('global.cancelAction')}</Button>
            <LoadingButton
              color="primary"
              variant="contained"
              type="submit"
              loading={formik.isSubmitting}
            >
              {t('global.createAction')}
            </LoadingButton>
          </DialogActions>
        </form>
      </FormikProvider>
    </Dialog>
  );
};

const TokenRevealDialog = ({ tokenData, onClose, t }) => {
  const [copied, setCopied] = useState(false);

  if (!tokenData) return null;

  const handleCopy = () => {
    navigator.clipboard.writeText(tokenData.token);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <Dialog open={!!tokenData} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>{t('mgmt.config.apiTokens.tokenCreated')}</DialogTitle>
      <DialogContent>
        <Stack spacing={2} style={{ marginTop: "10px" }}>
          <Alert severity="warning">
            {t('mgmt.config.apiTokens.copyWarning')}
          </Alert>
          <Stack direction="row" spacing={1} alignItems="center">
            <TextField
              fullWidth
              value={tokenData.token}
              InputProps={{ readOnly: true }}
              size="small"
              sx={{ fontFamily: "monospace" }}
            />
            <Tooltip title={copied ? t('global.copied') : t('global.copyFilenameSuffix')}>
              <IconButton onClick={handleCopy}>
                <CopyOutlined />
              </IconButton>
            </Tooltip>
          </Stack>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button variant="contained" onClick={onClose}>
          {t('global.done')}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

const TokenEventsDialog = ({ tokenName, onClose, t }) => {
  if (!tokenName) return null;

  return (
    <Dialog open={!!tokenName} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>{t('mgmt.config.apiTokens.events', { name: tokenName })}</DialogTitle>
      <DialogContent>
        <EventExplorerStandalone
          initLevel="info"
          initSearch={`{"object":"token@${tokenName}"}`}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{t('global.close')}</Button>
      </DialogActions>
    </Dialog>
  );
};

const permissionLabelKeys = {
  1: "mgmt.config.apiTokens.permAdminRead",
  2: "mgmt.config.apiTokens.permAdmin",
  10: "mgmt.config.apiTokens.permUsersRead",
  11: "mgmt.config.apiTokens.permUsers",
  20: "mgmt.config.apiTokens.permResourcesRead",
  21: "mgmt.config.apiTokens.permResources",
  30: "mgmt.config.apiTokens.permConfigRead",
  31: "mgmt.config.apiTokens.permConfig",
};

const formatPermissions = (permissions, t) => {
  if (!permissions || permissions.length === 0) return [];
  return permissions
    .filter((p) => p < 100)
    .map((p) => permissionLabelKeys[p] ? t(permissionLabelKeys[p]) : t('mgmt.config.apiTokens.permUnknown', { id: p }));
};

const ConfigAPITokens = () => {
  const { t } = useTranslation();
  const [tokens, setTokens] = useState(null);
  const [loading, setLoading] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [revealedToken, setRevealedToken] = useState(null);
  const [monitorToken, setMonitorToken] = useState(null);
  const [confirmDelete, setConfirmDelete] = useState(null);
  const { hasPermission } = useClientInfos();

  const refresh = async () => {
    setLoading(true);
    try {
      const res = await API.apiTokens.list();
      setTokens(res.data || {});
    } catch (err) {
      setTokens({});
    }
    setLoading(false);
  };

  useEffect(() => {
    refresh();
  }, []);

  return (
    <Stack spacing={2}>
      <CreateTokenDialog
        t={t}
        open={createOpen}
        onClose={() => setCreateOpen(false)}
        onCreated={(data) => {
          setCreateOpen(false);
          setRevealedToken(data);
          refresh();
        }}
      />
      <TokenRevealDialog
        t={t}
        tokenData={revealedToken}
        onClose={() => setRevealedToken(null)}
      />
      <TokenEventsDialog
        t={t}
        tokenName={monitorToken}
        onClose={() => setMonitorToken(null)}
      />
      {confirmDelete && (
        <ConfirmModalDirect
          callback={() => {
            API.apiTokens.remove(confirmDelete).then(() => refresh());
          }}
          content={t('mgmt.config.apiTokens.deleteConfirm', { name: confirmDelete })}
          onClose={() => setConfirmDelete(null)}
        />
      )}

      <Stack direction="row" spacing={2}>
        <PermissionGuard permission={PERM_ADMIN}>
          <ResponsiveButton
            variant="contained"
            startIcon={<PlusOutlined />}
            onClick={() => setCreateOpen(true)}
          >
            {t('mgmt.config.apiTokens.newToken')}
          </ResponsiveButton>
        </PermissionGuard>
        <ResponsiveButton
          variant="outlined"
          startIcon={<ReloadOutlined />}
          onClick={refresh}
        >
          {t('global.refresh')}
        </ResponsiveButton>
      </Stack>

      {tokens === null ? (
        <center>
          <CircularProgress color="inherit" size={20} />
        </center>
      ) : Object.keys(tokens).length === 0 ? (
        <Alert severity="info">{t('mgmt.config.apiTokens.noTokens')}</Alert>
      ) : (
        <PrettyTableView
          data={Object.values(tokens)}
          getKey={(r) => r.name}
          columns={[
            {
              title: t('global.name'),
              field: (r) => (
                <Stack direction="row" alignItems="center" spacing={1}>
                  <span>{r.name}</span>
                  {r.tokenSuffix && (
                    <Chip
                      label={`***${r.tokenSuffix}`}
                      size="small"
                      sx={{ fontFamily: "monospace", fontSize: "0.75rem" }}
                    />
                  )}
                </Stack>
              ),
            },
            {
              title: t('global.owner'),
              field: (r) => r.owner || "-",
            },
            {
              title: t('global.permissions'),
              field: (r) => {
                const labels = formatPermissions(r.permissions, t);
                // For now, simplify to Full Access / Read Only based on whether PERM_ADMIN (1) is present
                const hasAdmin = r.permissions && r.permissions.includes(PERM_ADMIN);
                const displayLabel = hasAdmin ? t('mgmt.config.apiTokens.fullAccess') : t('mgmt.config.apiTokens.readOnly');
                const displayColor = hasAdmin ? "warning" : "info";
                return (
                  <Tooltip title={labels.join(", ") || t('global.none')}>
                    <Chip
                      label={displayLabel}
                      size="small"
                      color={displayColor}
                      sx={{ cursor: "pointer" }}
                    />
                  </Tooltip>
                );
              },
            },
            {
              title: t('global.createdAt'),
              field: (r) =>
                r.createdAt
                  ? dayjs(r.createdAt).format("L, LT")
                  : "-",
            },
            {
              title: t('mgmt.config.apiTokens.expiryColumn'),
              field: (r) => {
                if (!r.expiresAt) return t('mgmt.config.apiTokens.expiryNever');
                const exp = dayjs(r.expiresAt);
                const isExpired = exp.isBefore(dayjs());
                return (
                  <Chip
                    label={isExpired
                      ? t('mgmt.config.apiTokens.expired')
                      : exp.format("L, LT")}
                    size="small"
                    color={isExpired ? "error" : "default"}
                  />
                );
              },
            },
            {
              title: "",
              field: (r) => (
                <MenuButton>
                  <MenuItem onClick={() => setMonitorToken(r.name)}>
                    <ListItemIcon><SearchOutlined /></ListItemIcon>
                    <ListItemText>{t('global.monitor')}</ListItemText>
                  </MenuItem>
                  <PermissionGuard permission={PERM_ADMIN}>
                    <MenuItem onClick={() => setConfirmDelete(r.name)} sx={{ color: "error.main" }}>
                      <ListItemIcon><DeleteOutlined style={{ color: "inherit" }} /></ListItemIcon>
                      <ListItemText>{t('global.delete')}</ListItemText>
                    </MenuItem>
                  </PermissionGuard>
                </MenuButton>
              ),
            },
          ]}
        />
      )}
    </Stack>
  );
};

export default ConfigAPITokens;
