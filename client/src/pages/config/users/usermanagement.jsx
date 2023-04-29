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
import IsLoggedIn from '../../../isLoggedIn';
import { useEffect, useState } from 'react';
import PrettyTableView from '../../../components/tableView/prettyTableView';

const UserManagement = () => {
    const [isLoading, setIsLoading] = useState(false);
    const [openCreateForm, setOpenCreateForm] = React.useState(false);
    const [openDeleteForm, setOpenDeleteForm] = React.useState(false);
    const [openInviteForm, setOpenInviteForm] = React.useState(false);
    const [toAction, setToAction] = React.useState(null);

    const roles = ['Guest', 'User', 'Admin']

    const [rows, setRows] = useState(null);

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


        <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
                refresh();
        }}>Refresh</Button>&nbsp;&nbsp;
        <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
            setOpenCreateForm(true)
        }}>Create</Button><br /><br />


        {isLoading && <center><br /><CircularProgress /></center>}

        {!isLoading && rows && (<PrettyTableView 
            data={rows}
            getKey={(r) => r.nickname}
            columns={[
                {
                    title: 'User',
                    // underline: true,
                    field: (r) => <strong>{r.nickname}</strong>,
                },
                {
                    title: 'Status',
                    screenMin: 'sm',
                    field: (r) => {
                        const isRegistered = new Date(r.registeredAt).getTime() > 0;
                        const inviteExpired = new Date(r.registerKeyExp).getTime() < new Date().getTime();

                        return <>{isRegistered ? (r.role > 1 ? <Chip
                                icon={<KeyOutlined />}
                                label="Admin"
                            /> : <Chip
                                icon={<UserOutlined />}
                                label="User"
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
                            )}</>
                    }
                },
                {
                    title: 'Email',
                    screenMin: 'md', 
                    field: (r) => r.email,
                },
                {
                    title: 'Created At',
                    screenMin: 'lg',
                    field: (r) => new Date(r.createdAt).toLocaleString(),
                },
                {
                    title: 'Last Login',
                    screenMin: 'lg', 
                    field: (r) => {
                        const hasLastLogin = new Date(r.lastLogin).getTime() > 0;
                        return <>{hasLastLogin ? new Date(r.lastLogin).toLocaleString() : 'Never'}</>
                    },
                },
                {
                    title: '',
                    clickable: true,
                    field: (r) => {
                        const isRegistered = new Date(r.registeredAt).getTime() > 0;
                        const inviteExpired = new Date(r.registerKeyExp).getTime() < new Date().getTime();

                        return <>{isRegistered ?
                            (<Button variant="contained" color="primary" onClick={
                                () => {
                                    sendlink(r.nickname, 1);
                                }
                            }>Send password reset</Button>) :
                            (<Button variant="contained" className={inviteExpired ? 'shinyButton' : ''} onClick={
                                () => {
                                    sendlink(r.nickname, 2);
                                }
                            } color="primary">Re-Send Invite</Button>)
                        }
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setToAction(r.nickname);
                                setOpenDeleteForm(true);
                            }
                        }>Delete</Button></>
                    }
                },
            ]}
        />)}
    </>;
};

export default UserManagement;
