// material-ui
import * as React from 'react';
import {
    Button,
    Typography,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableHead,
    TableRow,
    Paper,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    TextField,
    CircularProgress,
    Chip,
    IconButton,
} from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined } from '@ant-design/icons';
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
    const [openEditEmail, setOpenEditEmail] = React.useState(false);
    const [toAction, setToAction] = React.useState(null);
    const [loadingRow, setLoadingRow] = React.useState(null);

    const roles = ['Guest', 'User', 'Admin']

    const [rows, setRows] = useState(null);

    function refresh() {
        setIsLoading(true);
        API.users.list()
        .then(data => {
            setLoadingRow(null);
            setRows(data.data);
            setIsLoading(false);
        })
    }

    useEffect(() => {
        refresh();
    }, [])

    function sendlink(nickname, formType) {
        API.users.invite({  
            nickname,
            formType: ""+formType,
        })
        .then((values) => {
            let sendLink = window.location.origin + '/cosmos-ui/register?t='+formType+'&nickname='+nickname+'&key=' + values.data.registerKey;
            setToAction({...values.data, nickname, sendLink, formType, formAction: formType === 2 ? 'invite them to the server' : 'let them reset their password'});
            setOpenInviteForm(true);
        });
    }

    return <>
        {openInviteForm ? <Dialog open={openInviteForm} onClose={() => setOpenInviteForm(false)}>
            <DialogTitle>Invite User</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <div style={{
                        paddingBottom: '15px',
                        maxWidth: '350px',
                    }}>
                        {toAction.emailWasSent ? 
                            <div>
                                <strong>An email has been sent</strong> with a link to {toAction.formAction}. Alternatively you can also share the link below:
                            </div> :
                            <div>
                                Send this link to {toAction.nickname} to {toAction.formAction}:
                            </div>
                        }
                    </div>
                    
                    <div>
                    <div style={{float: 'left', width: '300px', padding: '5px', background:'rgba(0,0,0,0.15)', whiteSpace: 'nowrap', wordBreak: 'keep-all', overflow: 'auto', fontStyle: 'italic'}}>{toAction.sendLink}</div>
                        <IconButton size="large" style={{float: 'left'}} aria-label="copy" onClick={
                            () => {
                                navigator.clipboard.writeText(toAction.sendLink);
                            }
                        }>
                            <CopyOutlined  />
                        </IconButton>
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
        
        <Dialog open={openEditEmail} onClose={() => setOpenEditEmail(false)}>
            <DialogTitle>Edit Email</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Use this form to invite edit {openEditEmail}'s Email.
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email-edit"
                    label="Email Address"
                    type="email"
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenEditEmail(false)}>Cancel</Button>
                <Button onClick={() => {
                    API.users.edit(openEditEmail, {
                        email: document.getElementById('c-email-edit').value,
                    }).then(() => {
                        setOpenEditEmail(false);
                        refresh();
                    });
                }}>Edit</Button>
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
                        sendlink(document.getElementById('c-nickname').value, 2);
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
            onRowClick = {(r) => {
                setOpenEditEmail(r.nickname);
            }}
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

                        if (loadingRow === r.nickname) {
                            return <div style={{textAlign: 'center'}}><CircularProgress /></div>
                        }

                        return <>{isRegistered ?
                            (<Button variant="contained" color="primary" onClick={
                                () => {
                                    setLoadingRow(r.nickname);
                                    sendlink(r.nickname, 1);
                                }
                            }>Send password reset</Button>) :
                            (<Button variant="contained" className={inviteExpired ? 'shinyButton' : ''} onClick={
                                () => {
                                    setLoadingRow(r.nickname);
                                    sendlink(r.nickname, 2);
                                }
                            } color="primary">Re-Send Invite</Button>)
                        }
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                setToAction(r.nickname);
                                setOpenDeleteForm(true);
                            }
                        }>Delete</Button>
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                API.users.reset2FA(r.nickname).then(() => {
                                    refresh();
                                });
                            }
                        }>Reset 2FA</Button></>
                    }
                },
            ]}
        />)}
    </>;
};

export default UserManagement;
