import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton, DeleteIconButton } from "../../components/delete";
import { CloudOutlined, CloudServerOutlined, CompassOutlined, DeleteOutlined, DesktopOutlined, EditOutlined, FolderOutlined, LaptopOutlined, LoadingOutlined, MobileOutlined, PlayCircleOutlined, PlusOutlined, ReloadOutlined, SearchOutlined, StopOutlined, TabletOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, IconButton, InputLabel, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
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

const getStatus = (job) => {
  if (job.Running) return 'running';
  if (job.LastRunSuccess) return 'success';
  if (job.LastRun === "0001-01-01T00:00:00Z") return 'never';
  return 'error';
}

export const CronManager = () => {
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
          }}>New Job</ResponsiveButton>
          <ResponsiveButton variant="outlined" startIcon={<ReloadOutlined />} onClick={() => {
            refresh();
          }}>Refresh</ResponsiveButton>
        </Stack>
        {Object.keys(cronJobs).map(scheduler => <div>
          <h4>{({
            "Custom": "Custom Jobs",
            "SnapRAID": "Parity Disks Jobs",
          }[scheduler])}</h4>
          <PrettyTableView 
            data={Object.values(cronJobs[scheduler])}
            getKey={(r) => `${r.Name}`}
            buttons={[
              // <MergerDialog disk={{name: '/dev/sda'}} refresh={refresh}/>,
            ]}
            columns={[
              {
                title: 'Name',
                field: (r) => r.Name,
              },
              {
                title: 'Schedule',
                field: (r) => crontabToText(r.Crontab),
              },
              {
                title: "Status",
                field: (r) => {
                  return <div style={{maxWidth: '400px'}} >{{
                    'running': <Alert icon={<LoadingOutlined />} severity={'info'} color={'info'}>Running since {r.LastStarted}</Alert>,
                    'success': <Alert severity={'success'} color={'success'}>Last run finished on {r.LastRun}</Alert>,
                    'error': <Alert severity={'error'} color={'error'}>Last run exited with an error on {r.LastRun}</Alert>,
                    'never': <Alert severity={'info'} color={'info'}>Never ran</Alert>
                    
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
                    : <Tooltip title="Run"><IconButton onClick={() => {
                      API.cron.run(scheduler, r.Name).then(() => {
                        refresh();
                      });
                    }}><PlayCircleOutlined /></IconButton></Tooltip>}
                  <Tooltip title="Logs"><IconButton onClick={() => {
                    setJobLogs(r);
                  }}><SearchOutlined /></IconButton></Tooltip>
                  {scheduler == "Custom" && <>
                    <Tooltip title="Edit"><IconButton onClick={() => {
                      setNewJob(r);
                    }}><EditOutlined /></IconButton></Tooltip>
                    <Tooltip title="Delete">
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