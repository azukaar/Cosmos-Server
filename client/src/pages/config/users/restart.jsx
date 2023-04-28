// material-ui
import * as React from 'react';
import { Alert, Button, Stack, Typography } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined } from '@ant-design/icons';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Paper from '@mui/material/Paper';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogContentText from '@mui/material/DialogContentText';
import DialogTitle from '@mui/material/DialogTitle';
import TextField from '@mui/material/TextField';
import CircularProgress from '@mui/material/CircularProgress';
import Chip from '@mui/material/Chip';
import IconButton from '@mui/material/IconButton';
import * as API from '../../../api';
import MainCard from '../../../components/MainCard';
import IsLoggedIn from '../../../IsLoggedIn';
import { useEffect, useState } from 'react';

function checkIsOnline() {
    API.isOnline().then((res) => {
        window.location.reload();
    }).catch((err) => {
        setTimeout(() => {
            checkIsOnline();
        }, 1000);
    });
}

const RestartModal = ({openModal, setOpenModal}) => {
    const [isRestarting, setIsRestarting] = useState(false);
    const [warn, setWarn] = useState(false);

    return <>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>{!isRestarting ? 'Restart Server?' : 'Restarting Server...'}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {warn && <div>
                        <Alert severity="warning" icon={<WarningOutlined />}>
                        The server is taking longer than expected to restart.<br />Consider troubleshouting the logs.
                        </Alert>
                    </div>}
                    {isRestarting ? 
                    <div style={{textAlign: 'center', padding: '20px'}}>
                        <CircularProgress />
                    </div>
                    : 'A restart is required to apply changes. Do you want to restart?'}
                </DialogContentText>
            </DialogContent>
            {!isRestarting && <DialogActions>
                <Button onClick={() => setOpenModal(false)}>Later</Button>
                <Button onClick={() => {
                    setIsRestarting(true);
                    API.config.restart()
                    setTimeout(() => {
                        checkIsOnline();
                    }, 1500)
                    setTimeout(() => {
                        setWarn(true);
                    }, 8000)
                }}>Restart</Button>
            </DialogActions>}
        </Dialog>
    </>;
};

export default RestartModal;
