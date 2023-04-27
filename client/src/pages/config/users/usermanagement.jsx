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

const UserManagement = () => {
    const [isLoading, setIsLoading] = useState(false);
    const [openCreateForm, setOpenCreateForm] = React.useState(false);
    const [openDeleteForm, setOpenDeleteForm] = React.useState(false);
    const [openInviteForm, setOpenInviteForm] = React.useState(false);
    const [toAction, setToAction] = React.useState(null);

    const roles = ['Guest', 'User', 'Admin']

    const [rows, setRows] = useState([]);

    function refresh() {
        setIsLoading(true);
        API.users.list()
        .then(data => {
            setRows(data.data);
            setIsLoading(false);
        })
    }

    useEffect(() => {
        refresh();
    }, [])

    function sendlink(nickname, formType) {
        API.users.invite({  
            nickname
        })
        .then((values) => {
            let sendLink = window.location.origin + '/ui/register?t='+formType+'&nickname='+nickname+'&key=' + values.data.registerKey;
            setToAction({...values.data, nickname, sendLink, formType});
            setOpenInviteForm(true);
        });
    }

    return <>
        {openInviteForm ? <Dialog open={openInviteForm} onClose={() => setOpenInviteForm(false)}>
            <IsLoggedIn />
            <DialogTitle>Invite User</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Send this link to {toAction.nickname} to invite them to the system:
                    <div>
                        <IconButton size="large" style={{float: 'left'}} aria-label="copy" onClick={
                            () => {
                                navigator.clipboard.writeText(toAction.sendLink);
                            }
                        }>
                            <CopyOutlined  />
                        </IconButton><div style={{float: 'left', width: '300px', padding: '5px', background:'rgba(0,0,0,0.15)', whiteSpace: 'nowrap', wordBreak: 'keep-all', overflow: 'auto', fontStyle: 'italic'}}>{toAction.sendLink}</div>
                    </div>
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => {
                    setOpenInviteForm(false);
                    refresh();
                }}>Close</Button>
            </DialogActions>
        </Dialog>: ''}

        <Dialog open={openDeleteForm} onClose={() => setOpenDeleteForm(false)}>
            <DialogTitle>Delete User</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Are you sure you want to delete user {toAction} ?
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenDeleteForm(false)}>Cancel</Button>
                <Button onClick={() => {
                    API.users.deleteUser(toAction)
                    .then(() => {
                        refresh();
                        setOpenDeleteForm(false);
                    })
                }}>Delete</Button>
            </DialogActions>
        </Dialog>

        <Dialog open={openCreateForm} onClose={() => setOpenCreateForm(false)}>
            <DialogTitle>Create User</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Use this form to invite a new user to the system.
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-nickname"
                    label="Nickname"
                    type="text"
                    fullWidth
                    variant="standard"
                />
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email"
                    label="Email Address (Optional)"
                    type="email"
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenCreateForm(false)}>Cancel</Button>
                <Button onClick={() => {
                    API.users.create({
                        nickname: document.getElementById('c-nickname').value,
                        email: document.getElementById('c-email').value,
                    }).then(() => {
                        setOpenCreateForm(false);
                        refresh();
                        sendlink(document.getElementById('c-nickname').value, 'create');
                    });
                }}>Create</Button>
            </DialogActions>
        </Dialog>

        <MainCard title="Users">
            <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
                refresh();
            }}>Refresh</Button>&nbsp;&nbsp;
            <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
                setOpenCreateForm(true)
            }}>Create</Button><br /><br />
            {isLoading ? <center><br /><CircularProgress color="inherit" /></center>
            : <TableContainer component={Paper}>
                <Table aria-label="simple table">
                    <TableHead>
                        <TableRow>
                            <TableCell>Nickname</TableCell>
                            <TableCell>Status</TableCell>
                            <TableCell>Created At</TableCell>
                            <TableCell>Last Login</TableCell>
                            <TableCell>Actions</TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {rows.map((row) => {
                            const isRegistered = new Date(row.registeredAt).getTime() > 0;
                            const inviteExpired = new Date(row.registerKeyExp).getTime() < new Date().getTime();

                            const hasLastLogin = new Date(row.lastLogin).getTime() > 0;

                            return (
                                <TableRow
                                    key={row.nickname}
                                    sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                                >
                                    <TableCell component="th" scope="row">
                                        &nbsp;&nbsp;<strong>{row.nickname}</strong>
                                    </TableCell>
                                    <TableCell>
                                        {isRegistered ? (row.role > 1 ? <Chip
                                            icon={<KeyOutlined />}
                                            label="Admin"
                                            variant="outlined"
                                        /> : <Chip
                                            icon={<UserOutlined />}
                                            label="User"
                                            variant="outlined"
                                        />) : (
                                            inviteExpired ? <Chip
                                                icon={<ExclamationCircleOutlined  />}
                                                label="Invite Expired"
                                                color="error"
                                            /> : <Chip
                                                icon={<WarningOutlined />}
                                                label="Invite Pending"
                                                color="warning"
                                            />
                                        )}
                                    </TableCell>
                                    <TableCell>
                                        {new Date(row.createdAt).toLocaleDateString()} -&nbsp;
                                        {new Date(row.createdAt).toLocaleTimeString()}
                                    </TableCell>
                                    <TableCell>
                                        {hasLastLogin ? <span>
                                            {new Date(row.lastLogin).toLocaleDateString()} -&nbsp;
                                            {new Date(row.lastLogin).toLocaleTimeString()}
                                        </span> : '-'}
                                    </TableCell>
                                    <TableCell>
                                        {isRegistered ?
                                            (<Button variant="contained" color="primary" onClick={
                                                () => {
                                                    sendlink(row.nickname, 1);
                                                }
                                            }>Send password reset</Button>) :
                                            (<Button variant="contained" className={inviteExpired ? 'shinyButton' : ''} onClick={
                                                () => {
                                                    sendlink(row.nickname, 2);
                                                }
                                            } color="primary">Re-Send Invite</Button>)
                                        }
                                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                                            () => {
                                                setToAction(row.nickname);
                                                setOpenDeleteForm(true);
                                            }
                                        }>Delete</Button></TableCell>
                                </TableRow>
                            )
                        })}
                    </TableBody>
                </Table>
            </TableContainer>}
        </MainCard>
    </>;
};

export default UserManagement;
