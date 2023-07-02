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
import IsLoggedIn from '../../../isLoggedIn';
import { useEffect, useState } from 'react';
import { isDomain } from '../../../utils/indexs';

function checkIsOnline() {
    API.isOnline().then((res) => {
        window.location.reload();
    }).catch((err) => {
        setTimeout(() => {
            checkIsOnline();
        }, 1000);
    });
}

const RestartModal = ({openModal, setOpenModal, config}) => {
    const [isRestarting, setIsRestarting] = useState(false);
    const [warn, setWarn] = useState(false);
    const needsRefresh = config && (config.HTTPConfig.HTTPSCertificateMode == "SELFSIGNED" ||
        !isDomain(config.HTTPConfig.Hostname))
    const isNotDomain = config && !isDomain(config.HTTPConfig.Hostname);

    return config ? (needsRefresh && <>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>Refresh Page</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    You need to refresh the page to apply changes. {isNotDomain && 'You are not using a domain names, the server will be offline for a few seconds to remap your docker ports.'}
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => {
                    window.location.reload(true);                
                }}>Refresh</Button>
            </DialogActions>
        </Dialog>
    </>)
    :(<>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>{!isRestarting ? 'Restart Server?' : 'Restarting Server...'}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {warn && <div>
                        <Alert severity="warning" icon={<WarningOutlined />}>
                        The server is taking longer than expected to restart.<br />Consider troubleshouting the logs. If you use a self-signed certificate, you might have to refresh and re-accept it.
                        </Alert>
                    </div>}
                    {isRestarting ? 
                    <div style={{textAlign: 'center', padding: '20px'}}>
                        <CircularProgress />
                    </div>
                    : 'Do you want to restart your server?'}
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
                    }, 20000)
                }}>Restart</Button>
            </DialogActions>}
        </Dialog>
    </>);
};

export default RestartModal;
