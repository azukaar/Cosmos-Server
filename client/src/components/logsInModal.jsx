// material-ui
import * as React from 'react';
import { Alert, Button, Stack, Typography } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined, ArrowUpOutlined, FileZipOutlined } from '@ant-design/icons';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import { useEffect, useState } from 'react';
import { smartDockerLogConcat, tryParseProgressLog } from '../utils/docker';

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
}

const LogsInModal = ({title, request, OnSuccess, OnError, closeAnytime, OnClose, initialLogs = [], alwaysShow = false }) => {
    const [openModal, setOpenModal] = useState(alwaysShow);
    const [logs, setLogs] = useState(initialLogs);
    const [done, setDone] = useState(closeAnytime);
    const preRef = React.useRef(null);

    const close = () => {
      OnClose && OnClose();
      setOpenModal(false);
    }

    useEffect(() => {
      setLogs(initialLogs);
      setDone(closeAnytime);

      if(request === null) return;

      try {
        request((newlog) => {
          setLogs((old) => smartDockerLogConcat(old, newlog));

          if(preRef.current)
            preRef.current.scrollTop = preRef.current.scrollHeight;
            
          if (newlog.includes('[OPERATION SUCCEEDED]')) {
            setDone(true);
            if(!alwaysShow)
              close();
            OnSuccess && OnSuccess();
          } else if (newlog.includes('[OPERATION FAILED]')) {
            setDone(true);
            OnError && OnError(newlog);
          } else {
            setOpenModal(true);
          }
        }).catch((err) => {
          setLogs((old) => smartDockerLogConcat(old, err.message));
          OnError && OnError(err);
        });
      } catch(err) {
        setLogs((old) => smartDockerLogConcat(old, err.message));
        OnError && OnError(err);
      };
    }, []);

    return <>
        <Dialog open={openModal} onClose={() => close()}>
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
                {alwaysShow && <Button onClick={() => close()} disabled={!done} color="primary">
                    Close
                </Button>}
            </DialogActions>
        </Dialog>
    </>;
};

export default LogsInModal;
