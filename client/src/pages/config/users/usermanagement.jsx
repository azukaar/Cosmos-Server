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
            <DialogTitle>{t('mgmt.usermgmt.inviteUserTitle')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    <div style={{
                        paddingBottom: '15px',
                        maxWidth: '350px',
                    }}>
                        {toAction.emailWasSent ? 
                            <div>
                                <strong>{t('mgmt.usermgmt.inviteUser.emailSentConfirmation')}</strong> {t('mgmt.usermgmt.inviteUser.emailSentwithLink')} {toAction.formAction}. {t('mgmt.usermgmt.inviteUser.emailSentAltShareLink')}
                            </div> :
                            <div>
                                {t('mgmt.usermgmt.inviteUser.emailSentAltShare')} {toAction.nickname} {t('mgmt.usermgmt.inviteUser.emailSentAltShareTo')} {toAction.formAction}:
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
                }}>{t('global.close')}</Button>
            </DialogActions>
        </Dialog>: ''}

        <Dialog open={openDeleteForm} onClose={() => setOpenDeleteForm(false)}>
            <DialogTitle>{t('mgmt.usermgmt.deleteUserTitle')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('mgmt.usermgmt.deleteUserConfirm')} {toAction} ?
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenDeleteForm(false)}>{t('global.cancelAction')}</Button>
                <Button onClick={() => {
                    API.users.deleteUser(toAction)
                    .then(() => {
                        refresh();
                        setOpenDeleteForm(false);
                    })
                }}>{t('global.delete')}</Button>
            </DialogActions>
        </Dialog>
        
        <Dialog open={openEditEmail} onClose={() => setOpenEditEmail(false)}>
            <DialogTitle>{t('mgmt.usermgmt.editEmailTitle')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('mgmt.usermgmt.editEmailText', { user: openEditEmail })}
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email-edit"
                    label={t('mgmt.usermgmt.editEmail.emailInput.emailLabel')}
                    type="email"
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenEditEmail(false)}>{t('global.cancelAction')}</Button>
                <Button onClick={() => {
                    API.users.edit(openEditEmail, {
                        email: document.getElementById('c-email-edit').value,
                    }).then(() => {
                        setOpenEditEmail(false);
                        refresh();
                    });
                }}>{t('global.edit')}</Button>
            </DialogActions>
        </Dialog>

        <Dialog open={openCreateForm} onClose={() => setOpenCreateForm(false)}>
            <DialogTitle>{t('mgmt.usermgmt.createUserTitle')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('mgmt.usermgmt.inviteUserText')}
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-nickname"
                    label={t('global.nicknameLabel')}
                    type="text"
                    fullWidth
                    variant="standard"
                />
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email"
                    label={t('mgmt.usermgmt.createUser.emailOptInput.emailOptLabel')}
                    type="email"
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenCreateForm(false)}>{t('global.cancelAction')}</Button>
                <Button onClick={() => {
                    API.users.create({
                        nickname: document.getElementById('c-nickname').value,
                        email: document.getElementById('c-email').value,
                    }).then(() => {
                        setOpenCreateForm(false);
                        refresh();
                        sendlink(document.getElementById('c-nickname').value, 2);
                    });
                }}>{t('global.createAction')}</Button>
            </DialogActions>
        </Dialog>


        <Button variant="contained" color="primary" startIcon={<SyncOutlined />} onClick={() => {
                refresh();
        }}>{t('global.refresh')}</Button>&nbsp;&nbsp;
        <Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
            setOpenCreateForm(true)
        }}>{t('global.createAction')}</Button><br /><br />


        {isLoading && <center><br /><CircularProgress /></center>}

        {!isLoading && rows && (<PrettyTableView 
            data={rows}
            onRowClick = {(r) => {
                setOpenEditEmail(r.nickname);
            }}
            getKey={(r) => r.nickname}
            columns={[
                {
                    title: t('global.user'),
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
                                label={t('mgmt.usermgmt.adminLabel')}
                            /> : <Chip
                                icon={<UserOutlined />}
                                label={t('global.user')}
                            />) : (
                                inviteExpired ? <Chip
                                    icon={<ExclamationCircleOutlined  />}
                                    label={t('mgmt.usermgmt.inviteExpiredLabel')}
                                    color="error"
                                /> : <Chip
                                    icon={<WarningOutlined />}
                                    label={t('mgmt.usermgmt.invitePendingLabel')}
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
                    title: t('global.createdAt'),
                    screenMin: 'lg',
                    field: (r) => new Date(r.createdAt).toLocaleString(),
                },
                {
                    title: t('mgmt.usermgmt.lastLogin'),
                    screenMin: 'lg', 
                    field: (r) => {
                        const hasLastLogin = new Date(r.lastLogin).getTime() > 0;
                        return <>{hasLastLogin ? new Date(r.lastLogin).toLocaleString() : t('global.never')}</>
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
                            }>{t('mgmt.usermgmt.sendPasswordResetButton')}</Button>) :
                            (<Button variant="contained" className={inviteExpired ? 'shinyButton' : ''} onClick={
                                () => {
                                    setLoadingRow(r.nickname);
                                    sendlink(r.nickname, 2);
                                }
                            } color="primary">{t('newInstall.usermgmt.inviteUser.resendInviteButton')}</Button>)
                        }
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                setToAction(r.nickname);
                                setOpenDeleteForm(true);
                            }
                        }>{t('global.delete')}</Button>
                        &nbsp;&nbsp;<Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                API.users.reset2FA(r.nickname).then(() => {
                                    refresh();
                                });
                            }
                        }>{t('mgmt.usermgmt.reset2faButton')}</Button></>
                    }
                },
            ]}
        />)}
    </>;
};

export default UserManagement;
