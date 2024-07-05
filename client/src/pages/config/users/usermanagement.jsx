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
import { useEffect, useState } from 'react';
import PrettyTableView from '../../../components/tableView/prettyTableView';
import { Trans, useTranslation } from 'react-i18next';

const UserManagement = () => {
    const { t } = useTranslation();
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
            <DialogTitle>{t('InviteUser')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <div style={{
                        paddingBottom: '15px',
                        maxWidth: '350px',
                    }}>
                        {toAction.emailWasSent ? 
                            <div>
                                <strong>{t('EmailSent')}</strong> {t('WithALink')} {toAction.formAction}. {t('AlternativeShareThisLink')}
                            </div> :
                            <div>
                                {t('SendThisTo')} {toAction.nickname} {t('To')} {toAction.formAction}:
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
                }}>{t('Close')}</Button>
            </DialogActions>
        </Dialog>: ''}

        <Dialog open={openDeleteForm} onClose={() => setOpenDeleteForm(false)}>
            <DialogTitle>Delete User</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('ConfirmDeleteUser')} {toAction} ?
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenDeleteForm(false)}>{t('Cancel')}</Button>
                <Button onClick={() => {
                    API.users.deleteUser(toAction)
                    .then(() => {
                        refresh();
                        setOpenDeleteForm(false);
                    })
                }}>{t('Delete')}</Button>
            </DialogActions>
        </Dialog>
        
        <Dialog open={openEditEmail} onClose={() => setOpenEditEmail(false)}>
            <DialogTitle>{t('EditEmail')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <Trans i18nKey="EditInvitedEmail">Use this form to invite edit {openEditEmail}\\'s Email.</Trans>
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email-edit"
                    label={t('EmailAddress')}
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
                }}>{t('Edit')}</Button>
            </DialogActions>
        </Dialog>

        <Dialog open={openCreateForm} onClose={() => setOpenCreateForm(false)}>
            <DialogTitle>{t('CreateUser')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('FormInviteUser')}
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-nickname"
                    label={t('Nickname')}
                    type="text"
                    fullWidth
                    variant="standard"
                />
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email"
                    label={t('EmailAddressOptional')}
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
        }}>{t('Refresh')}</Button>&nbsp;&nbsp;
        <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
            setOpenCreateForm(true)
        }}>{t('Create')}</Button><br /><br />


        {isLoading && <center><br /><CircularProgress /></center>}

        {!isLoading && rows && (<PrettyTableView 
            data={rows}
            onRowClick = {(r) => {
                setOpenEditEmail(r.nickname);
            }}
            getKey={(r) => r.nickname}
            columns={[
                {
                    title: t('User'),
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
                                label={t('Admin')}
                            /> : <Chip
                                icon={<UserOutlined />}
                                label={t('User')}
                            />) : (
                                inviteExpired ? <Chip
                                    icon={<ExclamationCircleOutlined  />}
                                    label={t('InviteExpired')}
                                    color="error"
                                /> : <Chip
                                    icon={<WarningOutlined />}
                                    label={t('InvitePending')}
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
                    title: t('CreatedAt'),
                    screenMin: 'lg',
                    field: (r) => new Date(r.createdAt).toLocaleString(),
                },
                {
                    title: t('LastLogin'),
                    screenMin: 'lg', 
                    field: (r) => {
                        const hasLastLogin = new Date(r.lastLogin).getTime() > 0;
                        return <>{hasLastLogin ? new Date(r.lastLogin).toLocaleString() : t('Never')}</>
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
                            }>{t('SendPasswordReset')}</Button>) :
                            (<Button variant="contained" className={inviteExpired ? 'shinyButton' : ''} onClick={
                                () => {
                                    setLoadingRow(r.nickname);
                                    sendlink(r.nickname, 2);
                                }
                            } color="primary">{t('ResendInvite')}</Button>)
                        }
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                setToAction(r.nickname);
                                setOpenDeleteForm(true);
                            }
                        }>{t('Delete')}</Button>
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                API.users.reset2FA(r.nickname).then(() => {
                                    refresh();
                                });
                            }
                        }>{t('Reset2FA')}</Button></>
                    }
                },
            ]}
        />)}
    </>;
};

export default UserManagement;
