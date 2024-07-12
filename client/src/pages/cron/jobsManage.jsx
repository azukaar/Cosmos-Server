import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton, DeleteIconButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DeleteOutlined, DesktopOutlined, EditOutlined, FolderOutlined, LaptopOutlined, LoadingOutlined, MobileOutlined, PlayCircleOutlined, PlusOutlined, ReloadOutlined, SearchOutlined, StopOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, Checkbox, CircularProgress, IconButton, InputLabel, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { crontabToText, isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import ResponsiveButton from "../../components/responseiveButton";
import MenuButton from "../../components/MenuButton";
import JobLogsDialog from "./jobLogs";
import NewJobDialog from "./newJob";
import { useTranslation } from 'react-i18next';

const getStatus = (job) => {
  if (job.Running) return 'running';
  if (job.LastRunSuccess) return 'success';
  if (job.LastRun === "0001-01-01T00:00:00Z") return 'never';
  return 'error';
}

export const CronManager = () => {
  const { t } = useTranslation();
  const [isAdmin, setIsAdmin] = useState(false);
  const [config, setConfig] = useState(null);
  const [cronJobs, setCronJobs] = useState([]);
  const [mergeDialog, setMergeDialog] = useState(null);
  const [loading, setLoading] = useState(false);
  const [jobLogs, setJobLogs] = useState(false);
  const [newJob, setNewJob] = useState(false);

  const deleteCronJob = async (job) => {
    let config = (await API.config.get()).data;
    if(!config.CRON) {
      config.CRON = {};
    }

    delete config.CRON[job.Name];

    return API.config.set(config).then((res) => {
      refresh && refresh();
    }).catch((err) => {
      refresh && refresh();
    });
  };

  const setEnabled = async (name, enabled) => {
    let config = (await API.config.get()).data;
    if(!config.CRON) {
      config.CRON = {};
    }

    config.CRON[name].Enabled = enabled;

    return API.config.set(config).then((res) => {
      refresh && refresh();
    }).catch((err) => {
      refresh && refresh();
    });
  }

  const refresh = async () => {
    setLoading(true);
    let _cronJobs = await API.cron.list();
    let configAsync = await API.config.get();
    setConfig(configAsync.data);
    setIsAdmin(configAsync.isAdmin);
    setCronJobs(_cronJobs.data);
    setLoading(false);
  };

  useEffect(() => {
    refresh();
  }, []);

  return <>
    {(config && cronJobs) ? <>
      {jobLogs && <JobLogsDialog job={jobLogs} OnClose={() => {
        setJobLogs(false);
      }} />}
      {newJob && <NewJobDialog refresh={refresh} job={newJob} OnClose={() => {
        setNewJob(false);
      }} />}
      <Stack spacing={2}>
        <Stack direction="row" spacing={2}>
          <ResponsiveButton variant="contained" startIcon={<PlusOutlined />} onClick={() => {
            setNewJob(true);
          }}>{t('mgmt.cron.newCronTitle')}</ResponsiveButton>
          <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
            refresh();
          }}>{t('global.refresh')}</ResponsiveButton>
        </Stack>
        {Object.keys(cronJobs).map(scheduler => <div>
          <h4>{({
            "Custom": t('mgmt.scheduler.customJobsTitle'),
            "SnapRAID": t('mgmt.scheduler.parityDiskJobsTitle'),
            "__OT__SnapRAID": t('mgmt.scheduler.oneTimeJobsTitle'),
          }[scheduler])}</h4>
          <PrettyTableView 
            data={Object.values(cronJobs[scheduler])}
            getKey={(r) => `${r.Name}`}
            buttons={[
              // <MergerDialog disk={{name: '/dev/sda'}} refresh={refresh}/>,
            ]}
            columns={[
              (scheduler == "Custom" && {
                title: t('global.enabled'),
                clickable:true, 
                field: (r, k) => <Checkbox disabled={loading} size='large' color={!r.Disabled ? 'success' : 'default'}
                  onChange={() => setEnabled(r.Name, r.Disabled)}
                  checked={!r.Disabled}
                />,
              }),
              {
                title: t('global.nameTitle'),
                field: (r) => r.Name,
              },
              {
                title: t('mgmt.scheduler.list.scheduleTitle'),
                field: (r) => crontabToText(r.Crontab, t),
              },
              {
                title: t('global.statusTitle'),
                field: (r) => {
                  return <div style={{maxWidth: '400px'}} >{{
                    'running': <Alert icon={<LoadingOutlined />} severity={'info'} color={'info'}>{t('mgmt.scheduler.list.status.runningSince')}{r.LastStarted}</Alert>,
                    'success': <Alert severity={'success'} color={'success'}>{t('mgmt.scheduler.list.status.lastRunFinishedOn')} {r.LastRun}, {t('mgmt.scheduler.list.status.lastRunFinishedOn.duration')} {(new Date(r.LastRun).getTime() - new Date(r.LastStarted).getTime()) / 1000}s</Alert>,
                    'error': <Alert severity={'error'} color={'error'}>{t('mgmt.scheduler.list.status.lastRunExitedOn')} {r.LastRun}</Alert>,
                    'never': <Alert severity={'info'} color={'info'}>{t('mgmt.scheduler.list.status.neverRan')}</Alert>
                    
                  }[getStatus(r)]}</div>
                }
              },
              {
                title: "",
                field: (r) => <>
                  {r.Running ? <Tooltip title="Stop"><IconButton disabled={!r.Cancellable} onClick={() => {
                    API.cron.stop(scheduler, r.Name).then(() => {
                      refresh();
                    });
                  }}><StopOutlined /></IconButton></Tooltip>
                    : <Tooltip title={t('mgmt.scheduler.list.action.run')}><IconButton onClick={() => {
                      API.cron.run(scheduler, r.Name).then(() => {
                        refresh();
                      });
                    }}><PlayCircleOutlined /></IconButton></Tooltip>}
                  <Tooltip title={t('mgmt.scheduler.list.action.logs')}><IconButton onClick={() => {
                    setJobLogs(r);
                  }}><SearchOutlined /></IconButton></Tooltip>
                  {scheduler == "Custom" && <>
                    <Tooltip title={t('global.edit')}><IconButton onClick={() => {
                      setNewJob(r);
                    }}><EditOutlined /></IconButton></Tooltip>
                    <Tooltip title={t('global.delete')}>
                      <DeleteIconButton onDelete={() => {
                        deleteCronJob(r);
                      }} />
                    </Tooltip>
                  </>}
                </>
              }
            ]}
          />

        </div>)} 
      </Stack>
    </> : <center>
      <CircularProgress color="inherit" size={20} />
    </center>}
  </>
};