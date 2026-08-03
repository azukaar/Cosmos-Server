// material-ui
import * as React from 'react';
import { Button, LinearProgress, Stack, Typography, Select, MenuItem, FormControl, InputLabel } from '@mui/material';
import { WarningOutlined, PlusCircleOutlined, CopyOutlined, ExclamationCircleOutlined , SyncOutlined, UserOutlined, KeyOutlined, CheckCircleFilled } from '@ant-design/icons';
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
import dayjs from 'dayjs';
import PermissionGuard from '../../../components/permissionGuard';
import { PERM_USERS } from '../../../utils/permissions';
import PrettyTabbedView from '../../../components/tabbedView/tabbedView';
import proFeatures from '../../../pro';

const UserManagement = () => {
    const { t } = useTranslation();
    const [isLoading, setIsLoading] = useState(false);
    const [openCreateForm, setOpenCreateForm] = React.useState(false);
    const [openDeleteForm, setOpenDeleteForm] = React.useState(false);
    const [openInviteForm, setOpenInviteForm] = React.useState(false);
    const [openEditUser, setOpenEditUser] = React.useState(null);
    const [editRole, setEditRole] = React.useState(1);
    const [createRole, setCreateRole] = React.useState(1);
    const [toAction, setToAction] = React.useState(null);
    const [loadingRow, setLoadingRow] = React.useState(null);
    const [status, setStatus] = React.useState(null);
    const [roles, setRoles] = React.useState({0: 'Guest', 1: 'User', 2: 'Admin'});

    const [rows, setRows] = useState(null);

    function refresh() {
        setIsLoading(true);
        API.users.list()
        .then(data => {
            setLoadingRow(null);
            setRows(data.data);
            setIsLoading(false);
        })

        API.getStatus().then((res) => {
          setStatus(res.data);
        });

        if (proFeatures.isPro()) {
          API.groups.list().then((groupsRes) => {
            if (groupsRes.data) {
              setRoles((prev) => {
                const updated = {...prev};
                Object.keys(groupsRes.data).forEach((k) => {
                  updated[k] = groupsRes.data[k].name;
                });
                return updated;
              });
            }
          });
        }
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

    const usersContent = <div style={{ maxWidth: "1200px", margin: "auto" }}>
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

        <Dialog open={!!openEditUser} onClose={() => setOpenEditUser(null)}>
            <DialogTitle>{t('mgmt.usermgmt.editEmailTitle')}</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    {t('mgmt.usermgmt.editEmailText', { user: openEditUser ? openEditUser.nickname : '' })}
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="c-email-edit"
                    label={t('mgmt.usermgmt.editEmail.emailInput.emailLabel')}
                    type="email"
                    fullWidth
                    variant="standard"
                    defaultValue={openEditUser ? openEditUser.email : ''}
                />
                <FormControl fullWidth variant="standard" margin="dense">
                    <InputLabel>Role</InputLabel>
                    <Select value={editRole} onChange={(e) => setEditRole(e.target.value)}>
                        {Object.entries(roles).map(([id, name]) => <MenuItem key={id} value={Number(id)}>{name}</MenuItem>)}
                    </Select>
                </FormControl>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenEditUser(null)}>{t('global.cancelAction')}</Button>
                <Button onClick={() => {
                    const payload = { role: editRole };
                    const emailVal = document.getElementById('c-email-edit').value;
                    if (emailVal) payload.email = emailVal;
                    API.users.edit(openEditUser.nickname, payload).then(() => {
                        setOpenEditUser(null);
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
                <FormControl fullWidth variant="standard" margin="dense">
                    <InputLabel>Role</InputLabel>
                    <Select value={createRole} onChange={(e) => setCreateRole(e.target.value)}>
                        {Object.entries(roles).map(([id, name]) => <MenuItem key={id} value={Number(id)}>{name}</MenuItem>)}
                    </Select>
                </FormControl>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpenCreateForm(false)}>{t('global.cancelAction')}</Button>
                <Button onClick={() => {
                    API.users.create({
                        nickname: document.getElementById('c-nickname').value,
                        email: document.getElementById('c-email').value,
                        role: createRole,
                    }).then(() => {
                        setOpenCreateForm(false);
                        refresh();
                        sendlink(document.getElementById('c-nickname').value, 2);
                    });
                }}>{t('global.createAction')}</Button>
            </DialogActions>
        </Dialog>


        <Stack direction="row" spacing={2}>
            <PermissionGuard permission={PERM_USERS}><Button variant="contained" color="primary" startIcon={<PlusCircleOutlined />} onClick={() => {
                setOpenCreateForm(true)
            }} disabled={(!proFeatures.isPro() && status && rows) ? (rows.length >= status.LicenceNumber) : false}>{t('global.createAction')}</Button></PermissionGuard>

            <Button variant="outlined" color="primary" startIcon={<SyncOutlined />} onClick={() => {
                    refresh();
            }}>{t('global.refresh')}</Button>

            {!proFeatures.isPro() && (
              <div>{t('mgmt.usermgmt.userSeatsUsed')} {rows ? rows.length : 0} / {status ? status.LicenceNumber : 0}
                  <LinearProgress style={{width: '100px'}} variant="determinate" value={(status && rows) ? (rows.length / status.LicenceNumber) * 100 : 0}
                  color={
                      (status && rows) ? (rows.length >= status.LicenceNumber ? 'error' : 'primary') : 'primary'
                  } />
              </div>
            )}
        </Stack>

        <br /><br />


        {isLoading && <center><br /><CircularProgress /></center>}

        {!isLoading && rows && (<PrettyTableView
            data={rows}
            onRowClick = {(r) => {
                setOpenEditUser(r);
                setEditRole(r.role);
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

                        return <>{isRegistered ? (<Chip
                                icon={<CheckCircleFilled  />}
                                label={t('global.active')}
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
                    title: t('global.role'),
                    screenMin: 'sm',
                    field: (r) => {
                        return roles[r.role] || r.role;
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
                    field: (r) => dayjs(new Date(r.createdAt)).format('L, LT'),
                },
                {
                    title: t('mgmt.usermgmt.lastLogin'),
                    screenMin: 'lg',
                    field: (r) => {
                        const hasLastLogin = new Date(r.lastLogin).getTime() > 0;
                        return <>{hasLastLogin ? dayjs(new Date(r.lastLogin)).format('L, LT') : t('global.never')}</>
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
                            (<PermissionGuard permission={PERM_USERS}><Button variant="contained" color="primary" onClick={
                                () => {
                                    setLoadingRow(r.nickname);
                                    sendlink(r.nickname, 1);
                                }
                            }>{t('mgmt.usermgmt.sendPasswordResetButton')}</Button></PermissionGuard>) :
                            (<PermissionGuard permission={PERM_USERS}><Button variant="contained" className={inviteExpired ? 'shinyButton' : ''} onClick={
                                () => {
                                    setLoadingRow(r.nickname);
                                    sendlink(r.nickname, 2);
                                }
                            } color="primary">{t('newInstall.usermgmt.inviteUser.resendInviteButton')}</Button></PermissionGuard>)
                        }
                        &nbsp;&nbsp;<PermissionGuard permission={PERM_USERS}><Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                setToAction(r.nickname);
                                setOpenDeleteForm(true);
                            }
                        }>{t('global.delete')}</Button></PermissionGuard>
                        &nbsp;&nbsp;<PermissionGuard permission={PERM_USERS}><Button variant="contained" color="error" onClick={
                            () => {
                                setLoadingRow(r.nickname);
                                API.users.reset2FA(r.nickname).then(() => {
                                    refresh();
                                });
                            }
                        }>{t('mgmt.usermgmt.reset2faButton')}</Button></PermissionGuard></>
                    }
                },
            ]}
        />)}
    </div>;

    const tabs = [
        {
            title: t('global.users', 'Users'),
            children: usersContent,
            url: '/users',
        },
    ];

    if (proFeatures.GroupsTab) {
        const GroupsTab = proFeatures.GroupsTab;
        tabs.push({
            title: t('mgmt.groups.title', 'Groups'),
            children: <GroupsTab />,
            url: '/groups',
        });
    }

    return <PrettyTabbedView
        rootURL="/cosmos-ui/config-users"
        tabs={tabs}
    />;
};

export default UserManagement;
