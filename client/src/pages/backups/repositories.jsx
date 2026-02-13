import React from "react";
import { useEffect, useState } from "react";
import * as API from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { CloudOutlined, CloudServerOutlined, DeleteOutlined, EditOutlined, FolderOutlined, LockOutlined, ReloadOutlined, WarningOutlined } from "@ant-design/icons";
import { Alert, Checkbox, Chip, CircularProgress, ListItemIcon, ListItemText, MenuItem, Stack } from "@mui/material";
import { crontabToText } from "../../utils/indexs";
import MenuButton from "../../components/MenuButton";
import ResponsiveButton from "../../components/responseiveButton";
import { useTranslation } from "react-i18next";
import { ConfirmModalDirect } from "../../components/confirmModal";
import BackupDialog, { BackupDialogInternal } from "./backupDialog";
import {simplifyNumber} from '../dashboard/components/utils';

export const Repositories = () => {
  const { t } = useTranslation();
  // const [isAdmin, setIsAdmin] = useState(false);
  let isAdmin = true;
  const [repositories, setRepositories] = useState(null);
  const [repoStats, setRepoStats] = useState({});
  const [loading, setLoading] = useState(false);
  const [deleteBackup, setDeleteBackup] = useState(null);
  const [editOpened, setEditOpened] = useState(null);
  const [unlocking, setUnlocking] = useState(null);

  const refresh = async () => {
    setLoading(true);
    setRepoStats({});
    let configAsync = await API.backups.listRepo();
    const repos = configAsync.data || {};
    setRepositories(repos);
    setLoading(false);

    // Fetch stats for each repo in parallel
    Object.values(repos).forEach((repo) => {
      fetchRepoStats(repo.id);
    });
  };

  const handleUnlock = async (repoId) => {
    setUnlocking(repoId);
    try {
      await API.backups.unlockRepository(repoId);
      // Re-fetch stats for this repo after unlock
      fetchRepoStats(repoId);
    } catch (error) {
      console.error('Failed to unlock repository:', error);
    } finally {
      setUnlocking(null);
      // refresh
      setTimeout(() => {
        refresh();
      }, 1000);
    }
  };

  const fetchRepoStats = async (repoId) => {
    try {
      const res = await API.backups.repoStats(repoId);
      setRepoStats(prev => ({ ...prev, [repoId]: { status: 'ok', data: res.data } }));
    } catch (error) {
      setRepoStats(prev => ({ ...prev, [repoId]: { status: 'error', error: error.message || String(error) } }));
    }
  };

  useEffect(() => {
    refresh();
  }, []);

  if (loading) {
    return (
      <Stack direction="row" spacing={2} justifyContent="center" alignContent={"center"} alignItems={"center"}>
        <CircularProgress />
      </Stack>
    );
  }

  const StatsCell = ({ repoId, render }) => {
    const stats = repoStats[repoId];
    if (!stats) return <CircularProgress size={16} />;
    if (stats.status === 'error') return <Chip label="Error" color="error" size="small" />;
    return render(stats.data);
  };

  return  <div style={{ maxWidth: "1200px", margin: "auto" }}>
    {(repositories) ? <>
      <Stack direction="row" spacing={2} justifyContent="flex-start">
        <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={refresh}>
          {t('global.refresh')}
        </ResponsiveButton>
      </Stack>
      <PrettyTableView
        linkTo={r => '/cosmos-ui/backups/repo/' + r.id}
        data={Object.values(repositories)}
        getKey={(r) => r.Name}
        columns={[
          {
            title: '',
            field: (r) => {
              const stats = repoStats[r.id];
              if (stats && stats.status === 'error') return <WarningOutlined style={{color: '#ff4d4f'}} />;
              return <FolderOutlined />;
            },
            style: {
              textAlign: 'right',
              width: '64px',
            },
          },
          {
            title: t('mgmt.backup.repository'),
            field: (r) => r.path,
            underline: true,
          },
          {
            title: t('mgmt.backup.repository.size'),
            field: (r) => <StatsCell repoId={r.id} render={(data) =>
              data && (simplifyNumber(data.total_uncompressed_size, "B") + " (" + simplifyNumber(data.total_size, "B") + " on disk)")
            } />,
            underline: true,
          },
          {
            title: t('mgmt.backup.repository.total_file_count'),
            field: (r) => <StatsCell repoId={r.id} render={(data) =>
              data && (simplifyNumber(data.total_file_count, ""))
            } />,
            underline: true,
          },
          {
            title: t('mgmt.backup.repository.snapshots_count'),
            field: (r) => <StatsCell repoId={r.id} render={(data) =>
              data && (data.snapshots_count)
            } />,
            underline: true,
            },
          {
            title: '',
            field: (r) => {
              if (!r.locks || r.locks.length === 0) return null;
              const lock = r.locks[0];
              const lockDate = new Date(lock.time);
              const ago = Math.round((Date.now() - lockDate.getTime()) / 1000 / 60);
              const agoText = ago < 60 ? `${ago}m ago` : ago < 1440 ? `${Math.round(ago / 60)}h ago` : `${Math.round(ago / 1440)}d ago`;
              return <Stack direction="row" spacing={1} alignItems="center">
                <Chip
                  icon={<LockOutlined />}
                  label={`Locked (${agoText})`}
                  color="warning"
                  size="small"
                  variant="outlined"
                />
                <ResponsiveButton
                  variant="outlined"
                  color="warning"
                  size="small"
                  startIcon={unlocking === r.id ? <CircularProgress size={16} /> : null}
                  disabled={unlocking === r.id}
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    handleUnlock(r.id);
                  }}
                >
                  Unlock
                </ResponsiveButton>
              </Stack>;
            },
          },
        ]}
      />
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </div>;
};
