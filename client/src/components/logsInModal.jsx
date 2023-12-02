// material-ui
import * as React from 'react';
import {
  Alert,
  Button,
  Stack,
  Typography,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
} from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined, ArrowUpOutlined, FileZipOutlined } from '@ant-design/icons';
import { useEffect, useState } from 'react';
import { smartDockerLogConcat, tryParseProgressLog } from '../utils/docker';

const preStyle = {
  backgroundColor: '#000',
  color: '#fff',
  padding: '10px',
  borderRadius: '5px',
  overflow: 'auto',
  maxHeight: '500px',
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
}

const LogsInModal = ({title, request, OnSuccess, OnError, closeAnytime, initialLogs = [], }) => {
    const [openModal, setOpenModal] = useState(false);
    const [logs, setLogs] = useState(initialLogs);
    const [done, setDone] = useState(closeAnytime);
    const preRef = React.useRef(null);

    useEffect(() => {
      setLogs(initialLogs);
      setDone(closeAnytime);

      if(request === null) return;

      request((newlog) => {
        setLogs((old) => smartDockerLogConcat(old, newlog));

        if(preRef.current)
          preRef.current.scrollTop = preRef.current.scrollHeight;
          
        if (newlog.includes('[OPERATION SUCCEEDED]')) {
          setDone(true);
          setOpenModal(false);
          OnSuccess && OnSuccess();
        } else if (newlog.includes('[OPERATION FAILED]')) {
          setDone(true);
          OnError && OnError(newlog);
        } else {
          setOpenModal(true);
        }
      });
    }, [request]);

    return <>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>{title}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                  <pre style={preStyle} ref={preRef}>
                    {logs.map((l) => {
                      return <div>{tryParseProgressLog(l)}</div>
                    })}
                  </pre>
                </DialogContentText>
            </DialogContent>
            <DialogActions>
            </DialogActions>
        </Dialog>
    </>;
};

export default LogsInModal;
