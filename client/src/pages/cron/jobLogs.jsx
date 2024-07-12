import React from "react";
import { useEffect, useState } from "react";
import * as API  from "../../api";
import PrettyTableView from "../../components/tableView/prettyTableView";
import { DeleteButton } from "../../components/delete";
import { CheckCircleOutlined, CloudOutlined, CompassOutlined, DesktopOutlined, ExpandOutlined, InfoCircleOutlined, LaptopOutlined, MenuFoldOutlined, MenuOutlined, MinusCircleFilled, MobileOutlined, NodeCollapseOutlined, PlusCircleFilled, PlusCircleOutlined, ReloadOutlined, SettingFilled, TabletOutlined, WarningFilled, WarningOutlined } from "@ant-design/icons";
import { Alert, Button, CircularProgress, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle, InputLabel, LinearProgress, ListItemIcon, ListItemText, MenuItem, Stack, Tooltip } from "@mui/material";
import { CosmosCheckbox, CosmosFormDivider, CosmosInputText } from "../config/users/formShortcuts";
import MainCard from "../../components/MainCard";
import { Formik } from "formik";
import { LoadingButton, TreeItem, TreeView } from "@mui/lab";
import ApiModal from "../../components/apiModal";
import ConfirmModal from "../../components/confirmModal";
import { PascalToSnake, isDomain } from "../../utils/indexs";
import UploadButtons from "../../components/fileUpload";
import { useTheme } from '@mui/material/styles';
import MiniPlotComponent from '../dashboard/components/mini-plot';
import LogLine from "../../components/logLine";
import { useTranslation } from 'react-i18next';

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '520px',
  maxWidth: '100%',
  width: '100%',
  margin: '0',
  position: 'relative',
  fontSize: '12px',
  fontFamily: 'monospace',
  whiteSpace: 'pre-wrap',
  wordWrap: 'break-word',
  wordBreak: 'break-all',
  lineHeight: '1.5',
  boxShadow: '0 0 10px rgba(0,0,0,0.5)',
  border: '1px solid rgba(255,255,255,0.1)',
  boxSizing: 'border-box',
  marginBottom: '10px',
  marginTop: '10px',
  marginLeft: '0',
  marginRight: '0',
  display: 'block',
  textAlign: 'left',
  verticalAlign: 'baseline',
  opacity: '1',
}

const JobLogsDialog = ({job, OnClose}) => {
  const { t } = useTranslation();
  const [jobFull, setJobFull] = useState(null);
  
  useEffect(() => {
    if (job) {
      API.cron.get(job.Scheduler, job.Name).then((res) => {
        setJobFull(res.data);
      });
    }
  }, [job]);

  return <Dialog open={job} onClose={() => {
      OnClose && OnClose();
    }}>
      <DialogTitle>{t('mgmt.scheduler.lastLogs')} {job.Name}</DialogTitle>
      <DialogContent>
          <DialogContentText>
            <pre style={preStyle}>
              {jobFull && jobFull.Logs && jobFull.Logs.map((log, i) => {
                return <LogLine message={log} />;
              })}
            </pre>
          </DialogContentText>
      </DialogContent>
      <DialogActions>
          <Button onClick={() => {
              OnClose && OnClose();
          }}>{t('global.close')}</Button>
      </DialogActions>
  </Dialog>
};

export default JobLogsDialog;