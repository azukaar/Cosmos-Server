// material-ui
import * as React from 'react';
import { Button, Typography } from '@mui/material';
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

const RestartModal = ({openModal, setOpenModal}) => {
    return <>
        <Dialog open={openModal} onClose={() => setOpenModal(false)}>
            <DialogTitle>Restart Server</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    A restart is required to apply changes. Do you want to restart?
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenModal(false)}>Later</Button>
                <Button onClick={() => {
                    API.config.restart()
                    .then(() => {
                        refresh();
                        setOpenModal(false);
                        setTimeout(() => {
                          window.location.reload();
                        }, 2000)
                    })
                }}>Restart</Button>
            </DialogActions>
        </Dialog>
    </>;
};

export default RestartModal;
